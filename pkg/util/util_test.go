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

package util

import "testing"

func TestPseudoUniqueSubstring(t *testing.T) {
	testCases := []struct {
		name              string
		uuid              string
		expectedSubstring string
	}{
		{
			name:              "actual uuid",
			uuid:              "09a2565a-7ac6-11eb-a603-2218f636630c",
			expectedSubstring: "09a2565a-7ac6",
		},
		{
			name:              "<= 13 length uuid",
			uuid:              "09a2565a-7ac6",
			expectedSubstring: "09a2565a-7ac6",
		},
		{
			name:              "empty string",
			uuid:              "",
			expectedSubstring: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			actualSubstring := PseudoUniqueSubstring(tc.uuid)
			if actualSubstring != tc.expectedSubstring {
				t.Errorf("invalid substring: expected %s, but got %s", tc.expectedSubstring, actualSubstring)
			}
		})
	}
}
