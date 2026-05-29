// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package wsl_test

import (
	"os"
	"testing"

	"github.com/syncthing/syncthing/lib/build"
	"github.com/syncthing/syncthing/lib/osutil/wsl"
)

func TestIsWSL(t *testing.T) {
	t.Parallel()
	// IsWSL() always returns false in Windows builds, even if the executable
	// is run inside a WSL environment.
	isWSL := false
	if !build.IsWindows {
		isWSL = os.Getenv("WSL_DISTRO_NAME") != ""
	}
	got := wsl.IsWSL()
	if got != isWSL {
		t.Errorf("IsWSL(): got %v, expected %v", got, isWSL)
	}
}
