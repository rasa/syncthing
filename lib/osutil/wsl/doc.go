// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

// Package wsl has Windows/WSL-specific functions. It compiles in all
// GOOSes, as the WSL functions are used by the Linux build, even though
// Windows is under the hood.
//
// This is a separate package in order to break an import cycle with the lib/fs
// package, which uses this package in its tests.
package wsl
