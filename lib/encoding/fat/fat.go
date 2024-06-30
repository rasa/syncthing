// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

//go:generate go run maketables.go -o tables.go

// Package fat encodes FAT reserved characters using the Unicode characters
// \uf000-\uf0ff which falls in a Private Use Area (PUA) in the Basic
// Multilingual Plane.
package fat

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/runes"
)

// It's probably best that we err on the side of caution, and only encode/decode
// the characters in Chars, and not the entire \uf000-\uf0ff range. See
// https://github.com/microsoft/WSL/issues/3200#issuecomment-389613611
// We can always expand the range in the future, if needed.
const (
	// BaseRune is the first rune in Unicode's Private Use Area plane.
	BaseRune = rune(0xf000)
	// NumChars is the range of characters we might encode (0-0xff).
	NumChars = 0x100 // 256
	// Encodes contains the characters the FAT encoder encodes.
	Encodes =
	/***/ "\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f" +
		"\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f" +
		`"*:<>?|` // \x22, \x2a, \x3a, \x3c, \x3e, \x3f, \x7c
	// Nevers contains characters we never encode.
	Nevers = "\x00/\\" // \x00, \x2f, \x5c
	// PatternNevers contains the characters we don't encode in glob patterns.
	PatternNevers = "*?" // \x2a, \x3f
)

var puaDecodeTransformer = runes.Map(func(r rune) rune {
	if r < BaseRune || r >= (BaseRune+NumChars) {
		return r
	}
	if puaEncodes[r&^BaseRune] >= BaseRune {
		return r &^ BaseRune
	}

	return r
})

var puaEncodingTransformer = runes.Map(func(r rune) rune {
	if r >= 0 && r < NumChars {
		return puaEncodes[r]
	}

	return r
})

var puaPatternEncodingTransformer = runes.Map(func(r rune) rune {
	if r >= 0 && r < NumChars {
		return puaPatternEncodes[r]
	}

	return r
})

type fatEncoder struct{}

// NewDecoder returns a FAT decoder.
func (fatEncoder) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{
		Transformer: puaDecodeTransformer,
	}
}

// NewEncoder returns a FAT encoder.
func (fatEncoder) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{
		Transformer: puaEncodingTransformer,
	}
}

// PUA encodes FAT reserved characters using the Unicode characters
// \uf000-\uf0ff which falls in a Private Use Area (PUA) in the Basic
// Multilingual Plane.
var PUA encoding.Encoding = fatEncoder{}

type patternEncoder struct{}

// NewDecoder returns a PUAPattern decoder.
func (patternEncoder) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{
		Transformer: puaDecodeTransformer,
	}
}

// NewEncoder returns a PUAPattern encoder.
func (patternEncoder) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{
		Transformer: puaPatternEncodingTransformer,
	}
}

// PUAPattern is an encoder that encodes reserved characters (other
// than '*' and '?') using the Unicode Private Use Area (PUA) plane.
var PUAPattern encoding.Encoding = patternEncoder{}

// IsDecoded returns true if name has characters that would be encoded,
// otherwise false.
func IsDecoded(name string) bool {
	for _, r := range name {
		if r < 0 || r >= NumChars {
			continue
		}
		if puaEncodes[r] >= BaseRune {
			return true
		}
	}

	return false
}

// IsEncoded returns true if name has encoded characters, otherwise false.
func IsEncoded(name string) bool {
	for _, r := range name {
		if r < BaseRune || r >= (BaseRune+NumChars) {
			continue
		}
		if puaEncodes[r&0xff] >= BaseRune {
			return true
		}
	}

	return false
}

// Decode decodes any encoded FAT reserved characters found in name.
//
// The decoder will never return an error, as the only two errors it can
// return, ErrShortDst, and ErrShortSrc, will never occur, as the decoder
// never adds bytes to the destination buffer, it would only return the same
// number of bytes, or a smaller number.
//
// Also, the default buffer size for the encode/decode transformer is 4096.
// By running the tests with:
//
//	STTESTFATMAXLEN=10000 go test .
//
// every transformer buffer size from 1 to 10000 has been tested.
// Neither the decoder or the encoder has ever failed using these buffer values.
//
// See https://github.com/golang/text/blob/9c2f3a21/runes/runes.go#L211
// and https://github.com/golang/text/blob/9c2f3a21/transform/transform.go#L21
func Decode(name string) string {
	decoded, _ := PUA.NewDecoder().String(name)

	return decoded
}

// Encode encodes the FAT reserved characters found in name.
func Encode(name string, pattern bool) (string, error) {
	if pattern {
		return PUAPattern.NewEncoder().String(name)
	}

	return PUA.NewEncoder().String(name)
}
