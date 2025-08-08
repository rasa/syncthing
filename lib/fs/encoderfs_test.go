// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

import "testing"

var encoderTypes = map[EncoderType]string{
	EncoderTypeNone:   "none",
	EncoderTypeWSL:    "wsl",
	EncoderTypeRclone: "rclone",
	EncoderTypeUnset:  "unset",
}

var filesystemTypes = map[FilesystemType]string{
	FilesystemTypeBasic: "basic",
	FilesystemTypeFake:  "fake",
}

func TestEncoderTypes(t *testing.T) {
	for id := range encoderTypes {
		encoderType := EncoderType(id)
		if string(encoderType) == "unknown" {
			t.Errorf("Missing string for %v in encoderfs_types.go", encoderType)
		}
		text, _ := encoderType.MarshalText()
		var et EncoderType
		_ = et.UnmarshalText(text)
		if et != encoderType {
			t.Errorf("Bad/missing string for %v in encoderfs_types.go",
				encoderType)
		}
	}
}

func TestEncoderOptions(t *testing.T) {
	for id := range encoderTypes {
		encoderType := EncoderType(id)
		opt := EncoderTypeOption(encoderType)
		got := opt.String()
		var want string
		switch encoderType {
		case EncoderTypeWSL:
			want = new(OptionWSLEncoder).String()
		case EncoderTypeRclone:
			want = new(OptionRcloneEncoder).String()
		case EncoderTypeUnset, EncoderTypeNone:
			want = new(OptionNoneEncoder).String()
		default:
			t.Errorf("Missing test for EncoderType %v", encoderType)
		}
		if got != want {
			t.Errorf("FilesystemEncoderOption(%v): got %v, want %v", encoderType, got, want)
		}
	}
}

func TestEncoderNewFilesystem(t *testing.T) {
	testDir := t.TempDir()
	for encoderTypeID := range encoderTypes {
		encoderType := EncoderType(encoderTypeID)
		opts := []Option{EncoderTypeOption(encoderType)}
		for filesystemTypeID := range filesystemTypes {
			filesystemType := FilesystemType(filesystemTypeID)
			fs := NewFilesystem(filesystemType, testDir, opts...)
			switch encoderType {
			case EncoderTypeUnset:
				// s'll good man
			case EncoderTypeNone:
				// s'll good man
				// nfs, ok := unwrapFilesystem[*noneEncoderFS](fs)
				// if !ok {
				// 	t.Fatalf("NewFilesystem(%v) expected to instantiate an encoder",
				// 		encoderType)
				// }
				// got := nfs.EncoderType()
				// if encoderType != got {
				// 	t.Errorf("NewFilesystem(%v) expected %v, got %v",
				// 		encoderType, got, encoderType)
				// }
			case EncoderTypeWSL:
				ffs, ok := unwrapFilesystem[*wslEncoderFS](fs)
				if !ok {
					t.Fatalf("NewFilesystem(%v) expected to instantiate an encoder",
						encoderType)
				}
				got := ffs.EncoderType()
				if encoderType != got {
					t.Errorf("NewFilesystem(%v) expected %v, got %v",
						encoderType, got, encoderType)
				}
			case EncoderTypeRclone:
				ffs, ok := unwrapFilesystem[*rcloneEncoderFS](fs)
				if !ok {
					t.Fatalf("NewFilesystem(%v) expected to instantiate an encoder",
						encoderType)
				}
				got := ffs.EncoderType()
				if encoderType != got {
					t.Errorf("NewFilesystem(%v) expected %v, got %v",
						encoderType, got, encoderType)
				}
			default:
				t.Errorf("Missing test for %v encoder", encoderType)
			}
		}
	}
}
