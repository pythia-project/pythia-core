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

ENVIRONMENTS_DIR := $~/build

clean::
	-rm -r $(ENVIRONMENTS_DIR)

# The environments target is filled by the subdirectories
$(call add_target,environments,BUILD,Generate all environments)
all: environments
environments:

$(call include_subdirs, python)

# vim:set ts=4 sw=4 noet:
