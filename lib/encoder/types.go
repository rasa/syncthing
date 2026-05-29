// Copyright (C) 2026 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

// Package encoder contains the constants shared by all encoders.
package encoder

type EncoderType int32

const (
	// EncoderTypeNone does not encode filenames, and it's only instantiated in
	// our test suite. It is not used in non-test code.
	EncoderTypeNone EncoderType = 0
	// EncoderTypeRclone encodes characters reserved on vFAT/exFAT/NTFS/reFS and
	// similar filesystems. It does not encode filenames ending with spaces or
	// periods, which are accepted on Android, but are often rejected on
	// Windows. It also does not encode Windows' reserved filenames, such as
	// `NUL` or `CON.txt`.
	EncoderTypeRclone EncoderType = 1
	// EncoderTypeWSL encodes characters reserved on vFAT/exFAT/NTFS/reFS and
	// similar filesystems. It does not encode filenames ending with spaces or
	// periods, which are accepted on Android, but are often rejected on
	// Windows. It also does not encode Windows' reserved filenames, such as
	// `NUL` or `CON.txt`.
	EncoderTypeWSL EncoderType = 2
	// EncoderTypeUnset is not a filename encoder. It is only used to allow us
	// to override the default encoder type to WSL on Windows, if the user
	// hasn't set the default themselves.
	EncoderTypeUnset EncoderType = -1
)

func (t EncoderType) String() string {
	switch t {
	case EncoderTypeNone:
		return "none"
	case EncoderTypeRclone:
		return "rclone"
	case EncoderTypeWSL:
		return "wsl"
	case EncoderTypeUnset:
		return "unset"
	default:
		return "unknown"
	}
}

func (t EncoderType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *EncoderType) UnmarshalText(bs []byte) error {
	switch string(bs) {
	case "none":
		*t = EncoderTypeNone
	case "rclone":
		*t = EncoderTypeRclone
	case "wsl":
		*t = EncoderTypeWSL
	default:
		*t = EncoderTypeUnset
	}
	return nil
}
