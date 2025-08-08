// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

// Package consts contains the constants for the wsl encoder.
// It's a separate package to avoid a circular dependency for the maketables.go
// command.
package consts

import (
	"maps"
	"slices"
	"strings"
)

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

var encodeMap = map[rune]rune{
	0x00: '\uf000', // NUL
	0x01: '\uf001', // SOH
	0x02: '\uf002', // STX
	0x03: '\uf003', // ETX
	0x04: '\uf004', // EOT
	0x05: '\uf005', // ENQ
	0x06: '\uf006', // ACK
	0x07: '\uf007', // BEL
	0x08: '\uf008', // BS
	0x09: '\uf009', // HT
	0x0A: '\uf00a', // LF
	0x0B: '\uf00b', // VT
	0x0C: '\uf00c', // FF
	0x0D: '\uf00d', // CR
	0x0E: '\uf00e', // SO
	0x0F: '\uf00f', // SI
	0x10: '\uf010', // DLE
	0x11: '\uf011', // DC1
	0x12: '\uf012', // DC2
	0x13: '\uf013', // DC3
	0x14: '\uf014', // DC4
	0x15: '\uf015', // NAK
	0x16: '\uf016', // SYN
	0x17: '\uf017', // ETB
	0x18: '\uf018', // CAN
	0x19: '\uf019', // EM
	0x1A: '\uf01a', // SUB
	0x1B: '\uf01b', // ESC
	0x1C: '\uf01c', // FS
	0x1D: '\uf01d', // GS
	0x1E: '\uf01e', // RS
	0x1F: '\uf01f', // US
	0x22: '\uf022', // "
	0x2A: '\uf02a', // *
	0x2F: '\uf02f', // /
	0x3A: '\uf03a', // :
	0x3C: '\uf03c', // <
	0x3E: '\uf03e', // >
	0x3F: '\uf03f', // ?
	0x5C: '\uf05c', // \
	0x7C: '\uf07c', // |
	// DEL is a reserved character on android only.
	0x7F: '\uf0ff', // DEL

	// not used:
	// 0x20: '\uf020', // space ( )
	// 0x2E: '\uf02e', // period (.)
}

// Encodes contains the characters the WSL encoder encodes. Note: Windows
// doesn't reject filenames contains DEL (\x7f) characters but Android does.
// See https://tinyurl.com/63epckte .
var Encodes string

// Decodes contains runes that would be decoded.
var Decodes string

// EncodeMap contains a map of runes and their encoded equivalent.
var EncodeMap = map[rune]rune{}

// DecodeMap contains a map of runes and their decoded equivalent.
var DecodeMap = map[rune]rune{}

func init() {
	for _, i := range slices.Sorted(maps.Keys(encodeMap)) {
		b := rune(i)
		if strings.ContainsRune(Nevers, b) {
			continue
		}
		r := encodeMap[b]

		Encodes += string(b)
		EncodeMap[b] = r

		Decodes += string(r)
		DecodeMap[r] = b
	}
}
