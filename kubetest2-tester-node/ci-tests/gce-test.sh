#!/bin/bash

# Copyright 2023 The Kubernetes Authors.
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

make install-all

kubetest2 noop \
  --test=node \
  -- \
  --provider=gce \
  --repo-root="${GOPATH}/src/k8s.io/kubernetes" \
  --gcp-zone=us-west1-b \
  --instance-type=e2-standard-2 \
  --focus-regex="\[NodeConformance\]" \
  --test-args='--kubelet-flags="--cgroup-driver=systemd --runtime-cgroups=/system.slice/containerd.service"' \
  --image-config-file="${GOPATH}/src/k8s.io/test-infra/jobs/e2e_node/containerd/image-config-systemd.yaml"
