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

BUSYBOX_CONFIG := $~/busybox.config
BUSYBOX_VERSION := 1.21.0
BUSYBOX_OUTPUT := $(MKROOTFS_BUILD_DIR)/busybox

BUSYBOX_DIR := $(MKROOTFS_BUILD_DIR)/busybox-$(BUSYBOX_VERSION)
BUSYBOX_TREE := $(BUSYBOX_DIR)/extracted.stamp
BUSYBOX_ARCHIVE := $(BUSYBOX_DIR).tar.bz2
BUSYBOX_URL := http://busybox.net/downloads/$(notdir $(BUSYBOX_ARCHIVE))

BUSYBOX_MAKE := $(MAKE) -C $(BUSYBOX_DIR) CC="$(CC) -m32" HOSTCC="$(CC) -m32"

MKROOTFS_DEPENDS := \
	$~/mkrootfs.sh \
	$~/busybox.links \
	$~/functions.sh \
	$~/init.sh \
	$(BUSYBOX_OUTPUT)
MKROOTFS := $~/mkrootfs.sh -b $(MKROOTFS_BUILD_DIR)

$(call add_target, busybox, BUILD, Build busybox)
busybox: $(BUSYBOX_OUTPUT)

$(BUSYBOX_OUTPUT): $(BUSYBOX_CONFIG) $(BUSYBOX_TREE)
	cp $(BUSYBOX_CONFIG) $(BUSYBOX_DIR)/.config
	$(BUSYBOX_MAKE)
	cp $(BUSYBOX_DIR)/busybox $@

# The busybox build directory is secondary, so we can manually remove it after
# the compilation completed and it will not trigger a new extraction and
# compilation.
.SECONDARY: $(BUSYBOX_TREE)
$(BUSYBOX_TREE): $(BUSYBOX_ARCHIVE)
	@mkdir -p $(BUSYBOX_DIR)
	tar -xj -C $(BUSYBOX_DIR) -f $(BUSYBOX_ARCHIVE) --strip-components=1
	touch $@

$(BUSYBOX_ARCHIVE):
	@mkdir -p $(@D)
	wget -O $@ $(BUSYBOX_URL)

$(call add_target, busybox_menuconfig, MISC, Configure busybox)
busybox_menuconfig: $(BUSYBOX_TREE)
	[ -e $(BUSYBOX_CONFIG) ] \
		&& cp -p $(BUSYBOX_CONFIG) $(BUSYBOX_DIR)/.config \
		|| $(BUSYBOX_MAKE) defconfig
	$(BUSYBOX_MAKE) menuconfig
	cp -p $(BUSYBOX_DIR)/.config $(BUSYBOX_CONFIG)

$(call add_target, busybox_oldconfig, MISC, Upgrade busybox configuration)
busybox_oldconfig: $(BUSYBOX_TREE)
	cp -p $(BUSYBOX_CONFIG) $(BUSYBOX_DIR)/.config
	$(BUSYBOX_MAKE) oldconfig
	cp -p $(BUSYBOX_DIR)/.config $(BUSYBOX_CONFIG)

clean::
	-rm -rf $(BUSYBOX_DIR)
	-rm $(BUSYBOX_OUTPUT)

# vim:set ts=4 sw=4 noet:
