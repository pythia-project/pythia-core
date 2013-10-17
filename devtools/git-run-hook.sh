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

# Usage: git-run-hook.sh HOOKNAME arguments...
# Runs both local and tracked hook HOOKNAME

set -e -u

hooks_dir=$(git rev-parse --git-dir)/hooks
tracked_dir=$(dirname "$(readlink -e "$0")")/git-hooks

hook=${0##*/}
local_hook=${hooks_dir}/${hook}.local
tracked_hook=${tracked_dir}/${hook}

if [ -x "${local_hook}" ]; then
    "${local_hook}" "$@" || exit $?
fi
if [ -x "${tracked_hook}" ]; then
    "${tracked_hook}" "$@" || exit $?
fi
exit 0
