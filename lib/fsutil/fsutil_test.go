// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fsutil

import (
	"os"
	"strings"
	"testing"
)

func TestGetVolumeType(t *testing.T) {
	for _, volumeType := range VolumeTypes {
		envvar := volumeEnvvar(volumeType)
		tempDir, err := tempDirEnv(t, envvar)
		if err != nil {
			t.Error(err)
		}
		defer os.RemoveAll(tempDir)
		got, err := GetVolumeType(tempDir)
		if err != nil {
			t.Errorf("VolumeType(%v) failed: %v", tempDir, err)
		}
		if got != volumeType {
			t.Logf("VolumeType(%v): got %v, want %v", tempDir, got, volumeType)
		}
	}
}

func TestIsExt(t *testing.T) {
	envvar := volumeEnvvar(VolumeTypeExt)
	tempDir, err := tempDirEnv(t, envvar)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tempDir)
	got, err := IsExt(tempDir)
	if err != nil {
		t.Skipf("IsExt(%v) failed: %v", tempDir, err)
	}
	want := true
	if got != want {
		// Don't fail, as the filesystem may not be ext or ext-like.
		t.Skipf("IsExt(%v): got %v, want %v\nSet %v to test a specific directory",
			tempDir, got, want, envvar)
	}
}

func TestIsFat(t *testing.T) {
	envvar := volumeEnvvar(VolumeTypeFat)
	tempDir, err := tempDirEnv(t, envvar)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tempDir)
	got, err := IsFat(tempDir)
	if err != nil {
		t.Skipf("IsFat(%v) failed: %v", tempDir, err)
	}
	want := true
	if got != want {
		// Don't fail, as the filesystem probably isn't FAT or FAT-like.
		t.Skipf("IsFat(%v): got %v, want %v\nSet %v to test a specific directory",
			tempDir, got, want, envvar)
	}
}

// The following functions are also at the end of lib/fs/encoderfs_matrix_test.go
func volumeEnvvar(volumeType VolumeType) string {
	return "STFSTESTPATH" + strings.ToUpper(volumeType.String())
}

func tempDirEnv(t *testing.T, envvar string) (string, error) {
	t.Helper()

	if envvar != "" {
		dir := os.Getenv(envvar)
		if dir != "" {
			tempDir, err := fsutilTempDir(dir)
			if err != nil {
				t.Errorf("Cannot use directory %q (%v): %v", dir, envvar, err)
				return "", err
			}
			t.Logf("Testing using directory %q (%v)", tempDir, envvar)
			return tempDir, nil
		}
	}

	return t.TempDir(), nil
}

func fsutilTempDir(dir string) (string, error) {
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
