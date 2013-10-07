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

VM_BUILD_DIR := $(BUILD_DIR)/vm

UML_CONFIG := $~/config
UML_VERSION := 3.5
UML_PATCHES := $~/glibc2.16.patch
UML_OUTPUT := $(VM_BUILD_DIR)/linux

UML_DIR := $(VM_BUILD_DIR)/linux-$(UML_VERSION)
UML_TREE := $(UML_DIR)/extracted.stamp
UML_ARCHIVE := $(UML_DIR).tar.xz
UML_URL := http://www.kernel.org/pub/linux/kernel/v3.x/$(notdir $(UML_ARCHIVE))

UML_MAKE := $(MAKE) -C $(UML_DIR) ARCH=um SUBARCH=i386

$(call add_target,uml,BUILD,Build UML kernel)
all: uml
uml: $(UML_OUTPUT)

$(UML_OUTPUT): $(UML_CONFIG) $(UML_TREE)
	cp $(UML_CONFIG) $(UML_DIR)/.config
	$(UML_MAKE)
	cp $(UML_DIR)/linux $@

# The kernel build directory is secondary, so we can manually remove it after
# the compilation completed and it will not trigger a new extraction and
# compilation.
.SECONDARY: $(UML_TREE)
$(UML_TREE): $(UML_ARCHIVE)
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
uml_oldconfig: checkdepends $(UML_TREE)
	cp -p $(UML_CONFIG) $(UML_DIR)/.config
	$(UML_MAKE) oldconfig
	cp -p $(UML_DIR)/.config $(UML_CONFIG)

clean::
	-rm -rf $(UML_DIR)
	-rm $(UML_OUTPUT)

# vim:set ts=4 sw=4 noet: