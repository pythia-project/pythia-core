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

MKROOTFS_BUILD_DIR := $(BUILD_DIR)/mkrootfs

MKROOTFS_INIT := $(MKROOTFS_BUILD_DIR)/init
MKROOTFS_INIT_SOURCES := $~/init.c

$(MKROOTFS_INIT): $(MKROOTFS_INIT_SOURCES)
	$(CC) -static -m32 -O3 -o $@ $*
	strip $@

MKROOTFS_DEPENDS := \
	$~/mkrootfs.sh \
	$~/functions.sh \
	$(MKROOTFS_INIT)
MKROOTFS := $~/mkrootfs.sh -b $(MKROOTFS_BUILD_DIR)

clean::
	-rm $(MKROOTFS_INIT)

# vim:set ts=4 sw=4 noet:
