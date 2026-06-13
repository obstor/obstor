/*
 * MinIO Cloud Storage, (C) 2017 MinIO, Inc.
 * PGG Obstor, (C) 2021-2026 PGG, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/urfave/cli"
)

// Test RegisterBackendCommand
func TestRegisterBackendCommand(t *testing.T) {
	var err error

	cmd := cli.Command{Name: "test"}
	err = RegisterBackendCommand(cmd)
	if err != nil {
		t.Errorf("RegisterBackendCommand got unexpected error: %s", err)
	}
}

// Test running a registered backend command with a flag
func TestRunRegisteredBackendCommand(t *testing.T) {
	var err error

	flagName := "test-flag"
	flagValue := "foo"

	cmd := cli.Command{
		Name: "test-run-with-flag",
		Flags: []cli.Flag{
			cli.StringFlag{Name: flagName},
		},
		Action: func(ctx *cli.Context) {
			if actual := ctx.String(flagName); actual != flagValue {
				t.Errorf("value of %s expects %s, but got %s", flagName, flagValue, actual)
			}
		},
	}

	err = RegisterBackendCommand(cmd)
	if err != nil {
		t.Errorf("RegisterBackendCommand got unexpected error: %s", err)
	}

	if err = newApp("obstor").Run(
		[]string{"obstor", "backend", cmd.Name, fmt.Sprintf("--%s", flagName), flagValue}); err != nil {
		t.Errorf("running registered backend command got unexpected error: %s", err)
	}
}

// Test parseBackendEndpoint
func TestParseBackendEndpoint(t *testing.T) {
	testCases := []struct {
		arg         string
		endPoint    string
		secure      bool
		errReturned bool
	}{
		{"http://127.0.0.1:9000", "127.0.0.1:9000", false, false},
		{"https://127.0.0.1:9000", "127.0.0.1:9000", true, false},
		{"http://demo.obstor.net:9000", "demo.obstor.net:9000", false, false},
		{"https://demo.obstor.net:9000", "demo.obstor.net:9000", true, false},
		{"ftp://127.0.0.1:9000", "", false, true},
		{"ftp://demo.obstor.net:9000", "", false, true},
		{"demo.obstor.net:9000", "demo.obstor.net:9000", true, false},
	}

	for i, test := range testCases {
		endPoint, secure, err := ParseBackendEndpoint(test.arg)
		errReturned := err != nil

		if endPoint != test.endPoint ||
			secure != test.secure ||
			errReturned != test.errReturned {
			t.Errorf("Test %d: expected %s,%t,%t got %s,%t,%t",
				i+1, test.endPoint, test.secure, test.errReturned,
				endPoint, secure, errReturned)
		}
	}
}

// Test validateBackendArguments
func TestValidateBackendArguments(t *testing.T) {
	nonLoopBackIPs := localIP4.FuncMatch(func(ip string, matchString string) bool {
		return !strings.HasPrefix(ip, "127.")
	}, "")
	if len(nonLoopBackIPs) == 0 {
		t.Fatalf("No non-loop back IP address found for this host")
	}
	nonLoopBackIP := nonLoopBackIPs.ToSlice()[0]

	testCases := []struct {
		serverAddr   string
		endpointAddr string
		valid        bool
	}{
		{":9000", "http://localhost:9001", true},
		{":9000", "http://google.com", true},
		{"123.123.123.123:9000", "http://localhost:9000", false},
		{":9000", "http://localhost:9000", false},
		{":9000", nonLoopBackIP + ":9000", false},
	}
	for i, test := range testCases {
		err := ValidateBackendArguments(test.serverAddr, test.endpointAddr)
		if test.valid && err != nil {
			t.Errorf("Test %d expected not to return error but got %s", i+1, err)
		}
		if !test.valid && err == nil {
			t.Errorf("Test %d expected to fail but it did not", i+1)
		}
	}
}
