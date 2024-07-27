// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

// Package consts contains the constants for the fat encoder.
// It's a separate package to avoid a circular dependency for the maketables.go
// command.
package consts

const (
	// BaseRune is the first rune in Unicode's Private Use Area plane.
	BaseRune = rune(0xf000)
	// NumChars is the range of characters we might encode (0-0xff).
	NumChars = 0x100 // 256
	// Nevers contains characters we never encode.
	Nevers = "\x00/\\" // \x00, \x2f, \x5c
	// PatternNevers contains the characters we don't encode in glob patterns.
	PatternNevers = "*?" // \x2a, \x3f
)
