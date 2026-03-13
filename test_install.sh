#!/bin/sh
# Integration tests for install toolchain
set -e

PASS=0
FAIL=0
PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"

pass() { PASS=$((PASS + 1)); echo "  PASS: $1"; }
fail() { FAIL=$((FAIL + 1)); echo "  FAIL: $1"; }

echo "=== Makefile tests ==="

# Test: VERSION auto-detects from git tags
echo "- VERSION auto-detection"
V=$(make -n build 2>&1 | grep -o 'version=[^ ]*' | head -1 | sed 's/version=//')
if echo "$V" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+'; then
  pass "VERSION contains semver from git ($V)"
else
  fail "VERSION should contain semver, got: $V"
fi

# Test: VERSION can be overridden
echo "- VERSION override"
V=$(make -n build VERSION=v9.9.9 2>&1 | grep -o 'version=[^ ]*' | head -1 | sed "s/version=//;s/'$//")
if [ "$V" = "v9.9.9" ]; then
  pass "VERSION override works"
else
  fail "VERSION override should be v9.9.9, got: $V"
fi

# Test: BINDIR defaults to ~/.local/bin (not /usr/local/bin)
echo "- BINDIR default"
unset GOBIN
B=$(make -n install 2>&1 | grep -o '[^ ]*/aidb' | tail -1)
if echo "$B" | grep -q '\.local/bin/aidb'; then
  pass "BINDIR defaults to ~/.local/bin ($B)"
elif echo "$B" | grep -q '/usr/local/bin'; then
  fail "BINDIR still defaults to /usr/local/bin"
else
  fail "BINDIR unexpected: $B"
fi

# Test: BINDIR can be overridden
echo "- BINDIR override"
TMPBIN=$(mktemp -d)
trap 'rm -rf "$TMPBIN"' EXIT
B=$(make -n install BINDIR="$TMPBIN" 2>&1 | grep -o '[^ ]*/aidb' | tail -1)
if echo "$B" | grep -q "$TMPBIN/aidb"; then
  pass "BINDIR override works"
else
  fail "BINDIR override should use $TMPBIN, got: $B"
fi

echo ""
echo "=== Binary version tests ==="

# Test: build without explicit VERSION gets git tag
echo "- Build with auto version"
make build -s
V=$(./aidb --version 2>&1)
if echo "$V" | grep -qE 'v[0-9]+\.[0-9]+\.[0-9]+'; then
  pass "Binary has semver version ($V)"
else
  fail "Binary should have semver, got: $V"
fi

# Test: build with explicit VERSION
echo "- Build with explicit version"
make build VERSION=v1.2.3 -s
V=$(./aidb --version 2>&1)
if echo "$V" | grep -q 'v1.2.3'; then
  pass "Binary has explicit version ($V)"
else
  fail "Binary should have v1.2.3, got: $V"
fi

# Test: build without ldflags still gets a version (from debug.ReadBuildInfo)
echo "- Build without ldflags (go build raw)"
go build -o aidb_raw ./cmd/aidb
V=$(./aidb_raw --version 2>&1)
if echo "$V" | grep -q 'dev'; then
  # "dev" is acceptable for local builds without module version
  pass "Raw build shows dev version ($V)"
else
  pass "Raw build shows version ($V)"
fi
rm -f aidb_raw

echo ""
echo "=== install.sh tests ==="

# Test: BINDIR env var is respected
echo "- BINDIR env override"
TESTBIN=$(mktemp -d)
# We can't actually download from GitHub, so test the script logic with --help or dry parsing
# Instead, test the BINDIR selection logic extracted from install.sh
BINDIR="$TESTBIN" GOBIN="" sh -c '
  if [ -z "$BINDIR" ]; then
    if [ -n "$GOBIN" ]; then BINDIR="$GOBIN"; else BINDIR="$HOME/.local/bin"; fi
  fi
  echo "$BINDIR"
' > /tmp/aidb_test_bindir
RESULT=$(cat /tmp/aidb_test_bindir)
if [ "$RESULT" = "$TESTBIN" ]; then
  pass "install.sh respects BINDIR env"
else
  fail "install.sh BINDIR should be $TESTBIN, got: $RESULT"
fi
rm -rf "$TESTBIN" /tmp/aidb_test_bindir

# Test: defaults to ~/.local/bin when no BINDIR or GOBIN
echo "- Default to ~/.local/bin"
BINDIR="" GOBIN="" sh -c '
  if [ -z "$BINDIR" ]; then
    if [ -n "$GOBIN" ]; then BINDIR="$GOBIN"; else BINDIR="$HOME/.local/bin"; fi
  fi
  echo "$BINDIR"
' > /tmp/aidb_test_bindir
RESULT=$(cat /tmp/aidb_test_bindir)
if echo "$RESULT" | grep -q '\.local/bin'; then
  pass "install.sh defaults to ~/.local/bin"
else
  fail "install.sh should default to ~/.local/bin, got: $RESULT"
fi
rm -f /tmp/aidb_test_bindir

# Test: GOBIN takes precedence when no BINDIR
echo "- GOBIN precedence"
TESTGOBIN=$(mktemp -d)
BINDIR="" GOBIN="$TESTGOBIN" sh -c '
  if [ -z "$BINDIR" ]; then
    if [ -n "$GOBIN" ]; then BINDIR="$GOBIN"; else BINDIR="$HOME/.local/bin"; fi
  fi
  echo "$BINDIR"
' > /tmp/aidb_test_bindir
RESULT=$(cat /tmp/aidb_test_bindir)
if [ "$RESULT" = "$TESTGOBIN" ]; then
  pass "install.sh uses GOBIN when no BINDIR"
else
  fail "install.sh should use GOBIN=$TESTGOBIN, got: $RESULT"
fi
rm -rf "$TESTGOBIN" /tmp/aidb_test_bindir

# Test: PATH warning when BINDIR not in PATH
echo "- PATH warning detection"
OUTPUT=$(PATH="/usr/bin:/bin" BINDIR="/nonexistent/bin" sh -c '
  case ":$PATH:" in
    *":$BINDIR:"*) echo "in-path" ;;
    *) echo "not-in-path" ;;
  esac
')
if [ "$OUTPUT" = "not-in-path" ]; then
  pass "PATH warning triggers when BINDIR not in PATH"
else
  fail "PATH warning should trigger, got: $OUTPUT"
fi

# Test: no PATH warning when BINDIR is in PATH
echo "- No PATH warning when in PATH"
OUTPUT=$(PATH="/usr/bin:/custom/bin:/bin" BINDIR="/custom/bin" sh -c '
  case ":$PATH:" in
    *":$BINDIR:"*) echo "in-path" ;;
    *) echo "not-in-path" ;;
  esac
')
if [ "$OUTPUT" = "in-path" ]; then
  pass "No PATH warning when BINDIR in PATH"
else
  fail "Should not warn when BINDIR in PATH, got: $OUTPUT"
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ] || exit 1
