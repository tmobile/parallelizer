## Copyright (c) 2020 Kevin L. Mitchell
##
## Licensed under the Apache License, Version 2.0 (the "License"); you
## may not use this file except in compliance with the License.  You
## may obtain a copy of the License at
##
##      http://www.apache.org/licenses/LICENSE-2.0
##
## Unless required by applicable law or agreed to in writing, software
## distributed under the License is distributed on an "AS IS" BASIS,
## WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
## implied.  See the License for the specific language governing
## permissions and limitations under the License.

# Packages to test; can be overridden at the command line
PACKAGES    = ./...

PKG_ROOT    = $(shell grep '^module ' go.mod | awk '{print $$NF}')

GO          = go
GOFMT       = gofmt
GOIMPORTS   = goimports
GOLINT      = golint

SOURCES     = $(shell find . -name \*.go -print)

COVER_OUT   = coverage.out
COVER_HTML  = coverage.html

CLEAN       = $(COVER_OUT) $(COVER_HTML)

ifeq ($(CI),true)
FORMAT_TARG = format-test
MOD_ARG     = -mod=readonly
else
FORMAT_TARG = format
MOD_ARG     =
endif

all: test

format-test:
	@all=`( \
		$(GOIMPORTS) -l -local $(PKG_ROOT) $(SOURCES); \
		$(GOFMT) -l -s $(SOURCES) \
	) | sort -u`; \
	if [ "$$all" != "" ]; then \
		echo "The following files require formatting updates:"; \
		echo; \
		echo "$$all"; \
		echo; \
		echo "Use \"make format\" to reformat these files."; \
		exit 1; \
	fi

format:
	$(GOIMPORTS) -l -local $(PKG_ROOT) -w $(SOURCES)
	$(GOFMT) -l -s -w $(SOURCES)

lint:
	$(GOLINT) -set_exit_status $(PACKAGES)

vet:
	$(GO) vet $(PACKAGES)

test-only:
	$(GO) test $(MOD_ARG) -race -coverprofile=$(COVER_OUT) $(PACKAGES)

test: $(FORMAT_TARG) lint vet test-only

cover: $(COVER_HTML)

$(COVER_OUT): $(SOURCES)
	$(MAKE) test

$(COVER_HTML): $(COVER_OUT)
	$(GO) tool cover -html=$(COVER_OUT) -o $(COVER_HTML)

clean:
	rm -f $(CLEAN)
