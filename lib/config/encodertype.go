// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import "github.com/syncthing/syncthing/lib/fs"

type EncoderType int32

const (
	EncoderTypeNone   EncoderType = 0
	EncoderTypeWSL    EncoderType = 1
	EncoderTypeRclone EncoderType = 2
	EncoderTypeUnset  EncoderType = -1
)

func (t EncoderType) String() string {
	switch t {
	case EncoderTypeNone:
		return "none"
	case EncoderTypeWSL:
		return "wsl"
	case EncoderTypeRclone:
		return "rclone"
	case EncoderTypeUnset:
		return "unset"
	default:
		return "unknown"
	}
}

func (t EncoderType) ToEncoderType() fs.EncoderType {
	switch t {
	case EncoderTypeNone:
		return fs.EncoderTypeNone
	case EncoderTypeWSL:
		return fs.EncoderTypeWSL
	case EncoderTypeRclone:
		return fs.EncoderTypeRclone
	default:
		return fs.EncoderTypeUnset
	}
}

func (t EncoderType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *EncoderType) UnmarshalText(bs []byte) error {
	switch string(bs) {
	case "none":
		*t = EncoderTypeNone
	case "wsl":
		*t = EncoderTypeWSL
	case "rclone":
		*t = EncoderTypeRclone
	default:
		*t = EncoderTypeUnset
	}
	return nil
}

func (t *EncoderType) ParseDefault(str string) error {
	return t.UnmarshalText([]byte(str))
}
