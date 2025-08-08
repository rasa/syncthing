// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

import (
	"testing"

	"github.com/syncthing/syncthing/lib/encoder/rclone"
)

func newRcloneEncoderFS(root string) *rcloneEncoderFS {
	bfs := newBasicFilesystem(root)
	ffs := new(rcloneEncoderFS)
	ffs.Filesystem = bfs
	ffs.Encoder = ffs
	ffs.encoderType = EncoderTypeRclone
	ffs.decoder = rclone.Rclone.NewDecoder()
	ffs.encoder = rclone.Rclone.NewEncoder()
	ffs.patternEncoder = rclone.RclonePattern.NewEncoder()
	ffs.SetRooter(ffs)
	return ffs
}

func TestEncoderRclone(t *testing.T) {
	tempDir := t.TempDir()
	opts := []Option{new(OptionRcloneEncoder)}
	fs := NewFilesystem(FilesystemTypeBasic, tempDir, opts...)
	ffs, ok := unwrapFilesystem[*rcloneEncoderFS](fs)
	if !ok {
		t.Fatalf("NewFilesystem(%v) failed to instantiate a Rclone encoder", opts[0].String())
	}
	encoderType := ffs.EncoderType()
	if encoderType != EncoderTypeRclone {
		t.Errorf("NewFilesystem(%v) got %v, want %v",
			EncoderTypeRclone, encoderType, EncoderTypeRclone)
	}
}
