#!/usr/bin/env bash
set -euo pipefail

# the directory with the patched_rpms
RPM_DIR="${RPM_DIR:-./patched_rpms}"
# do extraction here.
WORKDIR="${WORKDIR:-./_debug_extract}"

# we inspect these in order to validate the manual patch have been applied.
VMLINUX_RPM="${VMLINUX_RPM:-kernel-debuginfo-4.18.0-448.el8.x86_64.rpm}"
COMMON_RPM="${COMMON_RPM:-kernel-debuginfo-common-x86_64-4.18.0-448.el8.x86_64.rpm}"

START_SYMBOL="${START_SYMBOL:-run_posix_cpu_timers}"

need() { command -v "$1" >/dev/null 2>&1 || { echo "Missing $1"; exit 1; }; }
need rpm2cpio
need cpio
need gdb
need realpath
need find

mkdir -p "$WORKDIR"
rm -rf "${WORKDIR:?}/"*

echo "[*] Extracting debuginfo RPMs -> $WORKDIR ..."
(
  cd "$WORKDIR"
  rpm2cpio "../$RPM_DIR/$COMMON_RPM" | cpio -idmv >/dev/null
  rpm2cpio "../$RPM_DIR/$VMLINUX_RPM" | cpio -idmv >/dev/null
)

VMLINUX="$(find "$WORKDIR" -type f -name vmlinux | head -n1)"
[[ -n "$VMLINUX" ]] || { echo "Could not find vmlinux"; exit 1; }

VMLINUX_ABS="$(realpath "$VMLINUX")"
DEBUGSRC_ABS="$(realpath "$WORKDIR/usr/src/debug")"

echo "[*] vmlinux: $VMLINUX_ABS"
echo "[*] debug sources root: $DEBUGSRC_ABS"
echo

##GDB configuration, takes us up to inspect the second patch of the Task3
GDB_INIT="$WORKDIR/.gdbinit_verify"
cat > "$GDB_INIT" <<EOF
set pagination off
set confirm off
set print pretty on
set debuginfod enabled off

# Remap build-time source prefix to our extracted tree, in order to work.
set substitute-path /usr/src/debug $DEBUGSRC_ABS

# (optional) also add to search dirs
directory $DEBUGSRC_ABS

echo \\n[verify.sh] Loaded vmlinux. 
## 2 lists and we can see our patch.
list $START_SYMBOL
list  
EOF

#run gdb
exec gdb -q -x "$GDB_INIT" "$VMLINUX_ABS"
