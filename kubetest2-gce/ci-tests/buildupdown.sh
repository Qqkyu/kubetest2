#!/bin/bash

# Copyright 2018 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

REPO_ROOT="${GOPATH}"/src/k8s.io/cloud-provider-gcp

make install
make install-deployer-gce
make install-tester-ginkgo


# TODO(spiffxp): remove this when cloudprovider-gcp has a .bazelversion file
export USE_BAZEL_VERSION=5.3.0
# TODO(spiffxp): remove this when gce-build-up-down job updated to do this,
#                or when bazel 5.3.0 is preinstalled on kubekins image
if [ "${CI}" == "true" ]; then
  go install github.com/bazelbuild/bazelisk@latest
  mkdir -p /tmp/use-bazelisk
  ln -s "$(go env GOPATH)/bin/bazelisk" /tmp/use-bazelisk/bazel
  export PATH="/tmp/use-bazelisk:${PATH}"
fi

if [[ -f "${REPO_ROOT}/ginko-test-package-version.env" ]]; then
  TEST_PACKAGE_VERSION=$(cat "${REPO_ROOT}/ginko-test-package-version.env")
  export TEST_PACKAGE_VERSION
  echo "TEST_PACKAGE_VERSION set to ${TEST_PACKAGE_VERSION}"
else
  export TEST_PACKAGE_VERSION="v1.25.0"
  echo "TEST_PACKAGE_VERSION - Falling back to v1.25.0"
fi;

cd "${GOPATH}/src/k8s.io/kubernetes"
# kubetest2 against k/k
kubetest2 gce \
    -v=2 \
    --repo-root=. \
    --build \
    --up \
    --down \
    --legacy-mode \
    --test=ginkgo \
    --target-build-arch=linux/arm64 \
    --master-size=t2a-standard-2 \
    --node-size=t2a-standard-2 \
    --env=KUBE_MASTER_OS_DISTRIBUTION=ubuntu \
    --env=KUBE_IMAGE_FAMILY=ubuntu-2204-lts-arm64 \
    --env=KUBE_NODE_OS_DISTRIBUTION=ubuntu \
    -- \
    --focus-regex='Secrets should be consumable via the environment' \
    --skip-regex='\[Driver:.gcepd\]|\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]' \
    --timeout=30m

cd "${REPO_ROOT}"
# kubetest2 against cloud-provider-gcp
kubetest2 gce \
    -v=2 \
    --repo-root=. \
    --build \
    --up \
    --down \
    --test=ginkgo \
    --master-size=e2-standard-2 \
    --node-size=e2-standard-2 \
    -- \
    --test-package-version="${TEST_PACKAGE_VERSION}" \
    --focus-regex='Secrets should be consumable via the environment' \
    --skip-regex='\[Driver:.gcepd\]|\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]' \
    --timeout=30m
