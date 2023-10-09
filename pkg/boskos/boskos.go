/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package boskos

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/klog/v2"
	"sigs.k8s.io/boskos/client"
	"sigs.k8s.io/boskos/common"
)

// const (for the run) owner string for consistency between up and down
var boskosOwner = os.Getenv("JOB_NAME") + "-kubetest2"

// NewClient creates a boskos client for kubetest2 deployers.
func NewClient(boskosLocation string) (*client.Client, error) {
	boskos, err := client.NewClient(
		boskosOwner,
		boskosLocation,
		"",
		"",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create boskos client: %s", err)
	}

	return boskos, nil
}

// Acquire acquires a resource for the given type and starts a heartbeat goroutine to keep the resource reserved.
func Acquire(boskosClient *client.Client, resourceType string, timeout, heartbeatInterval time.Duration, heartbeatClose chan struct{}) (*common.Resource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	boskosResource, err := boskosClient.AcquireWait(ctx, resourceType, "free", "busy")
	if err != nil {
		return nil, fmt.Errorf("failed to get a %q from boskos: %s", resourceType, err)
	}
	if boskosResource == nil {
		return nil, fmt.Errorf("boskos had no %s available", resourceType)
	}

	if heartbeatInterval != 0 {
		startBoskosHeartbeat(
			boskosClient,
			boskosResource,
			heartbeatInterval,
			heartbeatClose,
		)
	}

	return boskosResource, nil
}

// startBoskosHeartbeat starts a goroutine that sends periodic updates to boskos
// about the provided resource until the channel is closed. This prevents
// reaper from taking the resource from the deployer while it is still in use.
func startBoskosHeartbeat(boskosClient *client.Client, resource *common.Resource, interval time.Duration, close chan struct{}) {
	go func(c *client.Client, resource *common.Resource) {
		klog.V(2).Info("boskos hearbeat starting")

		for {
			select {
			case <-close:
				klog.V(2).Info("Boskos heartbeat func received signal to close")
				return
			case <-time.NewTicker(interval).C:
				klog.V(2).Info("Sending heartbeat to Boskos")
				if err := c.UpdateOne(resource.Name, "busy", nil); err != nil {
					klog.Warningf("[Boskos] Update of %s failed with %v", resource.Name, err)
				}
			}
		}
	}(boskosClient, resource)
}

// Release releases a resource.
func Release(client *client.Client, resourceNames []string, heartbeatClose chan struct{}) error {
	for _, name := range resourceNames {
		if err := client.Release(name, "dirty"); err != nil {
			return fmt.Errorf("failed to release %s: %s", name, err)
		}
	}
	close(heartbeatClose)
	return nil
}
