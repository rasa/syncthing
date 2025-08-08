// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

import (
	"testing"

	"github.com/syncthing/syncthing/lib/encoder/wsl"
)

func newWSLEncoderFS(root string) *wslEncoderFS {
	bfs := newBasicFilesystem(root)
	ffs := new(wslEncoderFS)
	ffs.Filesystem = bfs
	ffs.Encoder = ffs
	ffs.encoderType = EncoderTypeWSL
	ffs.decoder = wsl.WSL.NewDecoder()
	ffs.encoder = wsl.WSL.NewEncoder()
	ffs.patternEncoder = wsl.WSLPattern.NewEncoder()
	ffs.SetRooter(ffs)
	return ffs
}

func TestEncoderWSL(t *testing.T) {
	tempDir := t.TempDir()
	opts := []Option{new(OptionWSLEncoder)}
	fs := NewFilesystem(FilesystemTypeBasic, tempDir, opts...)
	ffs, ok := unwrapFilesystem[*wslEncoderFS](fs)
	if !ok {
		t.Fatalf("NewFilesystem(%v) failed to instantiate a WSL encoder", opts[0].String())
	}
	encoderType := ffs.EncoderType()
	if encoderType != EncoderTypeWSL {
		t.Errorf("NewFilesystem(%v) got %v, want %v",
			EncoderTypeWSL, encoderType, EncoderTypeWSL)
	}
}
