// Copyright (C) 2016 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

type FilesystemType int32

const (
	FilesystemTypeBasic FilesystemType = 0
	FilesystemTypeFake  FilesystemType = 1
)

func (t FilesystemType) String() string {
	switch t {
	case FilesystemTypeBasic:
		return "basic"
	case FilesystemTypeFake:
		return "fake"
	default:
		return "unknown"
	}
}

type EncoderType int32

const (
	// EncoderTypeNone does not encode filenames, and it's only instantiated in
	// our test suite. It is not used in non-test code.
	EncoderTypeNone EncoderType = 0
	// EncoderTypeFat encodes characters reserved on vFAT/exFAT/NTFS/reFS and
	// similar filesystems. It does not encode filenames ending with spaces or
	// periods, which are accepted on Android, but are often rejected on
	// Windows. It also does not encode Windows' reserved filenames, such as
	// `NUL` or `CON.txt`.
	EncoderTypeFat EncoderType = 1
	// EncoderTypeUnset is not a filename encoder. It is only used to allow us
	// to override the default encoder type to FAT on Windows, if the user
	// hasn't set the default themselves.
	EncoderTypeUnset EncoderType = -1
)

func (t EncoderType) String() string {
	switch t {
	case EncoderTypeNone:
		return "none"
	case EncoderTypeFat:
		return "fat"
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
	case "fat":
		*t = EncoderTypeFat
	default:
		*t = EncoderTypeUnset
	}
	return nil
}
