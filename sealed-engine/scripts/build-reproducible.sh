#!/bin/bash
# Reproducible build of the sealed engine.
#
#   build-reproducible.sh sign   --team TEAMID --bundle-id ID [--out DIR] [--allow-dirty]
#   build-reproducible.sh verify --team TEAMID --bundle-id ID [--out DIR] [--allow-dirty]
#
# Both modes build the committed source (git archive of HEAD, sealed-engine/
# only) at fixed paths under $WORK, from freshly cloned packages at the exact
# pins in Package.resolved, with the MLX kernel patch applied. Build paths end
# up in the binary, so every builder must use the same $WORK.
#
#   sign    Signs with the caller's Apple identity (-allowProvisioningUpdates).
#           Needs an unlocked login keychain. This is the build that gets published.
#   verify  No code signing at all (the linker's ad-hoc signature is suppressed
#           too). Anyone can run it; `sealed-verify rebuild` then compares the
#           result with a published signed app.
#
# Output: $OUT/SealedEngine.app and $OUT/manifest.json (default $OUT is $WORK/out-<mode>).
set -euo pipefail

EXPECTED_XCODE_BUILD="27A266a"
EXPECTED_METAL_BUILD="27A266a"
EXPECTED_XCODEGEN="2.46.0"
WORK="/private/tmp/sealed-engine-repro"
PATCH_NAME="mlx-swift-0.31.6-default-library-data.patch"

die() { echo "build-reproducible: $*" >&2; exit 1; }
say() { echo "== $*" >&2; }

# System tools first; Homebrew only supplies xcodegen.
export PATH="/usr/bin:/bin:/usr/sbin:/sbin:/opt/homebrew/bin:/usr/local/bin"
export DEVELOPER_DIR="${DEVELOPER_DIR:-/Applications/Xcode.app/Contents/Developer}"

MODE="${1:-}"; [ $# -gt 0 ] && shift
[ "$MODE" = sign ] || [ "$MODE" = verify ] || die "usage: $0 sign|verify --team TEAMID --bundle-id ID [--out DIR] [--allow-dirty]"
TEAM=""; BUNDLE_ID=""; OUT=""; ALLOW_DIRTY=0
while [ $# -gt 0 ]; do
  case "$1" in
    --team|--bundle-id|--out) [ $# -ge 2 ] || die "$1 needs a value" ;;
  esac
  case "$1" in
    --team) TEAM="$2"; shift 2 ;;
    --bundle-id) BUNDLE_ID="$2"; shift 2 ;;
    --out) OUT="$2"; shift 2 ;;
    --allow-dirty) ALLOW_DIRTY=1; shift ;;
    *) die "unknown argument: $1" ;;
  esac
done
[[ "$TEAM" =~ ^[A-Z0-9]{10}$ ]] || die "--team must be a 10-character Apple Team ID"
[[ "$BUNDLE_ID" =~ ^[A-Za-z0-9][A-Za-z0-9.-]*$ ]] || die "--bundle-id is not a valid bundle identifier"
[ -n "$OUT" ] || OUT="$WORK/out-$MODE"

# ---- Toolchain checks -------------------------------------------------------
XCODE_VERSION_OUT="$(xcodebuild -version)"
XCODE_BUILD="$(awk '/^Build version/ {print $3}' <<<"$XCODE_VERSION_OUT")"
[ "$XCODE_BUILD" = "$EXPECTED_XCODE_BUILD" ] || die "Xcode build is '$XCODE_BUILD', need $EXPECTED_XCODE_BUILD"
METAL_INFO="$(xcodebuild -showComponent MetalToolchain 2>&1)"
METAL_BUILD="$(awk -F': ' '/^Build Version/ {print $2}' <<<"$METAL_INFO")"
METAL_STATUS="$(awk -F': ' '/^Status/ {print $2}' <<<"$METAL_INFO")"
METAL_ID="$(awk -F': ' '/^Toolchain Identifier/ {print $2}' <<<"$METAL_INFO")"
[ "$METAL_STATUS" = installed ] || die "Metal toolchain not installed (xcodebuild -downloadComponent MetalToolchain)"
[ "$METAL_BUILD" = "$EXPECTED_METAL_BUILD" ] || die "Metal toolchain build is '$METAL_BUILD', need $EXPECTED_METAL_BUILD"
[ -n "$METAL_ID" ] || die "could not read the Metal toolchain identifier"
XCODEGEN="$(command -v xcodegen)" || die "xcodegen not on PATH"
XCODEGEN_VERSION_OUT="$("$XCODEGEN" --version)"
XCODEGEN_VERSION="$(awk '{print $2}' <<<"$XCODEGEN_VERSION_OUT")"
[ "$XCODEGEN_VERSION" = "$EXPECTED_XCODEGEN" ] || die "xcodegen is '$XCODEGEN_VERSION', need $EXPECTED_XCODEGEN"
HOST_OS_BUILD="$(sw_vers -buildVersion)"

