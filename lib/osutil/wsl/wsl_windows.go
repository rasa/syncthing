// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build windows
// +build windows

package wsl

// IsWSL returns true if we're running instead a Windows Subsystem for Linux
// (WSL) environment.
//
// I know, you're asking "why does the Windows build of IsWSL() return false?"
// Well, WSL can run executables built to run on Linux, and those built to run
// on Windows. For example, executing `whoami“ will run /usr/bin/whoami, and
// executing `whoami.exe` will run /mnt/c/Windows/System32/whoami.exe (if
// /mnt/c/Windows/System32 is in the path). But it doesn't appear to me that an
// executable built to run on Windows can tell it was started from inside a WSL
// environment. For example, the program doesn't see the WSL_DISTRO_NAME
// environment variable that other programs run from inside WSL see.
// Hence, this function must return false.
func IsWSL() bool {
	return false
}
