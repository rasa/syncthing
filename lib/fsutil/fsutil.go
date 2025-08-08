// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

// Package fsutil implements functions that directly access the filesystem,
// whereas the fs package abstracts the underlying OS's filesystem. Currently,
// these functions are only used by the test suite.
package fsutil

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"unicode"

	"github.com/syncthing/syncthing/lib/build"
	"github.com/syncthing/syncthing/lib/encoding/fat/consts"
)

// VolumeType is the disk volume format (ext or fat).
type VolumeType int

const (
	UnixTempPrefix = ".syncthing." // we don't want to depend on lib/fs

	// VolumeTypeUnknown is if we can't determine the volume type at all.
	VolumeTypeUnknown VolumeType = iota
	// VolumeTypeFat is if the volume is FAT or FAT-like, in that the volume
	// rejects filenames with the characters `"*:<>?|` and \x01-\x1f in them.
	// Fat is mixed-case to mimic Ext, and to mimic EncoderTypeFat
	// which is mixed case too.
	VolumeTypeFat
	// VolumeTypeExt is if the volume is not FAT-like, in that it accepts
	// filenames with any characters in them except `/` and NUL (\x00).
	VolumeTypeExt
)

var (
	mux    sync.Mutex
	volMap sync.Map

	// VolumeTypes is the list of valid volume types.
	VolumeTypes = []VolumeType{VolumeTypeFat, VolumeTypeExt}

	// ErrNotADirectory.
	ErrNotADirectory = errors.New("not a directory")
	// ErrCannotCreateDirectory.
	ErrCannotCreateDirectory = errors.New("cannot create a temp directory")
)

// GetVolumeType returns VolumeTypeFat if dir is on a FAT or FAT-like disk
// volume, VolumeTypeExt if it's not, or VolumeTypeUnknown if there's an
// error. The result is cached, as volumes don't change their type.
func GetVolumeType(dir string) (VolumeType, error) {
	mux.Lock()
	defer mux.Unlock()

	dir = filepath.Clean(dir)
	a, found := volMap.Load(dir)
	if found {
		i, _ := a.(int)
		volumeType := VolumeType(i)
		if volumeType != VolumeTypeUnknown {
			return volumeType, nil
		}
	}

	tempDir, err := getTempDir(dir)
	if err != nil {
		return VolumeTypeUnknown, err
	}
	defer os.Remove(tempDir) // ignore errors

	isFat, err := isFat(tempDir)
	if err != nil {
		return VolumeTypeUnknown, err
	}
	volumeType := VolumeTypeExt
	if isFat {
		volumeType = VolumeTypeFat
	}

	volMap.Store(dir, volumeType)

	return volumeType, nil
}

// IsExt returns true if dir is on a Ext or Ext-like disk volume, otherwise
// false.
func IsExt(dir string) (bool, error) {
	volumeType, err := GetVolumeType(dir)
	if err != nil {
		return false, err
	}

	return volumeType == VolumeTypeExt, nil
}

// IsFat returns true if dir is on a FAT or FAT-like disk volume, otherwise
// false.
func IsFat(dir string) (bool, error) {
	volumeType, err := GetVolumeType(dir)
	if err != nil {
		return false, err
	}

	return volumeType == VolumeTypeFat, nil
}

func getTempDir(dir string) (string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", ErrNotADirectory
	}
	dir, err = os.MkdirTemp(dir, UnixTempPrefix)
	if err != nil {
		return "", ErrCannotCreateDirectory
	}
	return dir, nil
}

func isFat(dir string) (bool, error) {
	path := filepath.Join(dir, UnixTempPrefix+".tmp")
	err := os.MkdirAll(path, 0o775)
	if err != nil {
		return false, err
	}
	defer os.Remove(path)

	for _, r := range reservedFATChars() {
		path := filepath.Join(dir, UnixTempPrefix+string(r)+".tmp")
		err := os.MkdirAll(path, 0o775)
		if err == nil {
			_ = os.Remove(path) // ignore errors
			return false, nil
		}
	}

	return true, nil
}

func reservedFATChars() []rune {
	runes := make([]rune, 0, len(consts.Encodes))
	for _, char := range consts.Encodes {
		// Skip over control characters as filenames containing them are rare.
		if unicode.IsControl(char) {
			continue
		}
		// We need to exclude colons from our tests on Windows, because if we try
		// to create a file named `acolon:.txt`, Windows will create a file named
		// `acolon`, with an Alternate Data Stream named `.txt`.
		if build.IsWindows && char == ':' {
			continue
		}
		runes = append(runes, char)
	}

	return runes
}
