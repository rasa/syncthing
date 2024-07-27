// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fsutil

func (t VolumeType) String() string {
	switch t {
	case VolumeTypeUnknown:
		return "unknown"
	case VolumeTypeExt:
		return "ext"
	case VolumeTypeFat:
		return "fat"
	default:
		return "unknown"
	}
}

func (t VolumeType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *VolumeType) UnmarshalText(bs []byte) error {
	switch string(bs) {
	case "unknown":
		*t = VolumeTypeUnknown
	case "ext":
		*t = VolumeTypeExt
	case "fat":
		*t = VolumeTypeFat
	default:
		*t = VolumeTypeUnknown
	}
	return nil
}
