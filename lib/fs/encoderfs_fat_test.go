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
	ffs := new(fatEncoderFS)
	ffs.Filesystem = bfs
	ffs.Encoder = ffs
	ffs.encoderType = EncoderTypeFat
	ffs.decoder = fat.PUA.NewDecoder()
	ffs.encoder = fat.PUA.NewEncoder()
	ffs.patternEncoder = fat.PUAPattern.NewEncoder()
	ffs.SetRooter(ffs)
	return ffs
}

func TestEncoderFAT(t *testing.T) {
	tempDir := t.TempDir()
	opts := []Option{new(OptionFatEncoder)}
	fs := NewFilesystem(FilesystemTypeBasic, tempDir, opts...)
	ffs, ok := unwrapFilesystem[*fatEncoderFS](fs)
	if !ok {
		t.Fatalf("NewFilesystem(%v) failed to instantiate a FAT encoder", opts[0].String())
	}
	encoderType := ffs.EncoderType()
	if encoderType != EncoderTypeFat {
		t.Errorf("NewFilesystem(%v) got %v, want %v",
			EncoderTypeFat, encoderType, EncoderTypeFat)
	}
}
