#!/bin/sh
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

set -e -u

if [ ${EUID} -ne 0 ]; then
    exec fakeroot "$0" "$@"
fi

# Paths
script_dir=$(readlink -f ${0%/*})
build_dir=${script_dir}/build
cache_dir=${script_dir}/cache

# Default values
keep_work=false
unset out_file work_dir

# Helper functions
msg() {
    echo "$(tput bold)$(tput setaf ${2:-4})::$(tput sgr0)$(tput bold) $1$(tput sgr0)"
}
err() {
    msg "$1" 1
}
cleanup() {
    ${keep_work} || [ -z "${work_dir:-}" ] || rm -rf "${work_dir}"
}
trap cleanup EXIT

# Parse arguments
usage() {
    cat <<EOF
Usage: $0 [options] -o FILE SCRIPT...

  SCRIPT...        Configuration scripts to execute
  -o FILE          Set output file
  -b DIR           Set directory containing prebuilds (default: ${build_dir})
  -c DIR           Set download cache firectory (default: ${cache_dir})
  -k               Keep temporary files
  -h               This help message
EOF
    exit ${1}
}

while getopts 'o:b:c:kh' arg; do
    case "${arg}" in
        o) out_file="${OPTARG}" ;;
        b) build_dir=$(readlink -f "${OPTARG}") ;;
        c) cache_dir=$(readlink -f "${OPTARG}") ;;
        k) keep_work=true ;;
        h) usage 0 ;;
        *) usage 1 ;;
    esac
done

shift $((OPTIND-1))

if [ -z "${out_file:-}" ]; then
    err "No output file specified."
    usage 1
fi

# Create work directory with base structure
work_dir=$(mktemp -d)
msg "Creating base rootfs structure in ${work_dir}..."
mkdir -p "${work_dir}"/{dev,proc,sys,tmp,etc,bin,usr/bin,task}

# Execute configuration scripts
for f in "$@"; do (
    msg "Executing configuration script ${f}..."
    f=$(readlink -f "${f}")
    cd "${work_dir}"
    . "${script_dir}/functions.sh"
    . "${f}"
) done

# Remove unwanted files and folders
msg "Removing unneeded folders..."
rm -rf "${work_dir}"/{,usr/}sbin
rm -rf "${work_dir}"/usr/share/{applications,binfmts,doc,info,lintian,locale,man,menu,pixmaps}
rm -rf "${work_dir}/tmp"  # Empty /tmp
mkdir -p "${work_dir}/tmp"

# Create static dev nodes
msg "Populating /dev..."
while read name type major minor perm; do
    mknod -m $perm "${work_dir}/dev/$name" $type $major $minor
done <<EOF
null            c       1       3       666
zero            c       1       5       666
full            c       1       7       666
random          c       1       8       666
urandom         c       1       9       666
console         c       5       1       600
ubda            b       98      0       400
ubdb            b       98      16      400
ubdc            b       98      32      400
EOF

# Create users
msg "Creating user accounts..."
cat >"${work_dir}/etc/passwd" <<EOF
root:!:0:0::/:/bin/false
master:!:1:1::/tmp:/bin/false
worker:!:2:2::/tmp:/bin/false
EOF
cat >"${work_dir}/etc/group" <<EOF
root:!:0:
master:!:1:
worker:!:2:
EOF

# Copy init program
msg "Installing init program..."
install -m0755 "${build_dir}/init" "${work_dir}/init"

msg "Cleaning up..."

# Make everything owned by root:root
# This also removes any SUID/SGID flag, which is a Good Thing (TM)
chown -RP 0:0 "${work_dir}"

# Build root squashfs image
msg "Building squashfs image..."
mksquashfs "${work_dir}" "${out_file}" -noappend -no-xattrs -comp lzo
