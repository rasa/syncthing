// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import "github.com/syncthing/syncthing/lib/encoder"

type EncoderType encoder.EncoderType

const (
	EncoderTypeNone   = EncoderType(encoder.EncoderTypeNone)
	EncoderTypeRclone = EncoderType(encoder.EncoderTypeRclone)
	EncoderTypeWSL    = EncoderType(encoder.EncoderTypeWSL)
	EncoderTypeUnset  = EncoderType(encoder.EncoderTypeUnset)
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

func (t EncoderType) ToEncoderType() encoder.EncoderType {
	switch t {
	case EncoderTypeNone:
		return encoder.EncoderTypeNone
	case EncoderTypeRclone:
		return encoder.EncoderTypeRclone
	case EncoderTypeWSL:
		return encoder.EncoderTypeWSL
	default:
		return encoder.EncoderTypeUnset
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
