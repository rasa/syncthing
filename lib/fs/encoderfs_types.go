// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

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
	case "unset":
		*t = EncoderTypeUnset
	default:
		*t = EncoderTypeNone
	}
	return nil
}

func (t *EncoderType) ParseDefault(str string) error {
	return t.UnmarshalText([]byte(str))
}
