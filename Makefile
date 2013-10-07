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

################################################################################
## Helper functions

# The $~ variable contains the current subdirectory. This variable is only
# valid on first expansion.

~ := .

# The $(call include_subdirs, subdirs...) function includes the pythia.mk
# makefiles from the specified subdirectories, setting $~ accordingly.

define include_subdir_template =
~ := $~/$(1)
include $~/$(1)/pythia.mk
~ := $(patsubst %/$(1),%,$~)
endef

include_subdirs = \
	$(foreach subdir,$(1),$(eval $(call include_subdir_template,$(subdir))))

# Targets and help formatting.
# Every phony target shall be defined with the
# $(call add_target,target,type,helpline) function, where name is the name of
# the target, type is the section in the help where the target belongs (see the
# HELP_*_TARGETS variables below) and helpline is a line of text.

define add_target_template =
.PHONY: $(1)
HELP_$(strip $(2))_TARGETS += printf "  %-20s%s\n" "$(strip $(1))" "$(strip $(3))";
endef

add_target = $(eval $(call add_target_template,$(1),$(2),$(3)))


################################################################################
## Main makefile

HELP_GENERIC_TARGETS :=
HELP_BUILD_TARGETS :=
HELP_MISC_TARGETS :=

# Generic targets

$(call add_target, all, GENERIC, Build all targets)
all:

$(call add_target, clean, GENERIC, Remove build outputs, but keep downloaded files)
clean::

$(call add_target, clear, GENERIC, Remove all build outputs)
clear:: clean

$(call add_target, help, MISC, Print this help)
help:
	@echo "Pythia build system."
	@echo
	@echo "Generic targets:"
	@$(HELP_GENERIC_TARGETS)
	@echo
	@echo "Specific build targets:"
	@$(HELP_BUILD_TARGETS)
	@echo
	@echo "Miscelaneous targets:"
	@$(HELP_MISC_TARGETS)


# Include subdirectories
$(call include_subdirs, go vm environments)


# Safeguard for not using $~ in second expansion
~ := tilde is not valid in commands

# vim:set ts=4 sw=4 noet:
