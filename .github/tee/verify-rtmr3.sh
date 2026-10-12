#!/usr/bin/env bash
# Post-deploy RTMR3 verification for a SecretVM instance.
#
# Polls `secretvm-cli vm attestation` until the live TDX quote's RTMR3 matches
# the value the deployed image should produce, under whichever RTMR3 scheme
# the VM uses (see proxy-router/scripts/compute-rtmr3.py):
#
#   attest-tool  pre-dstack images: expected = EXPECTED_RTMR3 (published in
#                the signed manifest at build time).
#   dstack       images with a dstack KMS (portal v0.0.34+): the VM folds its
#                own app-id into RTMR3, so the expected value is recomputed
#                here from COMPOSE_FILE + ROOTFS_SHA256 + the app-id read from
#                https://<vmDomain>:29343/info.
#
# Required env:
#   SECRETVM_API_KEY  SECRETVM_VM_UUID  EXPECTED_RTMR3  COMPOSE_FILE  ROOTFS_SHA256
# Optional env:
#   RTMR3_SCRIPT   (default proxy-router/scripts/compute-rtmr3.py)
#   ATTEST_PORT    (default 29343)
#   MAX_ATTEMPTS   (default 12)   SLEEP_SECS (default 30)
#   GITHUB_STEP_SUMMARY is appended to when set.
set -uo pipefail

: "${SECRETVM_API_KEY:?}" "${SECRETVM_VM_UUID:?}" "${EXPECTED_RTMR3:?}" "${COMPOSE_FILE:?}" "${ROOTFS_SHA256:?}"
RTMR3_SCRIPT="${RTMR3_SCRIPT:-proxy-router/scripts/compute-rtmr3.py}"
ATTEST_PORT="${ATTEST_PORT:-29343}"
MAX_ATTEMPTS="${MAX_ATTEMPTS:-12}"
SLEEP_SECS="${SLEEP_SECS:-30}"
SUMMARY="${GITHUB_STEP_SUMMARY:-/dev/null}"

# TDX Quote v4: 48-byte header + TD Report Body; RTMR3 at body offset 472 (48 bytes).
RTMR3_HEX_START=$(( (48 + 472) * 2 ))
RTMR3_HEX_LEN=96

extract_rtmr3() {
  python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    raw = data.get('result', '')
    rtmr3 = raw[$RTMR3_HEX_START:$RTMR3_HEX_START+$RTMR3_HEX_LEN]
    print(rtmr3 if len(rtmr3) == $RTMR3_HEX_LEN else '')
except Exception:
    print('')
" 2>/dev/null
}

