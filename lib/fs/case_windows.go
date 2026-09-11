// Copyright (C) 2026 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build windows

package fs

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Per https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/ns-ntifs-_file_case_sensitive_information
type fileCaseSensitiveInformation struct {
	Flags uint32
}

// EnableCaseSensitivity sets case sensitivity for the directory path.
// NOTE: The path must be an empty folder and on a local NTFS partition.
// Some applications may open a different file then the expected one if two differ only by case.
func EnableCaseSensitivity(path string) error {
	return setCaseSensitivity(path, true)
}

// DisableCaseSensitivity clears case sensitivity for the directory path.
func DisableCaseSensitivity(path string) error {
	return setCaseSensitivity(path, false)
}

// QueryCaseSensitivity returns whether case sensitivity is enabled for the directory path.
func QueryCaseSensitivity(path string) (bool, error) {
	h, err := openDir(path, windows.FILE_READ_ATTRIBUTES)
	if err != nil {
		return false, fmt.Errorf("Cannot open directory %q:: %w", path, err)
	}
	defer windows.CloseHandle(h)

	var info fileCaseSensitiveInformation

	err = windows.GetFileInformationByHandleEx(
		h,
		windows.FileCaseSensitiveInfo,
		(*byte)(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	)
	if err != nil {
		return false, err
	}

	return (info.Flags & windows.FILE_CS_FLAG_CASE_SENSITIVE_DIR) != 0, nil
}

func openDir(path string, access uint32) (windows.Handle, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}

	h, err := windows.CreateFile(
		p,
		access,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return 0, err
	}

	return h, nil
}

func setCaseSensitivity(path string, enable bool) error {
	h, err := openDir(path, windows.FILE_WRITE_ATTRIBUTES)
	for {
		if err != nil {
			break
		}
		defer windows.CloseHandle(h)

		var info fileCaseSensitiveInformation

		if enable {
			info.Flags = windows.FILE_CS_FLAG_CASE_SENSITIVE_DIR
		}

		err = windows.SetFileInformationByHandle(
			h,
			windows.FileCaseSensitiveInfo,
			(*byte)(unsafe.Pointer(&info)),
			uint32(unsafe.Sizeof(info)),
		)
		if err != nil {
			break
		}

		return nil
	}

	modes := map[bool]string{false: "disable", true: "enable"}

	return fmt.Errorf("Cannot %s case sensitivity on %s: %w", modes[enable], path, err)
}
