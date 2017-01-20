#!/usr/local/bin/bash

#http://www.apache.org/licenses/LICENSE-2.0.txt
#
#
#Copyright 2015 Intel Corporation
#
#Licensed under the Apache License, Version 2.0 (the "License");
#you may not use this file except in compliance with the License.
#You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing, software
#distributed under the License is distributed on an "AS IS" BASIS,
#WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#See the License for the specific language governing permissions and
#limitations under the License.

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"

# shellcheck source=scripts/common.sh
. "${__dir}/common.sh"

_info "project path: ${__proj_dir}"

git_version=$(_git_version)
go_build=(go build -ldflags "-w -X main.gitversion=${git_version}")

_info "git commit: $(git log --pretty=format:"%H" -1)"


# rebuild binaries:
export GOOS=${GOOS:-$(uname -s | tr '[:upper:]' '[:lower:]')}
export GOARCH=${GOARCH:-"amd64"}

# Disable CGO for builds by default, enable in freebsd
if [ "$GOOS" != "freebsd" ]; then
   _info "disabling cgo for freebsd"
   export CGO_ENABLED=0
fi


OS=$(uname -s)
if [[ "${OS}" == "Darwin" ]]; then
  p=$(type -p sysctl > /dev/null && sysctl -n hw.ncpu || echo "1")
elif [[ "${OS}" == "Linux" ]]; then
  p=$(type -p nproc > /dev/null && nproc || echo "1")
else
  p="1"
fi
p=${BUILD_JOBS:-"${p}"}

if [[ "${GOARCH}" == "amd64" ]]; then
  build_path="${__proj_dir}/build/${GOOS}/x86_64"
else
  build_path="${__proj_dir}/build/${GOOS}/${GOARCH}"
fi

mkdir -p "${build_path}/plugins"
_info "building plugins for ${GOOS}/${GOARCH} in ${p} parallels"
find "${__proj_dir}/plugin/" -type d -iname "snap-*" -print0 | xargs -0 -n 1 -P $p -I{} "${__dir}/build_plugin.sh" {}
