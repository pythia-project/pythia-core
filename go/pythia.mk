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

GO_DIR := $~
export GOPATH := $(abspath $(GO_DIR))

GO_PACKAGES := pythia

GO_SOURCES := $(shell find $~/src -name '*.go')
GO_TARGETS := $(patsubst $(GOPATH)/%,$(GO_DIR)/%, \
			$(shell go list -f '{{.Target}}' $(addsuffix /...,$(GO_PACKAGES))))

$(call add_target,go,BUILD,Build go code)
all: go
go: $(GO_TARGETS)

$(GO_TARGETS): $(GO_SOURCES)
	go install $(addsuffix /...,$(GO_PACKAGES))

clean::
	-rm -r $(GOPATH)/bin $(GOPATH)/pkg

clear::
	-rm -r $(addprefix $(GO_DIR)/src/,$(filter-out $(GO_PACKAGES),$(shell ls $(GOPATH)/src)))

# vim:set ts=4 sw=4 noet:
