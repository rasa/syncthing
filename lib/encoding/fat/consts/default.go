// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build !android
// +build !android

package consts

// Encodes contains the characters the FAT encoder encodes. Note: Windows
// doesn't reject filenames contains DEL (\x7f) characters but Android does.
// See https://tinyurl.com/msz94d9u .
const Encodes = ("\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f" +
	"\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f" +
	`"*:<>?|`) // \x22, \x2a, \x3a, \x3c, \x3e, \x3f, \x7c