# ---- Work directory: ours alone, one build at a time ---------------------------
mkdir -p "$WORK"
[ -O "$WORK" ] && [ ! -L "$WORK" ] || die "$WORK exists and is not owned by this user"
chmod 700 "$WORK"
mkdir "$WORK/.lock" 2>/dev/null || die "another build holds $WORK/.lock (remove it if no build is running)"
trap 'rmdir "$WORK/.lock"' EXIT

if [ -e "$OUT" ]; then
  # Only a previous run's output may be replaced.
  [ -d "$OUT" ] && [ ! -L "$OUT" ] || die "--out $OUT is not a directory"
  for f in "$OUT"/* "$OUT"/.[!.]*; do
    [ -e "$f" ] || [ -L "$f" ] || continue
    case "${f##*/}" in
      SealedEngine.app|manifest.json) ;;
      *) die "--out $OUT holds other files; refusing to delete it" ;;
    esac
  done
  rm -rf "$OUT"
fi

# ---- Committed source only ---------------------------------------------------
REPO="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"
COMMIT="$(git -C "$REPO" rev-parse HEAD)"
# The engine is built from sealed-engine/ alone, so its tree hash identifies the source.
TREE="$(git -C "$REPO" rev-parse "$COMMIT:sealed-engine")"
STATUS="$(git -C "$REPO" status --porcelain -- sealed-engine)"
DIRTY=false
if [ -n "$STATUS" ]; then
  DIRTY=true
  [ "$ALLOW_DIRTY" = 1 ] || die "sealed-engine/ has uncommitted changes; commit them or pass --allow-dirty (the build still uses HEAD)"
  say "WARNING: working tree is dirty; building HEAD $COMMIT, not the working tree"
fi
SRC="$WORK/src"
rm -rf "$SRC" "$WORK/dd" "$WORK/pkgs"
mkdir -p "$SRC"
git -C "$REPO" archive "$COMMIT" sealed-engine | tar -x -C "$SRC"
ENGINE="$SRC/sealed-engine"
RESOLVED="$ENGINE/Package.resolved"
PATCH="$ENGINE/patches/$PATCH_NAME"
[ -f "$RESOLVED" ] || die "Package.resolved is not committed"
[ -f "$PATCH" ] || die "MLX patch is not committed"
RESOLVED_SHA="$(shasum -a 256 "$RESOLVED" | awk '{print $1}')"
PATCH_SHA="$(shasum -a 256 "$PATCH" | awk '{print $1}')"

# ---- Identity and project ----------------------------------------------------
printf 'DEVELOPMENT_TEAM = %s\nPRODUCT_BUNDLE_IDENTIFIER = %s\n' "$TEAM" "$BUNDLE_ID" > "$ENGINE/Config/Local.xcconfig"
(cd "$ENGINE" && "$XCODEGEN" generate -q)
PROJ="$ENGINE/SealedEngine.xcodeproj"
SWIFTPM_DIR="$PROJ/project.xcworkspace/xcshareddata/swiftpm"
mkdir -p "$SWIFTPM_DIR"
cp "$RESOLVED" "$SWIFTPM_DIR/Package.resolved"

COMMON=(-project "$PROJ" -scheme SealedEngine -configuration Release
        -derivedDataPath "$WORK/dd" -clonedSourcePackagesDirPath "$WORK/pkgs"
        -onlyUsePackageVersionsFromResolvedFile
        -skipPackagePluginValidation -skipMacroValidation)

# ---- Fresh package clones at the pins; every checkout must be pristine -------
say "resolving packages"
xcodebuild -resolvePackageDependencies "${COMMON[@]}" >"$WORK/resolve.log" 2>&1 || { tail -20 "$WORK/resolve.log" >&2; die "package resolution failed"; }
cmp -s "$RESOLVED" "$SWIFTPM_DIR/Package.resolved" || die "resolution changed Package.resolved"
# Each pin's checkout: HEAD at the pinned revision, no modified, untracked or
# ignored files, and every submodule at its recorded commit and equally clean.
/usr/bin/python3 - "$RESOLVED" "$WORK/pkgs/checkouts" <<'PY' || die "a package checkout does not match its pin"
import json, os, subprocess, sys
pins = json.load(open(sys.argv[1]))["pins"]
root = sys.argv[2]
def git(d, *a):
    return subprocess.run(["git", "-C", d, *a], check=True, capture_output=True, text=True).stdout
bad = 0
expected_dirs = set()
for p in pins:
    name = p["location"].rstrip("/").rsplit("/", 1)[-1]
    name = name[:-4] if name.endswith(".git") else name
    expected_dirs.add(name)
    d = os.path.join(root, name)
    problems = []
    try:
        if git(d, "rev-parse", "HEAD").strip() != p["state"]["revision"]:
            problems.append("HEAD is not the pinned revision")
        if git(d, "status", "--porcelain", "--ignored", "--ignore-submodules=none").strip():
            problems.append("checkout is not pristine")
        for line in git(d, "submodule", "status", "--recursive").splitlines():
            if not line.startswith(" "):
                problems.append("submodule not at its recorded commit: " + line.strip())
        subs = git(d, "submodule", "foreach", "--quiet", "--recursive",
                   "git status --porcelain --ignored --ignore-submodules=none")
        if subs.strip():
            problems.append("a submodule is not pristine")
    except subprocess.CalledProcessError as e:
        problems.append("git failed: " + (e.stderr or "").strip())
    bad += bool(problems)
    print(("ok  " if not problems else "BAD ") + p["identity"] + " " + "; ".join(problems), file=sys.stderr)
