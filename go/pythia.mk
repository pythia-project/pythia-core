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

export GOPATH := $(abspath $~)

# Root pythia package name
GO_PACKAGE := pythia

GO_SOURCES := $(shell find $~/src -name '*.go')
GO_TARGETS := $(patsubst $(GOPATH)/%,$~/%, \
					$(shell go list -f '{{.Target}}' $(GO_PACKAGE)/...))

$(call add_target,go,BUILD,Build go code)
all: go
go: $(GO_TARGETS)

$(GO_TARGETS): $(GO_SOURCES)
	go install $(GO_PACKAGE)/...

clean::
	-rm -r $(GOPATH)/bin $(GOPATH)/pkg

clear::
	-rm -r $(filter-out $(GO_PACKAGE),$(shell ls $(GOPATH)/src))

# vim:set ts=4 sw=4 noet:
