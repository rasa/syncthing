// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

// build this file on all platforms, as a user may mount a disk that he intends
// to hook up to a Windows systems in the future:
//go:build aix || android || darwin || dragonfly || freebsd || illumos || ios || js || linux || netbsd || openbsd || plan9 || solaris || wasip1 || windows

package fs

import (
	"strings"
	"testing"

	"github.com/syncthing/syncthing/lib/encoding/fat"
)

func newWindowsEncoderFS(root string) *windowsEncoderFS {
	bfs := newBasicFilesystem(root)
	wfs := &windowsEncoderFS{
		encoderFS: encoderFS{
			Filesystem:  bfs,
			encoderType: EncoderTypeFat,
		},
		decoder:        fat.PUA.NewDecoder(),
		encoder:        fat.PUA.NewEncoder(),
		patternEncoder: fat.PUAPattern.NewEncoder(),
	}
	wfs.Encoder = wfs
	wfs.SetRooter(wfs)
	return wfs
}

func TestEncoderWindows(t *testing.T) {
	tempDir := t.TempDir()
	opts := []Option{new(OptionWindowsEncoder)}
	fs := NewFilesystem(FilesystemTypeBasic, tempDir, opts...)
	unwrappedFS, ok := unwrapFilesystem(fs, filesystemWrapperTypeEncoder)
	if !ok {
		t.Errorf("NewFilesystem(%v) got %v, want %v",
			EncoderTypeFat, "!filesystemWrapperTypeEncoder",
			"filesystemWrapperTypeEncoder")
	}
	wfs, ok := unwrappedFS.(*windowsEncoderFS)
	if !ok {
		t.Errorf("NewFilesystem(%v) failed to instantiate a Windows encoder", opts[0].String())
	}
	encoderType := wfs.EncoderType()
	if encoderType != EncoderTypeWindows {
		t.Errorf("NewFilesystem(%v) got %v, want %v",
			EncoderTypeWindows, encoderType, EncoderTypeWindows)
	}
}

type windowsEncoderTestCase struct {
	ins string
	out string
}

func TestEncoderWindowsEncode(t *testing.T) {
	var tests = []windowsEncoderTestCase{
		{"", ""},
		{"/", "/"},
		{".", "."},
		{"..", ".."},
		{"...", "..\uf02e"},
		{"CON", "CO\uf04e"},
		{"CON ", "CON\uf020"},
		{"CON.", "CON\uf02e"},
		{"CON .", "CON \uf02e"},
		{"CON. ", "CO\uf04e.\uf020"},
		{"CON.txt", "CO\uf04e.txt"},
		{" CON", " CON"},
		{" CON ", " CON\uf020"},
		{" CON.", " CON\uf02e"},
		{" CON .", " CON \uf02e"},
		{" CON. ", " CON.\uf020"},
		{" CON.txt", " CON.txt"},
		{".CON", ".CON"},
		{".CON ", ".CON\uf020"},
		{".CON.", ".CON\uf02e"},
		{".CON .", ".CON \uf02e"},
		{".CON. ", ".CON.\uf020"},
		{".CON.txt", ".CON.txt"},
		{"CON0", "CON0"},
		{"CON0 ", "CON0\uf020"},
		{"CON0.", "CON0\uf02e"},
		{"CON0 .", "CON0 \uf02e"},
		{"CON0. ", "CON0.\uf020"},
		{"CON0.txt", "CON0.txt"},
	}

	root := t.TempDir()
	fs := newWindowsEncoderFS(root)
	var slashes = []string{"/"}
	if pathSeparatorString != "/" {
		slashes = append(slashes, pathSeparatorString)
	}
	for _, test := range tests {
		for _, slash := range slashes {
			ins := test.ins
			out := test.out
			ins = strings.ReplaceAll(ins, "/", slash)
			out = strings.ReplaceAll(out, "/", slash)
			want := out
			got, err := fs.encode(ins, false)
			if err != nil {
				t.Errorf("encode(%q): got err %v (%T), want nil", ins, err, err)
			}
			if got != want {
				t.Errorf("encode(%q): got %q, want %q", ins, got, want)
			}
		}
	}
}
