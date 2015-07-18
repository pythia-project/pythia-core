# Copyright 2013-2015 The Pythia Authors.
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

TASKS_DIR := $~
export TASKS_OUT_DIR := $(OUT_DIR)/tasks

TASKS_BASE := $(basename $(notdir $(wildcard $(TASKS_DIR)/*.task)))
TASKS := $(foreach t,$(TASKS_BASE),$(TASKS_OUT_DIR)/$(t).task $(TASKS_OUT_DIR)/$(t).sfs)

$(call add_target,tasks,BUILD,Generate test tasks)
all: tasks
tasks: $(TASKS)

$(TASKS_OUT_DIR)/%.task: $(TASKS_DIR)/%.task
	@mkdir -p $(@D)
	cp $< $@

$(TASKS_OUT_DIR)/%.sfs: $$(wildcard $(TASKS_DIR)/$$*/*)
	@mkdir -p $(@D)
	mksquashfs $+ $@ -all-root -comp lzo -noappend

# vim:set ts=4 sw=4 noet:
