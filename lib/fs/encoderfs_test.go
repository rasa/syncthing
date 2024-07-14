// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

import "testing"

const windowsUNCPrefix = `\\?\`

func TestEncoderTypes(t *testing.T) {
	for id := range FilesystemEncoderType_name {
		encoderType := FilesystemEncoderType(id)
		if string(encoderType) == "unknown" {
			t.Errorf("Missing string for %v in encoderfs_types.go", encoderType)
		}
		text, _ := encoderType.MarshalText()
		var et FilesystemEncoderType
		_ = et.UnmarshalText(text)
		if et != encoderType {
			t.Errorf("Bad/missing string for %v in encoderfs_types.go",
				encoderType)
		}
	}
}

func TestEncoderOptions(t *testing.T) {
	for id := range FilesystemEncoderType_name {
		encoderType := FilesystemEncoderType(id)
		opt := FilesystemEncoderOption(encoderType)
		got := opt.String()
		var want string
		switch encoderType {
		case FilesystemEncoderTypeUnset, FilesystemEncoderTypeNone:
			want = new(OptionNoneEncoder).String()
		case FilesystemEncoderTypeFat:
			want = new(OptionFatEncoder).String()
		default:
			t.Errorf("Missing test for FilesystemEncoderType %v", encoderType)
		}
		if got != want {
			t.Errorf("FilesystemEncoderOption(%v): got %v, want %v", encoderType, got, want)
		}
	}
}

func TestEncoderNewFilesystem(t *testing.T) {
	testDir := t.TempDir()
	for encoderTypeID := range FilesystemEncoderType_name {
		encoderType := FilesystemEncoderType(encoderTypeID)
		opts := []Option{FilesystemEncoderOption(encoderType)}
		for filesystemTypeID := range FilesystemType_name {
			filesystemType := FilesystemType(filesystemTypeID)
			fs := NewFilesystem(filesystemType, testDir, opts...)
			unwrappedFS, ok := unwrapFilesystem(fs, filesystemWrapperTypeEncoder)
			want := encoderType != FilesystemEncoderTypeUnset &&
				encoderType != FilesystemEncoderTypeNone
			if ok != want {
				t.Errorf("NewFilesystem(%v) got %v, want %v re instantiating an encodingFS",
					encoderType, ok, want)
			}
			switch encoderType {
			case FilesystemEncoderTypeUnset, FilesystemEncoderTypeNone:
				// s'll good man
			case FilesystemEncoderTypeFat:
				ffs, ok := unwrappedFS.(*fatEncoderFS)
				if !ok {
					t.Errorf("NewFilesystem(%v) expected to instantiate an encoder",
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
