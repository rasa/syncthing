// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

import (
	"errors"
	"fmt"
	"io"
	iofs "io/fs"
	"math"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/syncthing/syncthing/lib/build"
	"github.com/syncthing/syncthing/lib/encoding/fat"
	"github.com/syncthing/syncthing/lib/encoding/fat/consts"
	"github.com/syncthing/syncthing/lib/fsutil"
	"github.com/syncthing/syncthing/lib/osutil/wsl"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type encoderTest struct {
	decodeOK bool
	encodeOK bool
}

//             Decoded   Encoded
// Vol Encoder Filenames Filenames
// --- ------- --------- ---------
// Ext None    ok        ok
// Ext FAT     ok        fail
// Fat None    fail      ok
// Fat FAT     ok        fail

var encoderTestMatrix = map[fsutil.VolumeType]map[EncoderType]encoderTest{
	fsutil.VolumeTypeExt: {
		EncoderTypeNone: {
			decodeOK: true,
			encodeOK: true,
		},
		EncoderTypeFat: {
			decodeOK: true,
			encodeOK: false,
		},
	},
	fsutil.VolumeTypeFat: {
		EncoderTypeNone: {
			decodeOK: false,
			encodeOK: true,
		},
		EncoderTypeFat: {
			decodeOK: true,
			encodeOK: false,
		},
	},
}

type globTest struct {
	pattern   string
	recursive bool
}

type (
	namePair  [2]string
	namePairs []namePair
)

type testFunc func(name string) error

type statResult struct {
	name      string
	size      int
	modTime   time.Time
	isRegular bool
	isDir     bool
	isSymlink bool
}

var (
	msgNoSymlinks      sync.Once
	msgFailsOnFat      sync.Once
	msgNoWindowsDecode sync.Once
	msgIsExt           sync.Once
	msgIsFAT           sync.Once
	totalTests         int
)

var (
	isFATMap = make(map[fsutil.VolumeType]bool, len(fsutil.VolumeTypes))
	isFAT    bool
)

func TestEncoderMatrix(tttt *testing.T) {
	_, ok := unwrapFilesystem(testFs, filesystemWrapperTypeEncoder)
	if ok {
		// TestMain runs thru all the tests twice. Once using a plain BasicFilesytem,
		// and again using an fatEncoderFS. So we can skip this test in
		// TestMain's 2nd pass as it's redundant.
		tttt.Skipf("Skipping as TestMain has already run this test in pass 1")
		return
	}

	testEncoderCheckForNewTypes(tttt)
	testEncoderCheckVolumeTypes(tttt)

	// The linter complains that strings.Title() is deprecated.
	title := func(s string) string {
		return cases.Title(language.English).String(s)
	}

	keepDirs := strings.Contains(strings.ToLower(os.Getenv("STTRACE")), "keepdirs")

	// BasicFS / FakeFS
	for filesystemTypeID := range filesystemTypes {
		filesystemType := FilesystemType(filesystemTypeID)
		tttName := title(filesystemType.String()) + "FS"
		tttt.Run(tttName, func(ttt *testing.T) {
			// NoneEncoder / FatEncoder
			for encoderTypeID := range encoderTypes {
				encoderType := EncoderType(encoderTypeID)
				if encoderType == EncoderTypeUnset {
					continue
				}
				ttName := title(encoderType.String()) + "Encoder"
				ttt.Run(ttName, func(tt *testing.T) {
					// ExtVolume / FatVolume
					for _, volumeType := range fsutil.VolumeTypes {
						tName := title(volumeType.String()) + "Volume"
						tt.Run(tName, func(t *testing.T) {
							var rootURI string
							var err error
							if filesystemType == FilesystemTypeFake {
								rootURI = fmt.Sprintf("/%s?nostfolder=true&volume=%s&encoder=%s",
									path.Join(tttName, ttName, tName), volumeType, encoderType)
							}
							isFAT = volumeType == fsutil.VolumeTypeFat
							if filesystemType == FilesystemTypeBasic {
								isFAT = isFATMap[volumeType]
								switch volumeType {
								case fsutil.VolumeTypeExt:
									if isFAT {
										msgIsFAT.Do(func() {
											tt.Logf("Skipping %v/%v test as %v is a FAT-like filesystem",
												tttName, ttName, rootURI)
										})
										return
									}
								case fsutil.VolumeTypeFat:
									if !isFAT {
										msgIsExt.Do(func() {
											tt.Logf("Skipping %v/%v test as %v is not a FAT-like filesystem",
												tttName, ttName, rootURI)
										})
										return
									}
								}

								rootURI, err = testEncoderGetRootURI(t, volumeType)
								if !keepDirs {
									// Remove the temp dir itself, not just all its files.
									defer os.RemoveAll(rootURI)
								}
								if err != nil {
									tt.Fatal(err)
								}
							}
							// ok checks are done in testEncoderCheckForNewTypes()
							encoderTests := encoderTestMatrix[volumeType]
							encoderTest := encoderTests[encoderType]
							opts := []Option{EncoderTypeOption(encoderType)}
							fs := NewFilesystem(filesystemType, rootURI, opts...)
							runEncoderTests(t, fs, encoderTest)
						})
					}
				})
			}
		})
	}
	tttt.Logf("%d total tests run", totalTests)
}

func testEncoderCheckForNewTypes(t *testing.T) {
	// BasicFS / FakeFS
	for filesystemTypeID := range filesystemTypes {
		filesystemType := FilesystemType(filesystemTypeID)
		switch filesystemType {
		case FilesystemTypeBasic, FilesystemTypeFake:
		default:
			t.Fatalf("bug: need to add FilesystemType %q to test matrix",
				filesystemType)
		}
		// NoneEncoder / FatEncoder
		for encoderTypeID := range encoderTypes {
			encoderType := EncoderType(encoderTypeID)
			if encoderType == EncoderTypeUnset {
				continue
			}
			// ExtVolume / FatVolume
			for _, volumeType := range fsutil.VolumeTypes {
				switch volumeType {
				case fsutil.VolumeTypeExt:
				case fsutil.VolumeTypeFat:
				default:
					t.Fatalf("bug: need to add volumeType %q to switch statement", volumeType)
				}
				encoderTests, ok := encoderTestMatrix[volumeType]
				if !ok {
					t.Fatalf("bug: need to add volumeType %q to test matrix", volumeType)
				}
				_, ok = encoderTests[encoderType]
				if !ok {
					t.Fatalf("bug: need to add encoderType %q to test matrix", encoderType)
				}
			}
		}
	}
}

func testEncoderCheckVolumeTypes(t *testing.T) {
	t.Helper()

	for _, volumeType := range fsutil.VolumeTypes {
		dir, err := testEncoderGetRootURI(t, volumeType)
		if err != nil {
			msg := "Cannot create a temporary directory"
			envvar := volumeEnvvar(volumeType)
			value := os.Getenv(envvar)
			if value != "" {
				msg += fmt.Sprintf("Should %v be %q?", envvar, value)
			}
			t.Fatal(msg)
		}
		defer os.RemoveAll(dir)
		isFATLike, err := fsutil.IsFat(dir)
		if err != nil {
			t.Fatal(err)
		}
		isFATMap[volumeType] = isFATLike
		warning := ""
		switch volumeType {
		case fsutil.VolumeTypeExt:
			if isFATLike {
				warning = "(FAT-like filesystem)"
			}
		case fsutil.VolumeTypeFat:
			if !isFATLike {
				warning = "(Non-FAT filesystem)"
			}
		default:
			t.Fatalf("bug: need to add volumeType %q to switch statement", volumeType)
		}
		t.Logf("%sVolume: using %q %v", volumeType.String(), dir, warning)
	}
}

func runEncoderTests(t *testing.T, fs Filesystem, encTest encoderTest) {
	t.Helper()

	// DirNames0
	var testFunc testFunc = func(_ string) error {
		if _, err := fs.DirNames("/"); err != nil {
			return err
		}

		return nil
	}
	t.Run("DirNames0", func(t *testing.T) {
		// Calling DirNames("/") with no files should always succeed so we override
		// default encoderTest entry. We test this first as it's used to verify
		// the tests following it.
		runEncoderTest(t, fs, encoderTest{true, true}, testFunc)
	})

	// Create
	// Next we test Create, as we need to create files for all succeeding tests.
	testFunc = func(name string) error {
		fd, err := fs.Create(name)
		if err != nil {
			return _err(err)
		}
		defer fd.Close()
		writeLen := len(name)
		if written, err := fd.Write([]byte(name)); err != nil {
			return _err(err)
		} else if written != writeLen {
			return _errf("write error: wrote %d bytes, want %d bytes",
				written, writeLen)
		}
		if err = fd.Sync(); err != nil {
			return _err(err)
		}
		if ret, err := fd.Seek(0, io.SeekStart); err != nil {
			return _err(err)
		} else if ret != 0 {
			return _errf("seek error: got %d, want %d", ret, 0)
		}
		bytes := make([]byte, writeLen)
		if read, err := fd.Read(bytes); err != nil {
			return _err(err)
		} else if read != writeLen {
			return _errf("read error: read %d bytes, want %d bytes",
				read, writeLen)
		}
		if fd.Name() != name {
			return _errf("fd.Name(): got %v, want %v", fd.Name(), name)
		}

		return nil
	}
	t.Run("Create", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// DirNames
	// The first time testing DirNames and Glob, we test using all generated
	// filenames.
	maxNames := -1
	testFunc = func(name string) error {
		names, err := testEncoderGetGlobFiles(fs, name, maxNames, false)
		if err != nil {
			return err
		}
		// On the remaining tests we only need to test with one file, as the first
		// test performed all the testing needed, but we need at least one file to
		// test with on these succeeding tests, or the test will fail.
		maxNames = 1
		entries, err := fs.DirNames("/")
		if err != nil {
			return err
		}
		if !testEncoderDirsEqual(entries, names, false) {
			return _errf("got %v, want %v", entries, names)
		}

		return nil
	}
	t.Run("DirNames", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// Chtimes
	testFunc = func(name string) error {
		if err := testEncoderCreateFile(fs, name); err != nil {
			return err
		}
		now := time.Now()

		return fs.Chtimes(name, now, now)
	}
	t.Run("Chtimes", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// Mkdir
	testFunc = func(name string) error {
		return fs.Mkdir(name, 0o775)
	}
	t.Run("Mkdir", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// MkdirAll
	testFunc = func(name string) error {
		if err := fs.MkdirAll(name, 0o775); err != nil {
			return err
		}

		want := statResult{
			name:  name,
			isDir: true,
		}

		if err := testEncoderStat(t, fs, want); err != nil {
			return err
		}

		return nil
	}
	t.Run("MkdirAll", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// Open
	testFunc = func(name string) error {
		if err := testEncoderCreateFile(fs, name); err != nil {
			return err
		}
		fd, err := fs.Open(name)
		if err != nil {
			return err
		}
		fd.Close()

		return nil
	}
	t.Run("Open", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// OpenFile
	testFunc = func(name string) error {
		fd, err := fs.OpenFile(name, os.O_RDONLY, 0o664)
		if err == nil {
			fd.Close()
			return _errf("got no error opening a non-existing file %q", name)
		}

		if fd, err = fs.OpenFile(name, os.O_RDWR|os.O_CREATE, 0o664); err != nil {
			return err
		}
		fd.Close()

		if fd, err = fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o664); err == nil {
			fd.Close()
			return _errf("created an existing file while told not to: %q", name)
		}

		if fd, err = fs.OpenFile(name, os.O_RDWR|os.O_CREATE, 0o664); err != nil {
			return err
		}
		fd.Close()

		if fd, err = fs.OpenFile(name, os.O_RDWR, 0o664); err != nil {
			return err
		}
		fd.Close()

		return nil
	}
	t.Run("OpenFile", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	//  Remove
	testFunc = func(name string) error {
		if err := testEncoderCreateFile(fs, name); err != nil {
			return err
		}
		if err := fs.Remove(name); err != nil {
			return err
		}

		return nil
	}
	t.Run("Remove", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// RemoveAll
	testFunc = func(name string) error {
		if err := testEncoderCreateFile(fs, name); err != nil {
			return err
		}
		if err := fs.RemoveAll(name); err != nil {
			return err
		}

		return nil
	}
	t.Run("RemoveAll", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// Rename
	testFunc = func(name string) error {
		pairs, err := testEncoderNameToNamePairs(t, fs, name, true)
		if err != nil {
			return err
		}
		for _, pair := range pairs {
			oldName := pair[0]
			newName := pair[1]
			if err := testEncoderCreateFile(fs, oldName); err != nil {
				return err
			}
			if err := fs.Rename(oldName, newName); err != nil {
				return err
			}
		}

		return nil
	}
	t.Run("Rename", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// Stat
	testFunc = func(name string) error {
		if err := testEncoderCreateFile(fs, name); err != nil {
			return err
		}

		want := statResult{
			name:      name,
			size:      len(name),
			isRegular: true,
		}

		return testEncoderStat(t, fs, want)
	}
	t.Run("Stat", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// Stat-Dir
	testFunc = func(name string) error {
		dirs := []string{
			filepath.ToSlash(path.Join("reg-1", name)),
			filepath.ToSlash(path.Join(name+"-2", "reg-2")),
			filepath.ToSlash(path.Join(name+"-3", name)),
		}
		for _, dir := range dirs {
			if err := fs.MkdirAll(dir, 0o775); err != nil {
				return err
			}

			want := statResult{
				name:  dir,
				isDir: true,
			}

			if err := testEncoderStat(t, fs, want); err != nil {
				return err
			}
		}

		return nil
	}
	t.Run("Stat-Dir", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// Glob
	maxNames = -1 // See DirNames above
	testFunc = func(name string) error {
		tests := []globTest{
			{"*", false},
			{"0x*", false},
			{"?x*", false},
			{"?x?*", false},
		}

		if fs.Type() != FilesystemTypeFake {
			recursiveTests := []globTest{
				{"*/*", true},
				{"*/0x*", true},
				{"*/?x*", true},
				{"*/?x?*", true},
				{"0x*/*", true},
				{"?x*/*", true},
				{"?x?*/*", true},
				{"0x*/0x*", true},
				{"?x*/0x*", true},
				{"?x?*/0x*", true},
				{"0x*/?x*", true},
				{"0x*/?x?*", true},
				{"?x*/?x*", true},
				{"?x?*/?x*", true},
				{"?x?*/?x?*", true},
			}
			tests = append(tests, recursiveTests...)
		}

		for _, test := range tests {
			names, err := testEncoderGetGlobFiles(fs, name, maxNames, test.recursive)
			if err != nil {
				return err
			}
			entries, err := fs.Glob(test.pattern)
			if err != nil {
				return err
			}
			if !testEncoderDirsEqual(entries, names, test.recursive) {
				return _errf("%v: got %v, want %v", test, entries, names)
			}
			testEncoderCleanup(t, fs)
		}
		maxNames = 1

		return nil
	}
	t.Run("Glob", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// File.Stat
	testFunc = func(name string) error {
		err := testEncoderCreateFile(fs, name)
		if err != nil {
			return err
		}
		fd, err := fs.Open(name)
		if err != nil {
			return err
		}
		defer fd.Close()
		fi, err := fd.Stat()
		if err != nil {
			return err
		}

		want := statResult{
			name:      name,
			size:      len(name),
			isRegular: true,
		}

		return testEncoderFileInfo(t, fs, want, fi)
	}
	t.Run("File.Stat", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// SameFileDiff
	testFunc = func(name string) error {
		pairs, err := testEncoderNameToNamePairs(t, fs, name, true)
		if err != nil {
			return err
		}
		for _, pair := range pairs {
			file1 := pair[0]
			file2 := pair[1]
			if err := testEncoderCreateFile(fs, file1); err != nil {
				return err
			}
			if err := testEncoderCreateFile(fs, file2); err != nil {
				return err
			}
		}
		for _, pair := range pairs {
			file1 := pair[0]
			file2 := pair[1]
			fi1, err := fs.Stat(file1)
			if err != nil {
				return err
			}
			fi2, err := fs.Stat(file2)
			if err != nil {
				return err
			}
			if b := fs.SameFile(fi1, fi2); b {
				return _errf("got %v, want %v for files '%v' and '%v'",
					b, false, file1, file2)
			}
		}

		return nil
	}
	t.Run("SameFileDiff", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// SameFileSame
	testFunc = func(name string) error {
		if err := testEncoderCreateFile(fs, name); err != nil {
			return err
		}
		fi1, err := fs.Stat(name)
		if err != nil {
			return err
		}
		fi2, err := fs.Stat(name)
		if err != nil {
			return err
		}
		if b := fs.SameFile(fi1, fi2); !b {
			return _errf("got %v, want %v for file %v", b, true, name)
		}

		return nil
	}
	t.Run("SameFileSame", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	if isFAT {
		msgFailsOnFat.Do(func() {
			t.Skipf("Skipping chmod/chown/symlink tests as FAT-like filesystems don't support them")
		})

		return
	}

	// Chmod
	testFunc = func(name string) error {
		if err := testEncoderCreateFile(fs, name); err != nil {
			return err
		}

		return fs.Chmod(name, 0o664)
	}
	t.Run("Chmod", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// Lchown
	testFunc = func(name string) error {
		err := testEncoderCreateFile(fs, name)
		if err != nil {
			return err
		}

		newUID := os.Getuid()
		newGID := os.Getgid()

		if os.Getuid() == 0 {
			// Our tests typically don't run with CAP_FOWNER.
			newUID = 1000 + rand.Intn(30000)
			newGID = 1000 + rand.Intn(30000)
		}

		if err = fs.Lchown(name, strconv.Itoa(newUID), strconv.Itoa(newGID)); err != nil {
			return err
		}
		fi, err := fs.Lstat(name)
		if err != nil {
			return err
		}

		if fi.Owner() != newUID {
			return _errf("Owner(): got %v, want %v", fi.Owner(), newUID)
		}
		if fi.Group() != newGID {
			return _errf("Group(): got %v, want %v", fi.Group(), newGID)
		}

		return nil
	}
	t.Run("Lchown", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	if !fs.SymlinksSupported() {
		msgNoSymlinks.Do(func() {
			t.Skipf("Skipping symlink tests as the %q filesystem doesn't support symlinks",
				fs.Type())
		})

		return
	}

	// CreateSymlink
	testFunc = func(name string) error {
		pairs, err := testEncoderNameToNamePairs(t, fs, name, false)
		if err != nil {
			return err
		}
		for _, pair := range pairs {
			target := pair[0]
			symlink := pair[1]
			if err = testEncoderCreateFile(fs, target); err != nil {
				return err
			}
			if err = fs.CreateSymlink(target, symlink); err != nil {
				return err
			}
		}

		return nil
	}
	t.Run("CreateSymlink", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// ReadSymlink
	testFunc = func(name string) error {
		pairs, err := testEncoderNameToNamePairs(t, fs, name, false)
		if err != nil {
			return err
		}
		for _, pair := range pairs {
			target := pair[0]
			symlink := pair[1]
			if err := testEncoderCreateFile(fs, target); err != nil {
				return err
			}
			if err := fs.CreateSymlink(target, symlink); err != nil {
				return err
			}
			if str, err := fs.ReadSymlink(symlink); err != nil {
				return err
			} else if str != target {
				return _errf("Wrong symlink destination got %v (%q), want %v (%q)",
					str, str, target, target)
			}
		}

		return nil
	}
	t.Run("ReadSymlink", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// Lstat
	testFunc = func(name string) error {
		pairs, err := testEncoderNameToNamePairs(t, fs, name, false)
		if err != nil {
			return err
		}
		for _, pair := range pairs {
			target := pair[0]
			symlink := pair[1]
			if err = testEncoderCreateFile(fs, target); err != nil {
				return err
			}
			if err = fs.CreateSymlink(target, symlink); err != nil {
				return err
			}
			if str, err := fs.ReadSymlink(symlink); err != nil {
				return err
			} else if str != target {
				return _errf("Wrong symlink destination: got %s, want %s",
					str, target)
			}
			fi, err := fs.Lstat(symlink)
			if err != nil {
				return err
			}

			want := statResult{
				name:      symlink,
				size:      len(symlink),
				isSymlink: true,
			}

			err = testEncoderFileInfo(t, fs, want, fi)
			if err != nil {
				return err
			}
		}

		return nil
	}
	t.Run("Lstat", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// Lchown-Symlink
	testFunc = func(name string) error {
		var err error
		pairs := namePairs{{name + "1", name + "2"}}
		for _, pair := range pairs {
			target := pair[0]
			symlink := pair[1]
			if err = testEncoderCreateFile(fs, target); err != nil {
				return err
			}
			if err = fs.CreateSymlink(target, symlink); err != nil {
				return err
			}
			if str, err := fs.ReadSymlink(symlink); err != nil {
				return err
			} else if str != target {
				return _errf("Wrong symlink destination: got %s, want %s",
					str, target)
			}

			newUID := os.Getuid()
			newGID := os.Getgid()

			if os.Getuid() == 0 {
				// Our tests typically don't run with CAP_FOWNER.
				newUID = 1000 + rand.Intn(30000)
				newGID = 1000 + rand.Intn(30000)
			}

			if err = fs.Lchown(symlink, strconv.Itoa(newUID), strconv.Itoa(newGID)); err != nil {
				return err
			}
			fi, err := fs.Lstat(symlink)
			if err != nil {
				return err
			}

			if fi.Owner() != newUID {
				return _errf("Owner(): got %v, want %v", fi.Owner(), newUID)
			}
			if fi.Group() != newGID {
				return _errf("Group(): got %v, want %v", fi.Group(), newGID)
			}
		}

		return nil
	}
	t.Run("Lchown-Symlink", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })

	// Stat-Symlink
	testFunc = func(name string) error {
		pairs, err := testEncoderNameToNamePairs(t, fs, name, false)
		if err != nil {
			return err
		}
		for _, pair := range pairs {
			target := pair[0]
			symlink := pair[1]
			if err = testEncoderCreateFile(fs, target); err != nil {
				return err
			}
			if err = fs.CreateSymlink(target, symlink); err != nil {
				return err
			}
			isDir := strings.Contains(filepath.ToSlash(target), "/")

			want := statResult{
				name:      symlink,
				size:      len(target),
				isRegular: !isDir,
				isDir:     isDir,
			}

			err = testEncoderStat(t, fs, want)
			if err != nil {
				return err
			}
		}

		return nil
	}
	t.Run("Stat-Symlink", func(t *testing.T) { runEncoderTest(t, fs, encTest, testFunc) })
}

func runEncoderTest(t *testing.T, fs Filesystem, encTest encoderTest, testFunc testFunc) {
	t.Helper()
	testName := path.Base(t.Name())

	regularFiles, decodedFiles, encodedFiles := testEncoderGetTestFilenames()

	want := map[bool]string{
		false: "no error",
		true:  "an error of some sort",
	}

	for _, name := range regularFiles {
		testEncoderCleanup(t, fs)
		err := testFunc(name)
		totalTests++
		if err != nil {
			testEncoderDumpDir(t, fs)
			t.Errorf("%v: %s: got %q (%T), want %v",
				testName, name, err, err, nil)
		}
	}

	for _, name := range encodedFiles {
		testEncoderCleanup(t, fs)
		err := testFunc(name)
		totalTests++
		ok := err == nil
		if encTest.encodeOK && ok {
			continue
		}
		if encTest.encodeOK != ok {
			t.Errorf("%v: %v: got '%v' (%T), want '%v'",
				testName, name, err, err, want[encTest.encodeOK])
			testEncoderDumpDir(t, fs)
			continue
		}
		testEncoderIsExpectedError(t, testName, name, err)
	}

	if build.IsWindows || wsl.IsWSL() {
		msgNoWindowsDecode.Do(func() {
			t.Log("Cannot run decoding tests in Cygwin/GitBash/Msys2/WSL/Windows environments.")
		})
		return
	}

	for _, name := range decodedFiles {
		testEncoderCleanup(t, fs)
		err := testFunc(name)
		totalTests++
		ok := err == nil
		if encTest.decodeOK && ok {
			continue
		}
		if encTest.decodeOK != ok {
			t.Errorf("%v: %v: got '%v' (%T), want '%v'",
				testName, name, err, err, want[encTest.decodeOK])
			testEncoderDumpDir(t, fs)
			continue
		}
		testEncoderIsExpectedError(t, testName, name, err)
	}
}

func testEncoderIsExpectedError(t *testing.T, testName string, name string, err error) {
	t.Helper()

	if errors.Is(err, os.ErrNotExist) {
		return
	}

	err1, ok := err.(*testEncoderError)
	if ok {
		err = err1.Unwrap()
	}

	err2, ok := err.(*iofs.PathError)
	if ok {
		err2u := err2.Unwrap()
		if errors.Is(err2u, os.ErrNotExist) {
			return
		}
		if strings.Contains(err2u.Error(), "invalid argument") {
			return
		}
		t.Logf("%v: %v: got unexpected PathError: %v (unwrapped: %v)", testName, name, err2, err2u)
	}

	err3, ok := err.(*os.LinkError)
	if ok {
		err3u := err3.Unwrap()
		if errors.Is(err3u, os.ErrNotExist) {
			return
		}
		if strings.Contains(err3u.Error(), "invalid argument") {
			return
		}
		t.Logf("%v: %v: got unexpected LinkError: %v (unwrapped: %v)", testName, name, err3, err3u)
	}

	t.Logf("%v: %v: got unexpected error '%v' (%T)", testName, name, err, err)
}

func testEncoderStat(t *testing.T, fs Filesystem, want statResult) error {
	t.Helper()

	name := want.name

	fi, err := fs.Stat(name)
	if err != nil {
		return _err(err)
	}

	// First, test everything but the modTime
	if err = testEncoderFileInfo(t, fs, want, fi); err != nil {
		return err
	}

	// Then set the modTime and test it
	now := time.Now()
	if err = fs.Chtimes(name, now, now); err != nil {
		return _err(err)
	}
	want.modTime = now

	if fi, err = fs.Stat(name); err != nil {
		return _err(err)
	}

	return testEncoderFileInfo(t, fs, want, fi)
}

func testEncoderFileInfo(t *testing.T, fs Filesystem, want statResult, fi FileInfo) error {
	t.Helper()

	if path.Base(filepath.ToSlash(fi.Name())) != path.Base(filepath.ToSlash(want.name)) {
		return _errf("Name(): got %v, want %v", fi.Name(), want.name)
	}
	if fi.IsRegular() != want.isRegular {
		return _errf("IsRegular(): got %v, want %v", fi.IsRegular(), want.isRegular)
	}
	if fi.IsDir() != want.isDir {
		return _errf("IsDir(): got %v, want %v", fi.IsDir(), want.isDir)
	}
	if fi.IsSymlink() != want.isSymlink {
		return _errf("IsSymlink(): got %v, want %v", fi.IsSymlink(), want.isSymlink)
	}
	if !fi.IsDir() && !fi.IsSymlink() {
		if fi.Size() != int64(want.size) {
			return _errf("Size(): got %v, want %v", fi.Size(), want.size)
		}
	}

	if want.modTime.IsZero() {
		return nil
	}

	diff := math.Abs(float64(fi.ModTime().Sub(want.modTime)))
	if diff > 100_000_000 { // .1 seconds
		return _errf("ModTime(): got %v, want %v, diff %f", fi.ModTime(), want.modTime, diff)
	}

	if isFAT || fs.Type() == FilesystemTypeFake {
		// The fakefs and FAT filesystems always returns 0 for uid/gids
		return nil
	}

	if fi.Owner() != os.Getuid() {
		return _errf("Owner(): got %v, want %v", fi.Owner(), os.Getuid())
	}
	if fi.Group() != os.Getgid() {
		return _errf("Group(): got %v, want %v", fi.Group(), os.Getgid())
	}

	return nil
}

func testEncoderFileStat(t *testing.T, fs Filesystem, want statResult) error {
	t.Helper()

	name := want.name

	fd, err := fs.Open(name)
	if err != nil {
		return _err(err)
	}

	fi, err := fd.Stat()
	if err != nil {
		return _err(err)
	}
	fd.Close()

	// First, test everything without calling Chtimes
	if err = testEncoderFileInfo(t, fs, want, fi); err != nil {
		return _err(err)
	}

	fd, err = fs.Open(name)
	if err != nil {
		return _err(err)
	}
	defer fd.Close()

	fi, err = fd.Stat()
	if err != nil {
		return _err(err)
	}

	now := time.Now()
	if err = fs.Chtimes(name, now, now); err != nil {
		return _err(err)
	}

	want.modTime = now
	return testEncoderFileInfo(t, fs, want, fi)
}

func testEncoderNameToNamePairs(t *testing.T, fs Filesystem, name string, dirs bool) (namePairs, error) {
	t.Helper()

	seeds := namePairs{
		{"0xreg-1", "0xreg-2"},     // two regular files
		{"0xreg-3", name + "-4"},   // a regular file and an encoded/decoded file
		{name + "-5", "0xreg-6"},   // an encoded/decoded file and a regular file
		{name + "-7", name + "-8"}, // two encoded/decoded files
	}

	var pairs namePairs
	pairs = append(pairs, seeds...)
	if dirs {
		// dir-x/file1
		// file2
		for i, pair := range seeds {
			pairs = append(pairs, namePair{
				fmt.Sprintf("0xdir-%d/%s", 1+i, pair[0]), // 1/2/3
				pair[1],
			})
		}

		// file1
		// dir-x/file2
		for i, pair := range seeds {
			pairs = append(pairs, namePair{
				pair[0],
				fmt.Sprintf("0xdir-%d/%s", 3+i, pair[1]), // 3/4/5
			})
		}
		// dir-x/file1
		// dir-y/file2
		for i, pair := range seeds {
			pairs = append(pairs, namePair{
				fmt.Sprintf("0xdir-%d/%s", 6+i*2, pair[0]), // 6/8/10
				fmt.Sprintf("0xdir-%d/%s", 7+i*2, pair[1]), // 7/9/11
			})
		}
		// dir-x/file1
		// dir-x/file2
		for i, pair := range seeds {
			pairs = append(pairs, namePair{
				fmt.Sprintf("0xdir-%d/%s", 12+i, pair[0]), // 12/13/14
				fmt.Sprintf("0xdir-%d/%s", 12+i, pair[1]), // 12/13/14
			})
		}
		for i := range pairs {
			pairs[i][0] += fmt.Sprintf("-%d", i*2+1)
			pairs[i][1] += fmt.Sprintf("-%d", i*2+2)
		}

		err := testEncoderEnsureDirs(fs, pairs)
		if err != nil {
			return nil, err
		}
	}

	return pairs, nil
}

func testEncoderEnsureDirs(fs Filesystem, pairs namePairs) error {
	for _, pair := range pairs {
		if err := testEncoderEnsureDir(fs, path.Dir(pair[0])); err != nil {
			return err
		}
		if err := testEncoderEnsureDir(fs, path.Dir(pair[1])); err != nil {
			return err
		}
	}

	return nil
}

func testEncoderEnsureDir(fs Filesystem, dir string) error {
	if dir == "" || dir == "." {
		return nil
	}

	fi, err := fs.Stat(dir)
	if err == nil && fi.IsDir() {
		return nil
	}
	err = fs.MkdirAll(dir, 0o775)
	if err != nil {
		return _err(err)
	}

	return nil
}

func testEncoderGetTestFilenames() ([]string, []string, []string) {
	var regularFiles []string
	var decodedFiles []string
	var encodedFiles []string

	var r rune
	for _, r = range testEncoderGetRegularRunes() {
		name := fmt.Sprintf("0x%04x-%c-regular", r, r)
		regularFiles = append(regularFiles, name)
		if testing.Short() {
			break
		}
	}

	for _, r = range testEncoderGetFatRunes() {
		// We need to exclude colons from our decode tests on Windows, because if
		// we try to create a file named `acolon:.txt`, Windows will create a file
		// named `acolon`, with an Alternate Data Stream named `.txt`.
		if build.IsWindows && r == ':' {
			continue
		}
		name := fmt.Sprintf("0x%04x-%c-decoded", r, r)
		decodedFiles = append(decodedFiles, name)
		if testing.Short() {
			break
		}
	}

	for _, r = range testEncoderGetFatRunes() {
		encoded := r | consts.BaseRune
		name := fmt.Sprintf("0x%04x-%c-encoded", encoded, encoded)
		encodedFiles = append(encodedFiles, name)
		if testing.Short() {
			break
		}
	}

	return regularFiles, decodedFiles, encodedFiles
}

func testEncoderGetRegularRunes() []rune {
	extraRunes := []rune{
		'\u07FF',
		'\u0800',
		'\uEFFF',
		'\uF100',
		unicode.ReplacementChar, // '\uFFFD' Represents invalid code points.
		'\uFFEF',
		'\uFFFF',
		'\U00010000',
		unicode.MaxRune - 1,
		unicode.MaxRune, // '\U0010FFFF',
	}

	maxRune := unicode.MaxLatin1 // \xff
	runes := make([]rune, 0, int(maxRune)+len(extraRunes))
	for r := rune(0); r <= maxRune; r++ {
		// Skip filenames with control characters and DEL (\x7f)
		if unicode.IsControl(r) {
			continue
		}
		// Skip encodable chars.
		if strings.ContainsRune(consts.Encodes, r) {
			continue
		}
		// Skip NUL (\x00), / (\x2f), \ (\x5c)
		if strings.ContainsRune(consts.Nevers, r) {
			continue
		}
		// We don't build on plan9, but others might:
		// https://github.com/syncthing/syncthing/blob/2794b042/.github/workflows/build-syncthing.yaml#L412
		if build.IsPlan9 {
			// Per https://9fans.github.io/plan9port/man/man9/intro.html:
			//  "Plan 9 names may contain any printable character (that is, any
			//  character outside hexadecimal 00-1F and 80-9F) except slash."
			if r >= '\x80' && r <= '\x9f' {
				continue
			}
		}
		runes = append(runes, r)
	}
	runes = append(runes, extraRunes...)

	return runes
}

func testEncoderGetFatRunes() []rune {
	// Do control characters last.
	input := consts.Encodes
	runes := utf8.RuneCountInString(input)
	output := make([]rune, runes)
	for _, r := range input {
		runes--
		output[runes] = r
	}
	return output
}

func testEncoderCreateFile(fs Filesystem, name string) error {
	dir := filepath.Dir(name)
	if dir != "" && dir != "." {
		err := fs.MkdirAll(dir, 0o755)
		if err != nil {
			return _err(err)
		}
	}
	fd, err := fs.Create(name)
	if err != nil {
		return _err(err)
	}
	defer fd.Close()
	written, err := fd.Write([]byte(name))
	if err != nil {
		return _err(err)
	}
	if written != len(name) {
		return _errf("write error: wrote %d bytes, want %d bytes", written, len(name))
	}

	return nil
}

func testEncoderDirsEqual(got []string, want []string, recursive bool) bool {
	names := make([]string, 0, len(want))
	for _, name := range got {
		if strings.Contains(name, ".stfolder") {
			continue
		}
		names = append(names, filepath.ToSlash(name))
	}
	slices.Sort(names)
	for _, name := range want {
		name = filepath.ToSlash(name)
		_, ok := slices.BinarySearch(names, name)
		if !ok {
			return false
		}
	}

	return true
}

func testEncoderCleanup(t *testing.T, fs Filesystem) {
	t.Helper()

	filenames, err := fs.DirNames("/")
	if err != nil {
		t.Fatalf("Cannot read directory %v", fs.URI())
	}
	for _, filename := range filenames {
		_ = fs.RemoveAll(filename)
	}
}

func testEncoderGetGlobFiles(fs Filesystem, name string, maxNames int, recursive bool) ([]string, error) {
	regularFiles, decodedFiles, encodedFiles := testEncoderGetTestFilenames()

	var names []string
	if fat.IsEncoded(name) {
		names = encodedFiles
	} else if fat.IsDecoded(name) {
		names = decodedFiles
	} else {
		names = regularFiles
	}

	if maxNames >= 0 && len(names) > maxNames {
		names = slices.Delete(names, maxNames, len(names))
	}

	for i := range names {
		if recursive {
			names[i] = filepath.ToSlash(filepath.Join(names[i], names[i]))
		}
		if err := testEncoderCreateFile(fs, names[i]); err != nil {
			return nil, err
		}
	}

	return names, nil
}

func testEncoderGetRootURI(t *testing.T, volumeType fsutil.VolumeType) (string, error) {
	t.Helper()

	return tempDirEnv(t, volumeEnvvar(volumeType))
}

// The following functions are also at the end of lib/fsutil/fsutil_test.go
func volumeEnvvar(volumeType fsutil.VolumeType) string {
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
		return "", fsutil.ErrNotADirectory
	}
	dir, err = os.MkdirTemp(dir, UnixTempPrefix+"*.tmp")
	if err != nil {
		return "", fsutil.ErrCannotCreateDirectory
	}

	return dir, nil
}

// Debugging functions

type testEncoderError struct {
	Err  error
	File string
	Line int
}

func (e *testEncoderError) Error() string {
	return fmt.Sprintf("%v (%s:%d)", e.Err.Error(), e.File, e.Line)
}

func (e *testEncoderError) Unwrap() error { return e.Err }

func _err(err error) error {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "n/a"
		line = 0
	}

	return &testEncoderError{
		Err:  err,
		File: path.Base(file),
		Line: line,
	}
}

func _errf(format string, a ...any) error {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "n/a"
		line = 0
	}

	return &testEncoderError{
		Err:  fmt.Errorf(format, a...),
		File: path.Base(file),
		Line: line,
	}
}

func testEncoderDumpDir(t *testing.T, fs Filesystem) {
	t.Helper()

	t.Logf("Dumping %q:", fs.URI())
	i := 1
	fs.Walk("/", func(path string, info FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)

			return err
		}
		s := filepath.ToSlash(path)
		if info.IsDir() {
			s += "/"
		}
		if info.IsSymlink() {
			s += " ->"
			target, err2 := fs.ReadSymlink(path)
			if err2 == nil {
				s += " " + target
			}
		}
		t.Logf("%d: %s", i, s)
		i++

		return nil
	})
}
