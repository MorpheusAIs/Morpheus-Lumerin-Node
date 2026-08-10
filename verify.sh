#!/usr/bin/env bash
#
# One-shot verification for the fix/ux-stability branch.
#
#   ./verify.sh
#
# Runs everything that could not be run in the environment where these changes
# were written: the Go compiler, the Go test suite, a TypeScript typecheck, and
# a full Electron build. Stops at the first failure and tells you what broke.
#
# Nothing here modifies your working tree apart from installing dependencies
# and producing build output.

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

run() {
  local label="$1"; shift
  if "$@"; then
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

printf '%sVerifying fix/ux-stability%s\n' "$BOLD" "$RESET"
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
# This is the highest-risk part of the branch: it was never compiled while the
# changes were being written.
step "Building proxy-router (never compiled during development)"
( cd proxy-router && run "go build ./..." go build ./... )

step "Vetting proxy-router"
( cd proxy-router && run "go vet ./..." go vet ./... )

step "Running proxy-router tests"
printf '%s  Note: these 61 test files have not run in CI for a long time.%s\n' "$YELLOW" "$RESET"
printf '%s  If something fails, check whether it also fails on main before%s\n' "$YELLOW" "$RESET"
printf '%s  assuming it is a new regression: git stash && git checkout main%s\n' "$YELLOW" "$RESET"
( cd proxy-router && run "go test ./..." go test ./... )

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
printf '%s  This also regenerates yarn.lock with the new test dependencies.%s\n' "$DIM" "$RESET"
( cd ui-desktop && run "yarn install" yarn install --network-timeout 600000 )

step "Running ui-desktop unit tests"
( cd ui-desktop && run "yarn test (expect 65 passing)" yarn test )

step "Typechecking ui-desktop"
( cd ui-desktop && run "yarn typecheck" yarn typecheck )

step "Building the desktop app"
case "$(uname -s)" in
  Darwin) BUILD_TARGET="build:mac" ;;
  Linux)  BUILD_TARGET="build:linux" ;;
  *)      BUILD_TARGET="build" ;;
esac
( cd ui-desktop && run "yarn $BUILD_TARGET" yarn "$BUILD_TARGET" )

# ---------------------------------------------------------------------- report
printf '\n%s────────────────────────────────────────%s\n' "$BOLD" "$RESET"
if [ ${#FAILED[@]} -eq 0 ]; then
  printf '%s%sAll checks passed.%s\n\n' "$BOLD" "$GREEN" "$RESET"
  cat <<'NEXT'
Run the app:

    cd ui-desktop && yarn dev

  First launch downloads the proxy-router, a llama.cpp + tinyllama demo model
  and an IPFS node into your app-data directory, so give it a few minutes and
  watch the startup progress. Subsequent launches are fast.

Now smoke-test the three reported symptoms by hand:

  1. Unresponsive UI
     Click rapidly between Wallet / Chat / Models / Providers / Agents for
     ~15 seconds. Buttons should stay responsive throughout.

  2. Double-staking
     Open a session, then switch to Wallet BEFORE it finishes. Come back to
     Chat. The session should be there, not the "Select payment method" screen.

  3. Slow tabs
     Open Providers and Models with a populated account. They should render
     in seconds, not minutes.

  Plus the new feature:
     Wallet tab -> "Send" tile. Try an invalid address and an over-balance
     amount first; both should be rejected before any gas is spent.

  And the error handling:
     Quit the proxy-router (Settings -> stop) and reload the Wallet tab. You
     should see "Not connected to your node", NOT a balance of 0.

If everything holds up:  git push -u origin fix/ux-stability
NEXT
  exit 0
else
  printf '%s%s%d check(s) failed:%s\n' "$BOLD" "$RED" "${#FAILED[@]}" "$RESET"
  for f in "${FAILED[@]}"; do printf '  %s- %s%s\n' "$RED" "$f" "$RESET"; done
  printf '\n%sScroll up for the actual compiler/test output.%s\n' "$DIM" "$RESET"
  exit 1
fi
