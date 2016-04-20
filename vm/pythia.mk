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

VM_BUILD_DIR := $~/build
VM_CACHE_DIR := $~/cache


################################################################################
## mkrootfs

ROOTFS_INIT := $(VM_BUILD_DIR)/init
ROOTFS_INIT_SOURCES := $~/init.c

$(ROOTFS_INIT): $(ROOTFS_INIT_SOURCES)
	@mkdir -p $(@D)
	$(CC) -static -m32 -O3 -o $@ $(ROOTFS_INIT_SOURCES)
	strip $@

# Environment makefiles can use $(MKROOTFS) to call mkrootfs.
MKROOTFS := $~/mkrootfs.sh -b $(VM_BUILD_DIR) -c $(VM_CACHE_DIR)

# Targets making use of $(MKROOTFS) shall depend on $(MKROOTFS_DEPS)
MKROOTFS_DEPS := $~/mkrootfs.sh $~/functions.sh $(ROOTFS_INIT)


################################################################################
## UML Kernel

UML_CONFIG := $~/uml.config
UML_VERSION := 4.4.6
UML_PATCHES := $~/uml-001-disable-umid.patch \
               $~/uml-002-quiet-startup.patch
export UML := $(VM_OUT_DIR)/uml

UML_DIR := $(VM_BUILD_DIR)/linux-$(UML_VERSION)
UML_TREE := $(UML_DIR)/extracted.stamp
UML_ARCHIVE := $(VM_CACHE_DIR)/linux-$(UML_VERSION).tar.xz
UML_URL := https://www.kernel.org/pub/linux/kernel/v4.x/$(notdir $(UML_ARCHIVE))

UML_MAKE := $(MAKE) -C $(UML_DIR) ARCH=um SUBARCH=i386

$(call add_target,uml,BUILD,Build UML kernel)
all: uml
uml: $(UML)

$(UML): $(UML_CONFIG) $(UML_TREE)
	@mkdir -p $(@D)
	cp $(UML_CONFIG) $(UML_DIR)/.config
	$(UML_MAKE)
	cp $(UML_DIR)/linux $@

# The kernel build directory is secondary, so we can manually remove it after
# the compilation completed and it will not trigger a new extraction and
# compilation.
.SECONDARY: $(UML_TREE)
$(UML_TREE): $(UML_ARCHIVE) $(UML_PATCHES)
	@mkdir -p $(UML_DIR)
	tar -xJ -C $(UML_DIR) -f $(UML_ARCHIVE) --strip-components=1
	for p in $(UML_PATCHES); do \
		patch -d $(UML_DIR) -p1 < $$p; \
	done
	touch $@

$(UML_ARCHIVE):
	@mkdir -p $(@D)
	wget -O $@ $(UML_URL)

$(call add_target,uml_menuconfig,MISC,Configure UML kernel)
uml_menuconfig: $(UML_TREE)
	[ -e $(UML_CONFIG) ] \
		&& cp -p $(UML_CONFIG) $(UML_DIR)/.config \
		|| $(UML_MAKE) defconfig
	$(UML_MAKE) menuconfig
	cp -p $(UML_DIR)/.config $(UML_CONFIG)

$(call add_target,uml_oldconfig,MISC,Upgrade UML kernel configuration)
uml_oldconfig: $(UML_TREE)
	cp -p $(UML_CONFIG) $(UML_DIR)/.config
	$(UML_MAKE) oldconfig
	cp -p $(UML_DIR)/.config $(UML_CONFIG)


################################################################################
## Busybox

# Note: environment targets including busybox shall depend on $(BUSYBOX)

BUSYBOX_CONFIG := $~/busybox.config
BUSYBOX_VERSION := 1.21.1
BUSYBOX := $(VM_BUILD_DIR)/busybox

BUSYBOX_DIR := $(VM_BUILD_DIR)/busybox-$(BUSYBOX_VERSION)
BUSYBOX_TREE := $(BUSYBOX_DIR)/extracted.stamp
BUSYBOX_ARCHIVE := $(VM_CACHE_DIR)/busybox-$(BUSYBOX_VERSION).tar.bz2
BUSYBOX_URL := https://busybox.net/downloads/$(notdir $(BUSYBOX_ARCHIVE))

BUSYBOX_MAKE := $(MAKE) -C $(BUSYBOX_DIR) CC="$(CC) -m32" HOSTCC="$(CC) -m32"

$(call add_target,busybox,BUILD,Build busybox)
busybox: $(BUSYBOX)

$(BUSYBOX): $(BUSYBOX_CONFIG) $(BUSYBOX_TREE)
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
	wget --no-check-certificate -O $@ $(BUSYBOX_URL)

$(call add_target,busybox_menuconfig,MISC,Configure busybox)
busybox_menuconfig: $(BUSYBOX_TREE)
	[ -e $(BUSYBOX_CONFIG) ] \
		&& cp -p $(BUSYBOX_CONFIG) $(BUSYBOX_DIR)/.config \
		|| $(BUSYBOX_MAKE) defconfig
	$(BUSYBOX_MAKE) menuconfig
	cp -p $(BUSYBOX_DIR)/.config $(BUSYBOX_CONFIG)

$(call add_target,busybox_oldconfig,MISC,Upgrade busybox configuration)
busybox_oldconfig: $(BUSYBOX_TREE)
	cp -p $(BUSYBOX_CONFIG) $(BUSYBOX_DIR)/.config
	$(BUSYBOX_MAKE) oldconfig
	cp -p $(BUSYBOX_DIR)/.config $(BUSYBOX_CONFIG)


################################################################################
## Cleaning

clean::
	-rm -r $(VM_BUILD_DIR)

clear::
	-rm -r $(VM_CACHE_DIR)

# vim:set ts=4 sw=4 noet:
