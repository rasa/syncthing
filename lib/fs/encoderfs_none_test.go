// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

import (
	"testing"
)

func newNoneEncoderFS(root string) *noneEncoderFS {
	bfs := newBasicFilesystem(root)
	nfs := new(noneEncoderFS)
	nfs.Filesystem = bfs
	nfs.Encoder = nfs
	nfs.encoderType = EncoderTypeFat
	nfs.SetRooter(nfs)
	return nfs
}

func TestEncoderNone(t *testing.T) {
	tempDir := t.TempDir()
	opts := []Option{new(OptionNoneEncoder)}
	fs := NewFilesystem(FilesystemTypeBasic, tempDir, opts...)
	_, ok := unwrapFilesystem(fs, filesystemWrapperTypeEncoder)
	if ok {
		t.Errorf("NewFilesystem(%v) expected not to instantiate a None encoder",
			opts[0].String())
	}
}
