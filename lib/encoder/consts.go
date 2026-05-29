// Copyright (C) 2026 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package encoder

const (
	// NumChars is the range of characters we might encode (0-0xff).
	NumChars = 0x100 // 256
	// Nevers contains characters we never encode.
	Nevers = "\x00/\\" // \x00, \x2f, \x5c
	// PatternNevers contains the characters we don't encode in glob patterns.
	PatternNevers = "\x00/*?" // \x2a, \x3f
	// EncodeDEL defines if the DEL character (\x7f) will be encoded or not.
	// The DEL characters is a reserved character on Android only, but for
	// compatibility purposes, we encoded it on Windows too.
	EncodeDEL = true
	// DefaultEncoder defines the default encoder on Windows/Android based
	// systems.
	DefaultEncoder = "rclone"
)
