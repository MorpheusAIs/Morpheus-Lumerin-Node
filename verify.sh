#!/usr/bin/env bash
#
# One-shot repository verification.
#
#   ./verify.sh
#
# Runs the Go compiler and test suite, the desktop unit tests and TypeScript
# typecheck, and a full Electron build. It runs every check it can and reports
# all failures at the end.
#
# It may create ui-desktop/.env from the checked-in example, install
# dependencies, and produce ignored build output.

set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$REPO_ROOT"

BOLD=$'\033[1m'; RED=$'\033[31m'; GREEN=$'\033[32m'; YELLOW=$'\033[33m'; DIM=$'\033[2m'; RESET=$'\033[0m'

FAILED=()
STEP=0

step() {
  STEP=$((STEP + 1))
  printf '\n%s[%d/9] %s%s\n' "$BOLD" "$STEP" "$1" "$RESET"
}

# Keep result accounting in this shell. Calling a helper from `( cd ... )`
# loses updates to FAILED when that subshell exits and can falsely report that
# all checks passed.
run_in() {
  local directory="$1"; shift
  local label="$1"; shift
  if ( cd "$directory" && "$@" ); then
    printf '%s  ✓ %s%s\n' "$GREEN" "$label" "$RESET"
    return 0
  else
    printf '%s  ✗ %s%s\n' "$RED" "$label" "$RESET"
    FAILED+=("$label")
    return 1
  fi
}

require() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf '%s  ✗ %s is not installed or not on PATH%s\n' "$RED" "$1" "$RESET"
    FAILED+=("$1 missing")
    return 1
  fi
  printf '%s  %s %s%s\n' "$DIM" "$1" "$($2 2>&1 | head -1)" "$RESET"
}

printf '%sVerifying repository%s\n' "$BOLD" "$RESET"
printf '%son %s%s\n' "$DIM" "$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown branch')" "$RESET"

step "Checking toolchain"
require go "go version"
require node "node --version"
if command -v yarn >/dev/null 2>&1; then
  require yarn "yarn --version"
else
  printf '%s  ✗ yarn is not installed (npm i -g yarn)%s\n' "$RED" "$RESET"
  FAILED+=("yarn missing")
fi

if [ ${#FAILED[@]} -gt 0 ]; then
  printf '\n%sInstall the missing tools above, then re-run.%s\n' "$RED" "$RESET"
  exit 1
fi

# ---------------------------------------------------------------- proxy-router
step "Building proxy-router"
run_in proxy-router "go build ./..." go build ./...

step "Vetting proxy-router"
run_in proxy-router "go vet ./..." go vet ./...

step "Running proxy-router tests"
run_in proxy-router "go test ./..." go test ./...

# ------------------------------------------------------------------ ui-desktop
step "Checking ui-desktop/.env"
if [ -f ui-desktop/.env ]; then
  printf '%s  ✓ .env present%s\n' "$GREEN" "$RESET"
else
  printf '%s  .env missing — creating it from .env.example%s\n' "$YELLOW" "$RESET"
  if cp ui-desktop/.env.example ui-desktop/.env; then
    printf '%s  ✓ created ui-desktop/.env (Base Mainnet defaults)%s\n' "$GREEN" "$RESET"
  else
    printf '%s  ✗ could not create ui-desktop/.env%s\n' "$RED" "$RESET"
    FAILED+=(".env")
  fi
fi

step "Installing ui-desktop dependencies"
run_in ui-desktop "yarn install" yarn install --network-timeout 600000

step "Running ui-desktop unit tests"
run_in ui-desktop "yarn test" yarn test

step "Typechecking ui-desktop"
run_in ui-desktop "yarn typecheck" yarn typecheck

step "Building the desktop app"
case "$(uname -s)" in
  Darwin)
    # Avoid colliding with a mounted release DMG and force the packaged app to
    # carry the proxy-router produced from this checkout.
    VERIFY_DMG_TITLE='MorpheusUI Verify ${version}-${arch}-'"$$"
    run_in ui-desktop "electron-vite build" yarn electron-vite build
    run_in ui-desktop "unsigned macOS package with current proxy-router" env \
      FORCE_BUNDLED_PROXY_ROUTER=1 \
      yarn electron-builder \
      --config electron.builder.config.ts \
      --mac \
      "--config.dmg.title=$VERIFY_DMG_TITLE"
    ;;
  Linux)
    run_in ui-desktop "yarn build:linux" yarn build:linux
    ;;
  *)
    run_in ui-desktop "yarn build" yarn build
    ;;
esac

# ---------------------------------------------------------------------- report
printf '\n%s────────────────────────────────────────%s\n' "$BOLD" "$RESET"
if [ ${#FAILED[@]} -eq 0 ]; then
  printf '%s%sAll checks passed.%s\n\n' "$BOLD" "$GREEN" "$RESET"
  cat <<'NEXT'
Run the app:

    cd ui-desktop && yarn dev

  First launch prepares the proxy-router and downloads optional legacy demo
  assets into your app-data directory, so give it a few minutes and watch the
  startup progress. Subsequent launches are faster.

Now smoke-test the main workflows by hand:

  1. Unresponsive UI
     Click rapidly between Wallet / Chat / Models / Providers / Agents for
     ~15 seconds. Buttons should stay responsive throughout.

  2. Double-staking
     Open a session, then switch to Wallet BEFORE it finishes. Come back to
     Chat. The session should be there, not the "Select payment method" screen.

  3. Slow tabs
     Open Providers and Models with a populated account. They should render
     in seconds, not minutes.

  4. Token transfer
     Wallet tab -> "Send" tile. Try an invalid address and an over-balance
     amount first; both should be rejected before any gas is spent.

  5. Connection errors
     Quit the proxy-router (Settings -> stop) and reload the Wallet tab. You
     should see "Not connected to your node", NOT a balance of 0.

  6. Session-gated Chat and Workspace
     Choose a marketplace model, select a duration, and explicitly open a
     session with MOR stake or one-off Direct Pay. Verify that exact active
     session works in both normal Chat and Workspace, while a closed or expired
     session does not. Workspace and schedules must never open, fund, renew,
     extend, or substitute a session automatically.

  7. Desktop hardening
     Confirm DevTools stays closed by default. Cancel wallet reset, session,
     provider-claim, and agent-permission confirmations once each and verify
     that no underlying action occurs.
NEXT
  exit 0
else
  printf '%s%s%d check(s) failed:%s\n' "$BOLD" "$RED" "${#FAILED[@]}" "$RESET"
  for f in "${FAILED[@]}"; do printf '  %s- %s%s\n' "$RED" "$f" "$RESET"; done
  printf '\n%sScroll up for the actual compiler/test output.%s\n' "$DIM" "$RESET"
  exit 1
fi
