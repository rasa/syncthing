// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

//go:generate go run maketables.go -o default_table.go
//go:generate go run maketables.go -android -o android_table.go

// Package fat encodes FAT reserved characters using the Unicode characters
// \uf000-\uf0ff which falls in a Private Use Area (PUA) in the Basic
// Multilingual Plane.
//
// NOTE: We only decode the characters in consts.Encode, and not the entire
// \uf000-\uf0ff range, as this is how Cygwin, GitBash, Msys2, and WSL all work.
// See also https://github.com/microsoft/WSL/issues/3200#issuecomment-389613611.
package fat

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/runes"

	"github.com/syncthing/syncthing/lib/encoding/fat/consts"
)

var puaDecodeTransformer = runes.Map(func(r rune) rune {
	if r < consts.BaseRune || r >= (consts.BaseRune+consts.NumChars) {
		return r
	}
	if puaEncodes[r&^consts.BaseRune] >= consts.BaseRune {
		return r &^ consts.BaseRune
	}

	return r
})

var puaEncodingTransformer = runes.Map(func(r rune) rune {
	if r >= 0 && r < consts.NumChars {
		return puaEncodes[r]
	}

	return r
})

var puaPatternEncodingTransformer = runes.Map(func(r rune) rune {
	if r >= 0 && r < consts.NumChars {
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
		if r < 0 || r >= consts.NumChars {
			continue
		}
		if puaEncodes[r] >= consts.BaseRune {
			return true
		}
	}

	return false
}

// IsEncoded returns true if name has encoded characters, otherwise false.
func IsEncoded(name string) bool {
	for _, r := range name {
		if r < consts.BaseRune || r >= (consts.BaseRune+consts.NumChars) {
			continue
		}
		if puaEncodes[r&0xff] >= consts.BaseRune {
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
func Decode(name string) (string, error) {
	return PUA.NewDecoder().String(name)
}

// Encode encodes the FAT reserved characters found in name.
func Encode(name string) (string, error) {
	return PUA.NewEncoder().String(name)
}

// EncodePattern encodes the FAT reserved characters found in the glob pattern.
func EncodePattern(pattern string) (string, error) {
	return PUAPattern.NewEncoder().String(pattern)
}

// MustDecode decodes any encoded FAT reserved characters found in name.
func MustDecode(name string) string {
	decoded, err := PUA.NewDecoder().String(name)
	if err != nil {
		panic("bug: fat.decode: " + err.Error())
	}
	return decoded
}

// MustEncode encodes the FAT reserved characters found in name.
func MustEncode(name string) string {
	encoded, err := PUA.NewEncoder().String(name)
	if err != nil {
		panic("bug: fat.encode: " + err.Error())
	}
	return encoded
}

// MustEncodePattern encodes the FAT reserved characters found in the glob pattern.
func MustEncodePattern(pattern string) string {
	encodedPattern, err := PUAPattern.NewEncoder().String(pattern)
	if err != nil {
		panic("bug: fat.encodePattern: " + err.Error())
	}
	return encodedPattern
}
