// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build !windows
// +build !windows

package wsl

import (
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/unix"
)

// IsWSL returns true if running under Windows Subsystem for Linux (WSL),
// otherwise false.
func IsWSL() bool {
	var uts unix.Utsname
	err := unix.Uname(&uts)
	if err == nil {
		release := unix.ByteSliceToString(uts.Release[:])
		if strings.Contains(strings.ToLower(release), "microsoft") {
			return true
		}
	}
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err == nil {
		return strings.Contains(strings.ToLower(string(data)), "microsoft")
	}
	data, err = os.ReadFile("/proc/version")
	if err == nil {
		return strings.Contains(strings.ToLower(string(data)), "microsoft")
	}
	path, err := exec.LookPath("wslpath")
	if err != nil {
		return false
	}

	return path == "/usr/bin/wslpath"
}
