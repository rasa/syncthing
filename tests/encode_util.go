// Copyright (C) 2023 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package integration

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/syncthing/syncthing/lib/encoding/fat"
	"github.com/syncthing/syncthing/lib/rand"
)

// srcType is the type of source encoder.
type srcType int

const (
	// srcTypeDecoded generates pre-encoded filenames on the source encoder.
	srcTypeDecoded srcType = iota
	// srcTypeEncoded generates encoded filenames on the source encoder.
	srcTypeEncoded
)

// srcType is the type of dest encoder.
type dstType int

const (
	// dstTypeDecoded saves pre-encoded filenames on the dest encoder.
	dstTypeDecoded dstType = iota
	// dstTypeEncoded saves encoded filenames on the dest encoder.
	dstTypeEncoded
	// dstTypeRejectEncoded saves encoded filenames, but rejects encode
	// filenames on the wire, on the dest encoder.
	dstTypeRejectEncoded
	// dstTypeSkipped indicates a skipped test as FAT filesystems cannot save
	// pre-encoded filenames.
	dstTypeSkipped
)

type walkResults struct {
	found   int
	missing int
}

// generateTreeWithPrefixes generates n files with random data in directory dir
// and returns the number of files created in the directory. prefixes is a
// string array of 0 to 2 elements. If chars is not empty, for each file
// created, the filename will be prefixed with next character in chars. Once all
// characters have been used, they will be reused. So if n an even number, and
// chars contains `_1_2`, then 50% of the files created will begin with `_`,
// and 25% of the files will begin with `1`. prefix contains a common prefix
// for all filenames, so if chars is `_1_2` and prefix is `s`, the first
// filename will be prefixed with 's_' and the second with 's1', etc.
func generateTreeWithPrefixes(t *testing.T, dir string, n int, chars string, prefix string) int {
	t.Helper()

	runes := []rune(chars)
	created := 0
	for i := 0; i < n; i++ {
		// Generate a random string. The first character is the directory
		// name, the rest is the file name.
		rnd := strings.ToLower(rand.String(16))
		sub := rnd[:1]
		file := rnd[1:]
		if len(runes) > 0 {
			// We add underscores so we can easily ignore them via .stignore. It
			// also makes the encoded characters stand out in certain fonts.
			file = "_" + string(runes[i%len(runes)]) + "_" + file
		}
		file = prefix + file
		size := 512<<10 + rand.Intn(1024)<<10 // between 512 KiB and 1.5 MiB
		err := os.MkdirAll(filepath.Join(dir, sub), 0o700)
		if err != nil {
			t.Fatal(err)
		}
		// Create the file with random data.
		lr := io.LimitReader(rand.Reader, int64(size))
		fd, err := os.Create(filepath.Join(dir, sub, file))
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(fd, lr)
		if err != nil {
			t.Fatal(err)
		}
		if err := fd.Close(); err != nil {
			t.Fatal(err)
		}
		created++
	}

	return created
}

// compareTreesByType compares the contents of two directories recursively. It
// reports any differences (other than missing files) as test failures.
// Returns the number of files that were found and missing.
func compareTreesByType(t *testing.T, a, b string, dstType dstType) walkResults {
	t.Helper()

	walkResults := walkResults{0, 0}

	// These will not match, so we ignore them.
	ignore := []string{".", ".stfolder"}

	if err := filepath.Walk(a, func(path string, aInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(a, path)
		if err != nil {
			return err
		}

		// We need to ignore any files under .stfolder, too.
		// See https://github.com/syncthing/syncthing/pull/9525
		if slices.ContainsFunc(ignore, func(ignore string) bool {
			return strings.HasPrefix(rel, ignore)
		}) {
			return nil
		}

		switch dstType {
		case dstTypeEncoded, dstTypeRejectEncoded:
			rel = fat.MustEncode(rel)
		case dstTypeDecoded:
			rel = fat.MustDecode(rel)
		case dstTypeSkipped:
			// added to quiet linter
		}

		isDir := aInfo.IsDir()

		bPath := filepath.Join(b, rel)
		bInfo, err := os.Stat(bPath)
		if err != nil {
			var pathError *fs.PathError
			if errors.As(err, &pathError) {
				err2u := pathError.Unwrap()
				if errors.Is(err2u, os.ErrNotExist) {
					if !isDir {
						walkResults.missing++
					}

					return nil
				}
			}

			return err
		}

		if !isDir {
			walkResults.found++
		}

		if aInfo.IsDir() != bInfo.IsDir() {
			t.Errorf("mismatched directory/file: %q", rel)
		}

		if aInfo.Mode() != bInfo.Mode() {
			t.Errorf("mismatched mode: %q", rel)
		}

		if aInfo.Mode().IsRegular() {
			if !aInfo.ModTime().Equal(bInfo.ModTime()) {
				t.Errorf("mismatched mod time: %q", rel)
			}

			if aInfo.Size() != bInfo.Size() {
				t.Errorf("mismatched size: %q", rel)
			}

			aHash, err := sha256file(path)
			if err != nil {
				return err
			}
			bHash, err := sha256file(bPath)
			if err != nil {
				return err
			}
			if aHash != bHash {
				t.Errorf("mismatched hash: %q", rel)
			}
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	return walkResults
}

// getTempDir returns a temporary directory. If STFSTESTPATH is set, it creates
// that directory, and returns it, otherwise, it returns t.TempDir().
func getTempDir(t *testing.T, prefix string) string {
	t.Helper()

	base := os.Getenv("STFSTESTPATH")
	if base != "" {
		err := os.MkdirAll(base, 0o700)
		if err != nil {
			t.Fatalf("Cannot create directory %v: %v", base, err)
		}
		dir, err := os.MkdirTemp(base, prefix)
		if err != nil {
			t.Fatalf("Cannot create directory in %v: %v", base, err)
		}

		return dir
	}

	return t.TempDir()
}

// cleanup removes the temporary directories after a successful test run.
func cleanup(dirs []string) {
	for _, dir := range dirs {
		_ = os.RemoveAll(dir)
	}
}

// If we want, we could simplify file_util.go, by replacing these functions:

// generateTree generates n files with random data in a temporary directory
// and returns the path to the directory.
// func generateTree(t *testing.T, n int) string {
// 	t.Helper()
// 	dir := t.TempDir()
// 	_ = generateTreeWithPrefixes(t, dir, n, "", "")

// 	return dir
// }

// // compareTrees compares the contents of two directories recursively. It
// // reports any differences as test failures. Returns the number of files
// // that were checked.
// func compareTrees(t *testing.T, a, b string) int {
// 	t.Helper()

// 	// We pass dstTypeSkipped so we don't encode or decode filenames
// 	walkResults := compareTreesByType(t, a, b, dstTypeSkipped)
// 	if walkResults.missing > 0 {
// 		t.Errorf("got %d files, want %d files", walkResults.found, walkResults.found+walkResults.missing)
// 	}

// 	return walkResults.found
// }
