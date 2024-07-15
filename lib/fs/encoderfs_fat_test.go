// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

import (
	"testing"

	"github.com/syncthing/syncthing/lib/encoding/fat"
)

func newFATEncoderFS(root string) *fatEncoderFS {
	bfs := newBasicFilesystem(root)
	ffs := &fatEncoderFS{
		encoderFS: encoderFS{
			Filesystem:  bfs,
			encoderType: EncoderTypeFat,
		},
		decoder:        fat.PUA.NewDecoder(),
		encoder:        fat.PUA.NewEncoder(),
		patternEncoder: fat.PUAPattern.NewEncoder(),
	}
	ffs.Encoder = ffs
	ffs.SetRooter(ffs)
	return ffs
}

func TestEncoderFAT(t *testing.T) {
	tempDir := t.TempDir()
	opts := []Option{new(OptionFatEncoder)}
	fs := NewFilesystem(FilesystemTypeBasic, tempDir, opts...)
	unwrappedFS, ok := unwrapFilesystem(fs, filesystemWrapperTypeEncoder)
	if !ok {
		t.Errorf("NewFilesystem(%v) got %v, want %v",
			EncoderTypeFat, "!filesystemWrapperTypeEncoder",
			"filesystemWrapperTypeEncoder")
	}
	ffs, ok := unwrappedFS.(*fatEncoderFS)
	if !ok {
		t.Errorf("NewFilesystem(%v) failed to instantiate a FAT encoder", opts[0].String())
	}
	encoderType := ffs.EncoderType()
	if encoderType != EncoderTypeFat {
		t.Errorf("NewFilesystem(%v) got %v, want %v",
			EncoderTypeFat, encoderType, EncoderTypeFat)
	}
}