extra = set(os.listdir(root)) - expected_dirs
if extra:
    print("BAD unexpected checkouts: " + ", ".join(sorted(extra)), file=sys.stderr)
    bad += 1
sys.exit(1 if bad else 0)
PY

# ---- Apply the MLX kernel patch exactly (git apply refuses fuzz) --------------
MLX="$WORK/pkgs/checkouts/mlx-swift"
git -C "$MLX" apply --check "$PATCH" || die "MLX patch does not apply cleanly to the pinned checkout"
git -C "$MLX" apply "$PATCH" || die "MLX patch failed to apply"
git -C "$MLX" apply --reverse --check "$PATCH" || die "MLX patch is not fully applied"

# ---- Build -------------------------------------------------------------------
say "building ($MODE)"
if [ "$MODE" = sign ]; then
  EXTRA=(-allowProvisioningUpdates)
else
  EXTRA=(CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO "CODE_SIGN_IDENTITY=" "SEALED_EXTRA_LDFLAGS=-Wl,-no_adhoc_codesign")
fi
xcodebuild build "${COMMON[@]}" -disableAutomaticPackageResolution "${EXTRA[@]}" >"$WORK/build.log" 2>&1 || { tail -30 "$WORK/build.log" >&2; die "build failed (log: $WORK/build.log)"; }

APP="$WORK/dd/Build/Products/Release/SealedEngine.app"
BIN="$APP/Contents/MacOS/SealedEngine"
[ -x "$BIN" ] || die "no executable at $BIN"
if grep -rl --include=Info.plist BuildMachineOSBuild "$APP" >/dev/null 2>&1; then
  die "an Info.plist still names the build host's macOS"
fi

# ---- Embedded kernels ----------------------------------------------------------
KDIR="$(mktemp -d "$WORK/kernel.XXXXXX")"
xcrun segedit "$BIN" -extract __TEXT __mlxlib "$KDIR/mlxlib" || die "no __TEXT,__mlxlib section in the executable"
KERNEL_SHA="$(shasum -a 256 "$KDIR/mlxlib" | awk '{print $1}')"
KERNEL_BYTES="$(stat -f %z "$KDIR/mlxlib")"
BUILT_METALLIB="$WORK/dd/Build/Products/Release/mlx-swift_Cmlx.bundle/Contents/Resources/default.metallib"
METALLIB_SHA="$(shasum -a 256 "$BUILT_METALLIB" | awk '{print $1}')"
[ "$METALLIB_SHA" = "$KERNEL_SHA" ] || die "embedded kernels differ from the built default.metallib"
rm -rf "$KDIR"

CDHASH=""
if [ "$MODE" = sign ]; then
  codesign --verify --strict "$APP" || die "signed app does not verify"
  CS_INFO="$(codesign -dvvvv "$APP" 2>&1)"
  CDHASH="$(awk -F'sha256=' '/CandidateCDHashFull sha256=/ {print $2; exit}' <<<"$CS_INFO")"
  [[ "$CDHASH" =~ ^[0-9a-f]{64}$ ]] || die "could not read the SHA-256 CD hash"
else
  if codesign -dv "$BIN" >/dev/null 2>&1; then die "verify build carries a code signature; -no_adhoc_codesign did not take effect"; fi
fi

# ---- Output --------------------------------------------------------------------
mkdir -p "$OUT"
ditto "$APP" "$OUT/SealedEngine.app"
/usr/bin/python3 - "$OUT/manifest.json" <<PY
import json, sys
m = {
  "schema": "sealed-engine-build/v1",
  "mode": "$MODE",
  "bundle_id": "$BUNDLE_ID",
  "team": "$TEAM",
  "cdhash_sha256": "$CDHASH" or None,
  "source_commit": "$COMMIT",
  "source_tree": "$TREE",
  "dirty": $( [ "$DIRTY" = true ] && echo True || echo False ),
  "xcode_build": "$XCODE_BUILD",
  "metal_toolchain": {"build": "$METAL_BUILD", "identifier": "$METAL_ID"},
  "xcodegen": "$XCODEGEN_VERSION",
  "host_macos_build": "$HOST_OS_BUILD",
  "work_dir": "$WORK",
  "package_resolved_sha256": "$RESOLVED_SHA",
  "mlx_patch_sha256": "$PATCH_SHA",
  "embedded_kernels": {"section": "__TEXT,__mlxlib", "bytes": $KERNEL_BYTES, "sha256": "$KERNEL_SHA"},
}
json.dump(m, open(sys.argv[1], "w"), indent=2)
open(sys.argv[1], "a").write("\n")
PY
say "done: $OUT"
cat "$OUT/manifest.json"
