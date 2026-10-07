#!/bin/bash
# Builds and runs the spawn-module arm suite on the test Mac, never on this machine.
# Usage: ATTEST_HOST=<ssh host> [SLOT=name] [MODULE=path/to/sealed_spawn.c] ./run-on-test-mac.sh [tests|tests/probes]
# MODULE lets a mutant stand in for the module (mutation check); default is the real one.
set -u
cd "$(dirname "$0")" || exit 2
H=${ATTEST_HOST:?set ATTEST_HOST to the test Mac ssh host}
sub=${1:-tests}; slot=${SLOT:-$(echo "$sub" | tr '/' '-')}; R=sealed-spawn-glm/$slot
G=../../Sources/SealedGuard; M=${MODULE:-$G/sealed_spawn.c}
st=$(mktemp -d); trap 'rm -rf "$st"' EXIT
mkdir -p "$st/impl/src"
cp -R "$G/include" "$st/include"; cp "$G/sealed_guard.c" "$st/impl/src/"; cp "$M" "$st/impl/src/sealed_spawn.c"
cp -R tests probes "$st/"
ssh -o ConnectTimeout=10 -o BatchMode=yes "$H" "mkdir -p ~/$R" || { echo "test Mac unreachable"; exit 3; }
rsync -a --delete --exclude build/ "$st/" "$H:$R/" || exit 3
ssh -o BatchMode=yes "$H" "cd ~/$R/$sub && perl -e 'alarm 600; exec @ARGV' make test 2>&1; echo \"make test exit: \$?\""
