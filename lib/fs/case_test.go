// Copyright (C) 2026 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/syncthing/syncthing/lib/rand"
)

func TestCaseSensitivity(t *testing.T) {
	if !build.IsWindows {
		t.Skip("Only supported on Windows")
		return
	}

	failTests := os.Getenv("STTESTCASESENSITIVITY") != ""
	dir := t.TempDir()
	filename := filepath.Join(dir, rand.String(8)+".tmp")
	path := TempName(filename)
	err := os.MkdirAll(path, 0o770)
	if err != nil {
		t.Fatalf("Cannot create directory %q: %v", path, err)
	}

	_, err = QueryCaseSensitivity(path)
	if err != nil {
		t.Fatalf("QueryCaseSensitivity(%q) failed: %v", path, err)
	}

	err = EnableCaseSensitivity(path)
	if err != nil {
		if failTests {
			t.Fatalf("EnableCaseSensitivity(%q) failed: %v", path, err)
		} else {
			t.Logf("EnableCaseSensitivity(%q) failed: %v", path, err)
		}
	}

	enabled, err := QueryCaseSensitivity(path)
	if err != nil {
		t.Fatalf("QueryCaseSensitivity(%q) failed: %v", path, err)
	}

	if !enabled {
		t.Fatalf("QueryCaseSensitivity(%q) failed: not enabled", path)
	}

	err = DisableCaseSensitivity(path)
	if err != nil {
		if failTests {
			t.Fatalf("EnableCaseSensitivity(%q) failed: %v", path, err)
		} else {
			t.Logf("EnableCaseSensitivity(%q) failed: %v", path, err)
		}
	}

	enabled, err = QueryCaseSensitivity(path)
	if err != nil {
		t.Fatalf("QueryCaseSensitivity(%q) failed: %v", path, err)
	}

	if enabled {
		t.Fatalf("QueryCaseSensitivity(%q) failed: enabled", path)
	}
}
