// Copyright (C) 2023 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package integration

import (
	"errors"
	"fmt"
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	rencoder "github.com/syncthing/syncthing/lib/encoder/rclone"
	wencoder "github.com/syncthing/syncthing/lib/encoder/wsl"
	"github.com/syncthing/syncthing/lib/fs"
	"github.com/syncthing/syncthing/lib/rand"
)

type TestMode int // @TODO REMOVE ME

const ( // @TODO REMOVE ME
	testModeNone TestMode = iota
	testModeWSL
	testModeRclone
)

var testMode TestMode = testModeWSL

var verbose = strings.Contains(os.Getenv("STTRACE"), "verbose")

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

var dstTypeMap = map[dstType]string{
	dstTypeDecoded:       "decoded",
	dstTypeEncoded:       "encoded",
	dstTypeRejectEncoded: "non-encoded",
	dstTypeSkipped:       "<skipped>", // not used
}

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
	subdirs := false
	for i := 0; i < n; i++ {
		// Generate a random string. The first character is the directory
		// name, the rest is the file name.
		rnd := strings.ToLower(rand.String(6))
		file := rnd
		if subdirs {
			dir = filepath.Join(dir, rnd[:1])
			file = rnd[1:]
		}
		if len(runes) > 0 {
			// We add underscores so we can easily ignore them via .stignore. It
			// also makes the encoded characters stand out in certain fonts.
			r := runes[i%len(runes)]
			file = "_" + string(r) + fmt.Sprintf("_%04x_", r) + file
			// file = fmt.Sprintf("_%c_%04x_", r, r) + file
		}
		file = prefix + file
		size := 512<<10 + rand.Intn(1024)<<10 // between 512 KiB and 1.5 MiB
		err := os.MkdirAll(dir, 0o700)
		if err != nil {
			t.Fatal(err)
		}
		// Create the file with random data.
		lr := io.LimitReader(rand.Reader, int64(size))
		fd, err := os.Create(filepath.Join(dir, file))
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
func compareTreesByType(t *testing.T, srcDir, dstDir string, dstType dstType, dstEncoder fs.EncoderType) walkResults {
	t.Helper()

	// walkResults := walkResults{0, 0}

	type files struct {
		srcBase string
		dstBase string
	}

	var found []files
	var missing []files
	var errmsgs []string
	var ignoreds []string

	// These will not match, so we ignore them.
	ignore := []string{".", ".stfolder"}

	var encode func(string) string
	var decode func(string) string

	switch dstEncoder {
	case fs.EncoderTypeWSL:
		encode = wencoder.MustEncode
		decode = wencoder.MustDecode
	case fs.EncoderTypeRclone:
		encode = rencoder.MustEncode
		decode = rencoder.MustDecode
	case fs.EncoderTypeNone:
		switch testMode {
		case testModeNone:
			encode = func(s string) string { return s }
			decode = func(s string) string { return s }
		case testModeWSL:
			encode = wencoder.MustEncode
			decode = wencoder.MustDecode
		case testModeRclone:
			encode = rencoder.MustEncode
			decode = rencoder.MustDecode
		}
	default:
		panic(fmt.Sprintf("bug: unexpected dstEncoder %v", dstEncoder))
	}

	if err := filepath.Walk(srcDir, func(srcPath string, srcInfo os.FileInfo, err error) error {
		if err != nil {
			errmsgs = append(errmsgs, err.Error())
			return err
		}

		rel, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			errmsgs = append(errmsgs, err.Error())
			return err
		}

		// We need to ignore any files under .stfolder, too.
		// See https://github.com/syncthing/syncthing/pull/9525
		if slices.ContainsFunc(ignore, func(ignore string) bool {
			return strings.HasPrefix(rel, ignore)
		}) {
			ignoreds = append(ignoreds, rel)
			return nil
		}

		srcBase := filepath.Base(rel)

		switch dstType {
		case dstTypeEncoded, dstTypeRejectEncoded:
			rel = encode(rel)
		case dstTypeDecoded:
			rel = decode(rel)
		case dstTypeSkipped:
			// added to quiet linter
		default:
			panic(fmt.Sprintf("bug: unexpected dstType %v", dstType))
		}

		dstBase := filepath.Base(rel)

		isDir := srcInfo.IsDir()

		dstPath := filepath.Join(dstDir, rel)
		dstInfo, err := os.Stat(dstPath)
		if err != nil {
			var pathError *iofs.PathError
			if errors.As(err, &pathError) {
				err2u := pathError.Unwrap()
				if errors.Is(err2u, os.ErrNotExist) {
					if !isDir {
						missing = append(missing, files{srcBase, dstBase})
					}

					return nil
				}
			}

			errmsgs = append(errmsgs, err.Error())
			return err
		}

		if !isDir {
			found = append(found, files{srcBase, dstBase})
		}

		if srcInfo.IsDir() != dstInfo.IsDir() {
			t.Errorf("mismatched directory/file: %q", rel)
		}

		if srcInfo.Mode() != dstInfo.Mode() {
			t.Errorf("mismatched mode: %q", rel)
		}

		if srcInfo.Mode().IsRegular() {
			if !srcInfo.ModTime().Equal(dstInfo.ModTime()) {
				t.Errorf("mismatched mod time: %q", rel)
			}

			if srcInfo.Size() != dstInfo.Size() {
				t.Errorf("mismatched size: %q", rel)
			}

			srcHash, err := sha256file(srcPath)
			if err != nil {
				return err
			}
			dstHash, err := sha256file(dstPath)
			if err != nil {
				return err
			}
			if srcHash != dstHash {
				t.Errorf("mismatched hash: %q", rel)
			}
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if verbose {
		for i, pair := range found {
			t.Logf("found    %3d: %-12v: dst: %-30q src: %-30q", i+1, dstTypeMap[dstType], pair.dstBase, pair.srcBase)
		}

		for i, pair := range missing {
			t.Logf("missing  %3d: %-12v: dst: %-30q src: %-30q", i+1, dstTypeMap[dstType], pair.dstBase, pair.srcBase)
		}

		for i, error := range errmsgs {
			t.Logf("error   %3d: %v", i+1, error)
		}

		// for i, ignored := range ignoreds {
		// 	t.Logf("ignored %3d: %v", i+1, ignored)
		// }
	}

	return walkResults{len(found), len(missing)}
}

// getTempDir returns a temporary directory. If STFSTESTPATH is set, it creates
// that directory, and returns it, otherwise, it returns t.TempDir().
func getTempDir(t *testing.T, prefix string) string {
	t.Helper()

	base := getFsTextPath()
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
	if getFsTextPath() != "" {
		return
	}
	for _, dir := range dirs {
		_ = os.RemoveAll(dir)
	}
}

func getFsTextPath() string {
	return os.Getenv("STFSTESTPATH")
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
