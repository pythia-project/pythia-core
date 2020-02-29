# Copyright 2013-2014 The Pythia Authors.
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

GO_DIR := $~
GO_BINDIR := $(GO_DIR)/bin
export GOPATH := $(abspath $(GO_DIR))
GO := GOPATH=$(GOPATH) go

GO_PACKAGES := pythia testutils
GO_INSTALL_BINARIES := pythia

GO_SOURCES := $(shell find $(GO_DIR)/src -name '*.go')
GO_TARGETS := $(patsubst $(abspath GO_DIR)/%,$(GO_DIR)/%, \
			$(shell $(GO) list -f '{{.Target}}' $(addsuffix /...,$(GO_PACKAGES))))

GO_OUT_BINARIES := $(addprefix $(OUT_DIR)/,$(GO_INSTALL_BINARIES))

$(call add_target,go,BUILD,Build go code)
all: go
go: $(GO_TARGETS) $(GO_OUT_BINARIES)

$(GO_TARGETS): $(GO_SOURCES)
	$(GO) install $(addsuffix /...,$(GO_PACKAGES))

$(OUT_DIR)/%: $(abspath $(GO_BINDIR))/%
	mkdir -p $(OUT_DIR)
	cp $< $@

clean::
	-rm -r $(GO_DIR)/bin $(GO_DIR)/pkg

clear::
	-rm -r $(addprefix $(GO_DIR)/src/,$(filter-out $(GO_PACKAGES),$(shell ls $(GO_DIR)/src)))

$(call add_target,go_env,MISC,Print Go environment variables)
go_env:
	@$(foreach var,GOPATH TOP_DIR UML VM_OUT_DIR TASKS_OUT_DIR,printf 'export %s="%s"\n' $(var) $($(var));)

$(call add_target,gotest,MISC,Run Go tests)
check: gotest
gotest: $$(ENV_BUSYBOX) $$(TASKS)
	$(GO) test $(addsuffix /...,$(GO_PACKAGES))

$(call add_target,godoc,MISC,Launch godoc server on port 6060)
godoc:
	$(GO)doc -http=localhost:6060 &

# vim:set ts=4 sw=4 noet:
