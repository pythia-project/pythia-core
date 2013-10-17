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


# Install git hooks symlinks

set -e -u

cd "$(dirname "$0")"

hooks="applypatch-msg
   pre-applypatch
   post-applypatch
   pre-commit
   prepare-commit-msg
   commit-msg
   post-commit
   pre-rebase
   post-checkout
   post-merge
   pre-push
   pre-receive
   update
   post-receive
   post-update
   pre-auto-gc
   post-rewrite"

hooks_dir=$(git rev-parse --show-cdup).git/hooks/
tools_dir=$(git rev-parse --show-prefix)

mkdir -p "${hooks_dir}"

echo "Installing git hooks..."

for hook in ${hooks}; do
    install=${hooks_dir}${hook}
    [ ! -f "${install}" ] || mv "${install}" "${install}.local"
    [ ! -e "${install}" ] || continue
    ln -s "../../${tools_dir}git-run-hook.sh" "${install}"
done
