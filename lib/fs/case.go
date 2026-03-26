// Copyright (C) 2026 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build !windows

package fs

import (
	"fmt"
	"runtime"
)

// EnableCaseSensitivity sets case sensitivity for the directory path.
func EnableCaseSensitivity(path string) error {
	return nil
}

// DisableCaseSensitivity clears case sensitivity for the directory path.
func DisableCaseSensitivity(path string) error {
	return nil
}

// QueryCaseSensitivity returns whether case sensitivity is enabled for the directory path.
func QueryCaseSensitivity(path string) (bool, error) {
	return false, fmt.Errorf("QueryCaseSensitivity is (yet) not supported on " + runtime.GOOS)
}
