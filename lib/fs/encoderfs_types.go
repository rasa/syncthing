// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

func (t FilesystemEncoderType) String() string {
	switch t {
	case FilesystemEncoderTypeNone:
		return "none"
	case FilesystemEncoderTypeFat:
		return "fat"
	case FilesystemEncoderTypeUnset:
		return "unset"
	default:
		return "unknown"
	}
}

func (t FilesystemEncoderType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *FilesystemEncoderType) UnmarshalText(bs []byte) error {
	switch string(bs) {
	case "none":
		*t = FilesystemEncoderTypeNone
	case "fat":
		*t = FilesystemEncoderTypeFat
	case "unset":
		*t = FilesystemEncoderTypeUnset
	default:
		*t = FilesystemEncoderTypeNone
	}
	return nil
}

func (t *FilesystemEncoderType) ParseDefault(str string) error {
	return t.UnmarshalText([]byte(str))
}
