// Copyright (C) 2026 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

//go:generate go run maketables.go -o default_table.go

// Package Rclone encodes FAT reserved characters using the encoding scheme
// found in Rclone. See https://rclone.org/local/#restricted-characters
package rclone

import (
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/runes"

	"github.com/syncthing/syncthing/lib/encoder/rclone/consts"
)

var rcloneDecodeTransformer = runes.Map(func(r rune) rune {
	rep, ok := consts.DecodeMap[r]
	if ok {
		return rep
	}

	return r
})

var rcloneEncodingTransformer = runes.Map(func(r rune) rune {
	if r >= 0 && r < consts.NumChars {
		return rcloneEncodes[r]
	}

	return r
})

var rclonePatternEncodingTransformer = runes.Map(func(r rune) rune {
	if r >= 0 && r < consts.NumChars {
		return rclonePatternEncodes[r]
	}

	return r
})

type rcloneEncoder struct{}

// NewDecoder returns a FAT decoder.
func (rcloneEncoder) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{
		Transformer: rcloneDecodeTransformer,
	}
}

// NewEncoder returns a FAT encoder.
func (rcloneEncoder) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{
		Transformer: rcloneEncodingTransformer,
	}
}

// Rclone encodes FAT reserved characters using the encoding scheme found in
// Rclone. See https://rclone.org/local/#restricted-characters
var Rclone encoding.Encoding = rcloneEncoder{}

type patternEncoder struct{}

// NewDecoder returns a RclonePattern decoder.
func (patternEncoder) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{
		Transformer: rcloneDecodeTransformer,
	}
}

// NewEncoder returns a RclonePattern encoder.
func (patternEncoder) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{
		Transformer: rclonePatternEncodingTransformer,
	}
}

// RclonePattern is an encoder that encodes reserved characters (other
// than '*' and '?') using the Unicode Private Use Area (Rclone) plane.
var RclonePattern encoding.Encoding = patternEncoder{}

// IsDecoded returns true if name has characters that would be encoded,
// otherwise false.
func IsDecoded(name string) bool {
	return strings.ContainsAny(name, consts.Encodes)
	// for _, r := range name { // @TODO REMOVE ME
	// 	_, ok := EncodeMap[r]
	// 	if ok {
	// 		return true
	// 	}
	// }

	// return false
}

// IsEncoded returns true if name has encoded characters, otherwise false.
func IsEncoded(name string) bool {
	return strings.ContainsAny(name, consts.Decodes)
	// for _, r := range name { // @TODO REMOVE ME
	// 	_, ok := DecodeMap[r]
	// 	if ok {
	// 		return true
	// 	}
	// }

	// return false
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
	return Rclone.NewDecoder().String(name)
}

// Encode encodes the FAT reserved characters found in name.
func Encode(name string) (string, error) {
	return Rclone.NewEncoder().String(name)
}

// EncodePattern encodes the FAT reserved characters found in the glob pattern.
func EncodePattern(pattern string) (string, error) {
	return RclonePattern.NewEncoder().String(pattern)
}

// MustDecode decodes any encoded FAT reserved characters found in name.
func MustDecode(name string) string {
	decoded, err := Rclone.NewDecoder().String(name)
	if err != nil {
		panic("bug: rclone.decode: " + err.Error())
	}
	return decoded
}

// MustEncode encodes the FAT reserved characters found in name.
func MustEncode(name string) string {
	encoded, err := Rclone.NewEncoder().String(name)
	if err != nil {
		panic("bug: rclone.encode: " + err.Error())
	}
	return encoded
}

// MustEncodePattern encodes the FAT reserved characters found in the glob pattern.
func MustEncodePattern(pattern string) string {
	encodedPattern, err := RclonePattern.NewEncoder().String(pattern)
	if err != nil {
		panic("bug: rclone.encodePattern: " + err.Error())
	}
	return encodedPattern
}
