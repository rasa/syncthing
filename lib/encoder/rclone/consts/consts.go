// Copyright (C) 2026 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

// Package consts contains the constants for the rclone encoder.
// It's a separate package to avoid a circular dependency for the maketables.go
// command.
package consts

import (
	"maps"
	"slices"
	"strings"
)

const (
	// NumChars is the range of characters we might encode (0-0xff).
	NumChars = 0x100 // 256
	// Nevers contains characters we never encode.
	Nevers = "\x00/\\" // \x00, \x2f, \x5c
	// PatternNevers contains the characters we don't encode in glob patterns.
	PatternNevers = "*?" // \x2a, \x3f
)

var encodeMap = map[rune]rune{
	0x00: '␀', // NUL, \u2400
	0x01: '␁', // SOH \u2401
	0x02: '␂', // STX \u2402
	0x03: '␃', // ETX \u2403
	0x04: '␄', // EOT \u2404
	0x05: '␅', // ENQ \u2405
	0x06: '␆', // ACK \u2406
	0x07: '␇', // BEL \u2407
	0x08: '␈', // BS \u2408
	0x09: '␉', // HT \u2409
	0x0A: '␊', // LF \u240a
	0x0B: '␋', // VT \u240b
	0x0C: '␌', // FF \u240c
	0x0D: '␍', // CR \u240d
	0x0E: '␎', // SO \u240e
	0x0F: '␏', // SI \u240f
	0x10: '␐', // DLE \u2410
	0x11: '␑', // DC1 \u2411
	0x12: '␒', // DC2 \u2412
	0x13: '␓', // DC3 \u2413
	0x14: '␔', // DC4 \u2414
	0x15: '␕', // NAK \u2415
	0x16: '␖', // SYN \u2416
	0x17: '␗', // ETB \u2417
	0x18: '␘', // CAN \u2418
	0x19: '␙', // EM \u2419
	0x1A: '␚', // SUB \u241a
	0x1B: '␛', // ESC \u241b
	0x1C: '␜', // FS \u241c
	0x1D: '␝', // GS \u241d
	0x1E: '␞', // RS \u241e
	0x1F: '␟', // US \u241f
	0x22: '＂', // " \uff02
	0x2A: '＊', // * \uff0a
	0x2F: '／', // / \uff0f
	0x3A: '：', // : \uff1a
	0x3C: '＜', // < \uff1c
	0x3E: '＞', // > \uff1e
	0x3F: '？', // ? \uff1f
	0x5C: '＼', // \ \uff3c
	0x7C: '｜', // | \uff5c
	// DEL is a reserved character on android only.
	0x7F: '␡', // DEL \u2421

	// not used:
	// 0x20: '␠', // space ( )
	// 0x2E: '．', // period (.)
}

// Encodes contains the characters the Rclone encoder encodes. Note: Windows
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
