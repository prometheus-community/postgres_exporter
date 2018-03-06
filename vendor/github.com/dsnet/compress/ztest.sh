#!/bin/bash
#
# Copyright 2017, Joe Tsai. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE.md file.

cd $(go list -f '{{ .Dir }}' github.com/dsnet/compress)

BOLD="\x1b[1mRunning: "
PASS="\x1b[32mPASS"
FAIL="\x1b[31mFAIL"
RESET="\x1b[0m"

echo -e "${BOLD}fmt${RESET}"
RET_FMT=$(find . -name "*.go" | egrep -v "/(_.*_|\..*|testdata)/" | xargs gofmt -d)
if [[ ! -z "$RET_FMT" ]]; then echo "$RET_FMT"; echo; fi

echo -e "${BOLD}test${RESET}"
RET_TEST=$(go test -race ./... | egrep -v "^(ok|[?])\s+")
if [[ ! -z "$RET_TEST" ]]; then echo "$RET_TEST"; echo; fi

echo -e "${BOLD}staticcheck${RESET}"
RET_SCHK=$(staticcheck \
	-ignore "
		github.com/dsnet/compress/internal/prefix/*.go:SA4016
		github.com/dsnet/compress/brotli/*.go:SA4016
	" ./... 2>&1)
if [[ ! -z "$RET_SCHK" ]]; then echo "$RET_SCHK"; echo; fi

echo -e "${BOLD}vet${RESET}"
RET_VET=$(go vet ./... 2>&1 |
	egrep -v "^flate/dict_decoder.go:(.*)WriteByte" |
	egrep -v "^exit status")
if [[ ! -z "$RET_VET" ]]; then echo "$RET_VET"; echo; fi

echo -e "${BOLD}lint${RESET}"
RET_LINT=$(golint ./... 2>&1 |
	egrep -v "should have comment(.*)or be unexported" |
	egrep -v "^(.*)type name will be used as(.*)by other packages" |
	egrep -v "^brotli/transform.go:(.*)replace i [+]= 1 with i[+]{2}" |
	egrep -v "^internal/prefix/prefix.go:(.*)replace symBits(.*) [-]= 1 with symBits(.*)[-]{2}" |
	egrep -v "^xflate/common.go:(.*)NoCompression should be of the form" |
	egrep -v "^exit status")
if [[ ! -z "$RET_LINT" ]]; then echo "$RET_LINT"; echo; fi

if [[ ! -z "$RET_FMT" ]] || [ ! -z "$RET_TEST" ] || [[ ! -z "$RET_VET" ]] || [[ ! -z "$RET_SCHK" ]] || [[ ! -z "$RET_LINT" ]] || [[ ! -z "$RET_SPELL" ]]; then
	echo -e "${FAIL}${RESET}"; exit 1
else
	echo -e "${PASS}${RESET}"; exit 0
fi
