#!/bin/busybox ash
# Copyright 2013 The Pythia Authors.
# This file is part of Pythia.
#
# Pythia is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, version 3 of the License.
#
# Pythia is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with Pythia.  If not, see <http://www.gnu.org/licenses/>.

alias bb=/bin/busybox-pythia

echo
echo "pythia: init"

# Helper functions
die() {
    echo "pythia: $*"
    bb halt -f
}

# Mount essential system filesystems
bb mount -t proc proc /proc -o nosuid,noexec,nodev
bb mount -t sysfs sys /sys -o nosuid,noexec,nodev
bb mount -t tmpfs none /tmp -o nodev,nosuid,mode=777,size=${disksize}m

# Mount task filesystem
bb mount -t squashfs /dev/ubdb /task -o ro || die "Unable to mount task fs."

# Set environment
export PATH=/bin:/usr/bin
export LANG=C
umask 002

# Execute steps
[ -e /task/control ] || die "Missing task control file."
step=0
bb cat /task/control |
while read user cmd; do
    let step++
    # Execute command with requested privileges
    cmd="ulimit -SHm $((ramsize*1024)); ulimit -SHp ${maxproc}; ${cmd}"
    case "${user}" in
    master) bb su -c "${cmd}" master ;;
    worker) bb su -c "${cmd}" worker </dev/null >/dev/null 2>/dev/null ;;
    *) die "Invalid user in task control file: ${user}." ;;
    esac
    exitcode=$?
    # Kill remaining processes
    bb ps -o user,pid |
        bb sed -n '/^root/d; s/^[^ ]* *\([0-9]*\)$/\1/p' |
        bb xargs -r bb kill -9
    # Die on error
    [ ${exitcode} -eq 0 ] || die "Step ${step} exited with code ${exitcode}."
done

# Shutdown
bb umount -a
bb halt -f
