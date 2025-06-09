// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/syncthing/syncthing/lib/rc"
)

const indexDbDir = "index-v2"

var generatedFiles = []string{"config.xml", "cert.pem", "key.pem"}

// From https://github.com/syncthing/syncthing/blob/4e56dbd8/lib/build/build.go#L39
var allowedVersionExp = regexp.MustCompile(`^v\d+\.\d+\.\d+(-[a-z0-9]+)*(\.\d+)*(\+\d+-g[0-9a-f]+|\+[0-9a-z]+)?(-[^\s]+)?$`)

func TestCLIVersion(t *testing.T) {
	// Not parellel or we'll get:
	// The process cannot access the file because it is being used by another process.
	// Also, if this test fails, all tests will fail.

	cmd := exec.Command(syncthingBinary, "version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	err := cmd.Run()
	if err != nil {
		t.Logf("syncthing version returned: %v", err)
	}

	output := stdout.String()
	parts := strings.Split(output, " ")
	if len(parts) < 2 {
		t.Errorf("Expected a space in version output, got %q", output)
		return
	}
	Version := parts[1]
	if Version != "unknown-dev" {
		// If not a generic dev build, version string should come from git describe
		if !allowedVersionExp.MatchString(Version) {
			t.Fatalf("Invalid version string %q;\n\tdoes not match regexp %v;\n\t`syncthing version` returned %q", Version, allowedVersionExp, output)
		}
	}
}

func TestCLIReset(t *testing.T) {
	instance := startInstance(t)

	// This extra work isn't required when testing locally, but on github, we're seeing
	// WARNING: Resetting database: remove C:\Users\RUNNER~1\AppData\Local\Temp\TestCLIReset3617350472\001\index-v0.14.0.db\000002.log: 
	// The process cannot access the file because it is being used by another process.

	api := rc.NewAPI(instance.apiAddress, instance.apiKey)
	t.Log("Calling /rest/system/ping")
	var dst map[string]string
	err := api.Post("/rest/system/ping", nil, &dst)
	if err != nil {
		t.Error(err)
	}

	if dst["ping"] != "pong" {
		t.Errorf("Expecting 'pong', got %q", dst["ping"])
	}
	t.Log("Received pong")

	// Shutdown instance after it created its files in syncthing's home directory.
	err = api.Post("/rest/system/shutdown", nil, nil)
	if err != nil {
		t.Error(err)
	}

	tries := 5
	for tries > 0 {
		tries--
		time.Sleep(1 * time.Second)
		t.Log("Calling /rest/system/ping")
		err = api.Post("/rest/system/ping", nil, &dst)
		if err != nil {
			break
		}
		if dst["ping"] != "pong" {
			t.Errorf("Expecting 'pong', got %q", dst["ping"])
		}
		t.Log("Received pong")	
	}

	dbDir := filepath.Join(instance.syncthingDir, indexDbDir)
	err = os.MkdirAll(dbDir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(syncthingBinary, "--no-browser", "--home", instance.syncthingDir, "debug", "reset-database")
	cmd.Env = basicEnv(instance.userHomeDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(dbDir)
	if err == nil {
		t.Errorf("the directory %q still exists, expected it to have been deleted", dbDir)
	}
}

func TestCLIGenerate(t *testing.T) {
	syncthingDir := t.TempDir()
	userHomeDir := t.TempDir()
	// generateDir := t.TempDir()

	cmd := exec.Command(syncthingBinary, "--no-browser", "--home", syncthingDir, "generate")
	cmd.Env = basicEnv(userHomeDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	found := walk(t, syncthingDir)
	// Sort list so binary search works.
	sort.Strings(found)

	// Verify that the files that should have been created have been.
	for _, want := range generatedFiles {
		_, ok := slices.BinarySearch(found, want)
		if !ok {
			t.Errorf("expected to find %q in %q", want, syncthingDir)
		}
	}
}

func TestCLIFirstStartup(t *testing.T) {
	// Startup instance.
	instance := startInstance(t)

	// Shutdown instance after it created its files in syncthing's home directory.
	api := rc.NewAPI(instance.apiAddress, instance.apiKey)
	err := api.Post("/rest/system/shutdown", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	found := walk(t, instance.syncthingDir)

	// Sort list so binary search works.
	sort.Strings(found)

	// Verify that the files that should have been created have been.
	for _, want := range generatedFiles {
		_, ok := slices.BinarySearch(found, want)
		if !ok {
			t.Errorf("expected to find %q in %q", want, instance.syncthingDir)
		}
	}
}