# The VM's public domain, for /info. Empty when the portal has not assigned one.
VM_DOMAIN=$(secretvm-cli -k "$SECRETVM_API_KEY" vm status "$SECRETVM_VM_UUID" 2>/dev/null \
  | python3 -c "
import sys, json
try:
    r = json.load(sys.stdin).get('result') or {}
    print((r.get('vmDomain') or r.get('custom_domain') or '').strip())
except Exception:
    print('')
" 2>/dev/null)

fetch_app_id() {
  [ -n "$VM_DOMAIN" ] || { echo ""; return; }
  curl -fsSk --max-time 10 "https://${VM_DOMAIN}:${ATTEST_PORT}/info" 2>/dev/null \
    | python3 -c "
import sys, json
try:
    v = (json.load(sys.stdin).get('dstack_app_id') or '').strip().lower()
    v = v[2:] if v.startswith('0x') else v
    bytes.fromhex(v)
    print(v)
except Exception:
    print('')
" 2>/dev/null
}

echo "⏳ Waiting for VM to reboot with new image..." >> "$SUMMARY"
[ -n "$VM_DOMAIN" ] && echo "   VM domain: \`$VM_DOMAIN\`" >> "$SUMMARY"

MATCHED=false
LIVE_RTMR3=""
APP_ID=""
SCHEME=""
EXPECTED_DSTACK=""

for i in $(seq 1 "$MAX_ATTEMPTS"); do
  echo "  Attestation poll attempt $i/$MAX_ATTEMPTS..."

  RAW_QUOTE=$(secretvm-cli -k "$SECRETVM_API_KEY" vm attestation "$SECRETVM_VM_UUID" 2>/dev/null) || true
  LIVE_RTMR3=$(echo "$RAW_QUOTE" | extract_rtmr3)

  if [ -z "$LIVE_RTMR3" ]; then
    echo "  No attestation data yet (VM may still be starting)"
  else
    echo "  Got RTMR3: ${LIVE_RTMR3:0:16}..."

    # Re-read the app-id each attempt: /info only answers once the VM is up,
    # and a redeploy can rotate it.
    APP_ID=$(fetch_app_id)
    if [ -n "$APP_ID" ]; then
      SCHEME="dstack"
      EXPECTED_DSTACK=$(python3 "$RTMR3_SCRIPT" "$COMPOSE_FILE" \
        --rootfs-sha256="$ROOTFS_SHA256" --dstack-app-id="$APP_ID")
    else
      SCHEME="attest-tool"
      EXPECTED_DSTACK=""
    fi

    if [ "$LIVE_RTMR3" = "$EXPECTED_RTMR3" ]; then
      MATCHED=true
      SCHEME="attest-tool"
    elif [ -n "$EXPECTED_DSTACK" ] && [ "$LIVE_RTMR3" = "$EXPECTED_DSTACK" ]; then
      MATCHED=true
    fi

    if [ "$MATCHED" = true ]; then
      {
        echo "🎉 **RTMR3 MATCH** — attestation verified (scheme: \`$SCHEME\`)"
        echo "   Live RTMR3:     \`$LIVE_RTMR3\`"
        if [ "$SCHEME" = "dstack" ]; then
          echo "   dstack app-id:  \`$APP_ID\`"
          echo "   Expected RTMR3: \`$EXPECTED_DSTACK\` (replayed from compose + rootfs + app-id)"
          echo "   Manifest RTMR3: \`$EXPECTED_RTMR3\` (attest-tool scheme; not applicable to this VM)"
        else
          echo "   Expected RTMR3: \`$EXPECTED_RTMR3\`"
        fi
      } >> "$SUMMARY"
      break
    fi

    echo "  RTMR3 mismatch (VM may still be rebooting)"
    echo "    live:                 $LIVE_RTMR3"
    echo "    expected attest-tool: $EXPECTED_RTMR3"
    [ -n "$EXPECTED_DSTACK" ] && echo "    expected dstack:      $EXPECTED_DSTACK (app_id=$APP_ID)"
  fi

  if [ "$i" -lt "$MAX_ATTEMPTS" ]; then
    sleep "$SLEEP_SECS"
  fi
done

if [ "$MATCHED" = false ]; then
  {
    echo "⚠️ **RTMR3 mismatch after ${MAX_ATTEMPTS} attempts** (${SLEEP_SECS}s × ${MAX_ATTEMPTS} = $(( SLEEP_SECS * MAX_ATTEMPTS ))s)"
    echo "   Live RTMR3:              \`${LIVE_RTMR3:-unavailable}\`"
    echo "   Expected (attest-tool):  \`$EXPECTED_RTMR3\`"
    if [ -n "$EXPECTED_DSTACK" ]; then
      echo "   Expected (dstack):       \`$EXPECTED_DSTACK\` (app_id \`$APP_ID\`)"
    elif [ -z "$VM_DOMAIN" ]; then
      echo "   dstack check skipped: VM has no domain, so /info could not be read."
    else
      echo "   dstack check skipped: https://${VM_DOMAIN}:${ATTEST_PORT}/info reported no dstack_app_id."
    fi
    echo "   This may indicate the VM is still rebooting, or there is a genuine mismatch."
    echo "   Check manually: \`secretvm-cli vm attestation $SECRETVM_VM_UUID\`"
  } >> "$SUMMARY"
  exit 1
fi
