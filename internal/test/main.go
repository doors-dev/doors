// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

const bothEnv = "BOTH"
const childEnv = "DOORS_E2E_CHILD"

// RunMain executes an e2e test package. When BOTH is set, the current test
// binary is re-run twice: once in default mode and once with LIMIT=1.
func RunMain(run func() int) {
	if os.Getenv(bothEnv) != "" && os.Getenv(childEnv) == "" {
		os.Exit(runBothModes())
	}
	os.Exit(run())
}

func runBothModes() int {
	defaultCode := runMode(false)
	limitCode := runMode(true)
	if defaultCode != 0 {
		return defaultCode
	}
	return limitCode
}

func runMode(limit bool) int {
	mode := "default"
	if limit {
		mode = "limit"
	}
	fmt.Fprintf(os.Stderr, "\n== BOTH mode: %s ==\n", mode)

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	env := make([]string, 0, len(os.Environ())+2)
	for _, item := range os.Environ() {
		if hasEnvKey(item, childEnv) || hasEnvKey(item, "LIMIT") {
			continue
		}
		env = append(env, item)
	}
	env = append(env, childEnv+"=1")
	if limit {
		env = append(env, "LIMIT=1")
	}
	cmd.Env = env

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		log.Fatal(err)
	}
	return 0
}

func hasEnvKey(item string, key string) bool {
	return len(item) > len(key) && item[:len(key)] == key && item[len(key)] == '='
}
