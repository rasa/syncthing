// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/syncthing/syncthing/lib/build"
	"github.com/syncthing/syncthing/lib/protocol"
)

// Encoder represents the encoding/decoding functions of an encoderFS.
type Encoder interface {
	decode(name string) string
	encode(name string, pattern bool) (string, error)
}

// encoderFS encodes filenames containing reserved characters so they can be
// saved to disk.
type encoderFS struct {
	Filesystem
	Encoder
	Rooter
	encoderType EncoderType
	pattern     bool // true to not encode * and ? in glob patterns
}

var debugEncoder bool

func init() {
	debugEncoder = strings.Contains(os.Getenv("STTRACE"), "encoder")
}

// EncoderTypeOption returns the Option for the passed encoder type.
func EncoderTypeOption(encoderType EncoderType) Option {
	switch encoderType {
	case EncoderTypeUnset, EncoderTypeNone:
		// This is only used by the test suite code, we don't instantiate None
		// encoders otherwise.
		return new(OptionNoneEncoder)
	case EncoderTypeFat:
		return new(OptionFatEncoder)
	default:
		panic("bug: unknown encoder " + encoderType.String())
	}
}

func (f *encoderFS) Chmod(name string, mode FileMode) error {
	return f.Filesystem.Chmod(name, mode)
}

func (f *encoderFS) Lchown(name string, uid, gid string) error {
	return f.Filesystem.Lchown(name, uid, gid)
}

func (f *encoderFS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return f.Filesystem.Chtimes(name, atime, mtime)
}

func (f *encoderFS) Create(name string) (File, error) {
	fd, err := f.Filesystem.Create(name)
	if err != nil {
		return nil, err
	}
	bfd, ok := fd.(basicFile)
	if ok {
		return encoderFile{bfd, f, name}, nil
	}
	ffd, ok := fd.(*fakeFile)
	if ok {
		// fakeFS emulates the encode/decode logic internally (as it also does
		// with the caseFS's logic) so we don't need to return encoderFile{ffd}.
		return ffd, nil
	}
	// Since we've just checked for the only two filesystems we currently have,
	// if we get here, it must mean we're adding a third filesystem, so perhaps
	// we need to panic, to let the developer know to add the filesystem here.
	return nil, fmt.Errorf("bug: expected a basicFile, found a %T (%v)", fd, name)
}

func (f *encoderFS) CreateSymlink(target, name string) error {
	// BasicFilesystem's CreateSymlink encodes the name, but not the target
	// (as it's already relative), so we need to encode it.
	encodedTarget, err := f.Encoder.encode(target, false)
	if err != nil {
		return err
	}
	return f.Filesystem.CreateSymlink(encodedTarget, name)
}

func (f *encoderFS) DirNames(name string) ([]string, error) {
	names, err := f.Filesystem.DirNames(name)
	if err != nil {
		return nil, err
	}
	decodes := make([]string, len(names))

	for i := range names {
		decodes[i] = f.Encoder.decode(names[i])
	}

	return decodes, err
}

func (f *encoderFS) Lstat(name string) (FileInfo, error) {
	fi, err := f.Filesystem.Lstat(name)
	if err != nil {
		return nil, err
	}
	bfi, ok := fi.(basicFileInfo)
	if ok {
		return encoderFileInfo{
			basicFileInfo: bfi,
			name:          filepath.Base(name),
		}, nil
	}
	ffi, ok := fi.(*fakeFileInfo)
	if ok {
		return ffi, nil
	}
	return nil, fmt.Errorf("bug: expected a basicFileInfo, found a %T (%v)", fi, fi.Name())
}

func (f *encoderFS) Mkdir(name string, perm FileMode) error {
	return f.Filesystem.Mkdir(name, perm)
}

func (f *encoderFS) MkdirAll(name string, perm FileMode) error {
	return f.Filesystem.MkdirAll(name, perm)
}

func (f *encoderFS) Open(name string) (File, error) {
	fd, err := f.Filesystem.Open(name)
	if err != nil {
		return nil, err
	}
	bfd, ok := fd.(basicFile)
	if ok {
		return encoderFile{bfd, f, name}, nil
	}
	ffd, ok := fd.(*fakeFile)
	if ok {
		return ffd, nil
	}
	return nil, fmt.Errorf("expected a basicFile, found a %T (%v)", fd, name)
}

func (f *encoderFS) OpenFile(name string, flags int, mode FileMode) (File, error) {
	fd, err := f.Filesystem.OpenFile(name, flags, mode)
	if err != nil {
		return nil, err
	}
	bfd, ok := fd.(basicFile)
	if ok {
		return encoderFile{bfd, f, name}, nil
	}
	ffd, ok := fd.(*fakeFile)
	if ok {
		return ffd, nil
	}
	return nil, fmt.Errorf("expected a basicFile, found a %T (%v)", fd, name)
}

func (f *encoderFS) ReadSymlink(name string) (string, error) {
	link, err := f.Filesystem.ReadSymlink(name)
	if err != nil {
		return "", err
	}
	return f.Encoder.decode(link), nil
}

func (f *encoderFS) Remove(name string) error {
	return f.Filesystem.Remove(name)
}

func (f *encoderFS) RemoveAll(name string) error {
	return f.Filesystem.RemoveAll(name)
}

func (f *encoderFS) Rename(old, new string) error {
	return f.Filesystem.Rename(old, new)
}

func (f *encoderFS) Stat(name string) (FileInfo, error) {
	fi, err := f.Filesystem.Stat(name)
	if err != nil {
		return nil, err
	}
	bfi, ok := fi.(basicFileInfo)
	if ok {
		return encoderFileInfo{
			basicFileInfo: bfi,
			name:          filepath.Base(name),
		}, nil
	}
	ffi, ok := fi.(*fakeFileInfo)
	if ok {
		return ffi, nil
	}
	return nil, fmt.Errorf("expected a basicFileInfo, found a %T (%v)", fi, fi.Name())
}

func (f *encoderFS) SymlinksSupported() bool {
	return f.Filesystem.SymlinksSupported()
}

func (f *encoderFS) Walk(path string, walkFunc WalkFunc) error {
	return f.Filesystem.Walk(path, walkFunc)
}

func (f *encoderFS) Watch(name string, ignore Matcher, ctx context.Context, ignorePerms bool) (<-chan Event, <-chan error, error) {
	return f.Filesystem.Watch(name, ignore, ctx, ignorePerms)
}

func (f *encoderFS) Hide(name string) error {
	return f.Filesystem.Hide(name)
}

func (f *encoderFS) Unhide(name string) error {
	return f.Filesystem.Unhide(name)
}

func (f *encoderFS) Glob(pattern string) ([]string, error) {
	f.pattern = true
	files, err := f.Filesystem.Glob(pattern)
	f.pattern = false
	return files, err
}

func (f *encoderFS) Roots() ([]string, error) {
	return f.Filesystem.Roots()
}

func (f *encoderFS) Usage(name string) (Usage, error) {
	return f.Filesystem.Usage(name)
}

func (f *encoderFS) Type() FilesystemType {
	return f.Filesystem.Type()
}

func (f *encoderFS) URI() string {
	return f.Filesystem.URI()
}

func (f *encoderFS) Options() []Option {
	return f.Filesystem.Options()
}

func (f *encoderFS) SameFile(fi1, fi2 FileInfo) bool {
	return f.Filesystem.SameFile(fi1, fi2)
	// Instead of modifying BasicFilesystem's SameFile(), I would have
	// preferred to have used:
	//		ef1, ok1 := fi1.(encoderFileInfo)
	//		ef2, ok2 := fi2.(encoderFileInfo)
	//		if ok1 && ok2 {
	//			return f.Filesystem.SameFile(ef1.osFileInfo(), ef2.osFileInfo())
	//		}
	// here, but got the error:
	//		cannot use ef1.osFileInfo() (value of type "io/fs".FileInfo) as
	// 		FileInfo value in argument to f.Filesystem.SameFile: "io/fs".FileInfo
	//		 does not implement FileInfo (missing method Group)
	// I tried many different ways to solve this, and eventually got an error
	// that fs.FileInfo's Mode() is different that os.FileInfo's Mode(), which
	// seemed insurmountable, so I reverted to the solution above.
}

func (f *encoderFS) PlatformData(name string, withOwnership, withXattrs bool, xattrFilter XattrFilter) (protocol.PlatformData, error) {
	return f.Filesystem.PlatformData(name, withOwnership, withXattrs, xattrFilter)
}

func (f *encoderFS) GetXattr(name string, xattrFilter XattrFilter) ([]protocol.Xattr, error) {
	return f.Filesystem.GetXattr(name, xattrFilter)
}

func (f *encoderFS) SetXattr(name string, xattrs []protocol.Xattr, xattrFilter XattrFilter) error {
	return f.Filesystem.SetXattr(name, xattrs, xattrFilter)
}

func (f *encoderFS) underlying() (Filesystem, bool) {
	return f.Filesystem, true
}

func (f *encoderFS) EncoderType() EncoderType {
	return f.encoderType
}

func (f *encoderFS) SetRooter(rooter Rooter) {
	rfs, ok := f.Filesystem.(Rooter)
	if ok {
		rfs.SetRooter(rooter)
		return
	}
	// The only time the above check will fail is if we're in the process of
	// adding a third filesystem (after basic and fake) and it doesn't (yet)
	// implement Rooter.
	panic(fmt.Sprintf("bug: cannot cast a %T filesystem to Rooter", f.Filesystem))
}

func (f *encoderFS) rooted(rel string) (string, error) {
	encodedName, err := f.encode(rel, f.pattern)
	if err != nil {
		return "", err
	}
	rfs, ok := f.Filesystem.(Rooter)
	if ok {
		rv, err := rfs.rooted(encodedName)
		return rv, err
	}
	msg := fmt.Sprintf("bug: cannot cast a %T filesystem to Rooter", f.Filesystem)
	slog.Error(msg)
	panic(msg)
}

func (f *encoderFS) unrooted(path string) string {
	rfs, ok := f.Filesystem.(Rooter)
	if ok {
		unrooted := rfs.unrooted(path)
		rv := f.decode(unrooted)
		return rv
	}
	msg := fmt.Sprintf("bug: cannot cast a %T filesystem to Rooter", f.Filesystem)
	slog.Error(msg)
	panic(msg)
}

// The encoderFile is essentially an os.File that lies about its Name().
type encoderFile struct {
	// basicFile contains the encoded name that's on disk.
	basicFile
	fs *encoderFS
	// name is the original, pre-encoded name.
	name string
}

func (f encoderFile) Name() string {
	return f.name
}

func (f encoderFile) Stat() (FileInfo, error) {
	fi, err := f.basicFile.Stat()
	if err != nil {
		return nil, err
	}
	name := f.fs.Encoder.decode(fi.Name())
	bfi, ok := fi.(basicFileInfo)
	if ok {
		return encoderFileInfo{
			basicFileInfo: bfi,
			name:          filepath.Base(name),
		}, err
	}

	return nil, fmt.Errorf("expected a basicFileInfo, found a %T (%v)", fi, fi.Name())
}

// Used by copyRange to unwrap to the real file and access SyscallConn
func (f encoderFile) unwrap() File {
	return f.basicFile
}

// The encoderFileInfo is an os.FileInfo that lies about the Name().
type encoderFileInfo struct {
	basicFileInfo
	name string
}

// Name returns the original, pre-encoded name.
func (fi encoderFileInfo) Name() string {
	return fi.name
}

func (fi encoderFileInfo) InodeChangeTime() time.Time {
	return fi.basicFileInfo.InodeChangeTime()
}

func (fi encoderFileInfo) Mode() FileMode {
	return fi.basicFileInfo.Mode()
}

func (fi encoderFileInfo) Owner() int {
	return fi.basicFileInfo.Owner()
}

func (fi encoderFileInfo) Group() int {
	return fi.basicFileInfo.Group()
}

func (fi *encoderFileInfo) osFileInfo() os.FileInfo {
	return fi.basicFileInfo.osFileInfo()
}

// DefaultEncoderType return the default encoder type for the OS. On Windows
// and Android we default to the FAT encoder per @calmh's comment at
// https://github.com/syncthing/syncthing/issues/9539#issuecomment-2141394377
// On all other systems, including WSL on Windows, we default to the None encoder.
func DefaultEncoderType() EncoderType {
	if build.IsWindows || build.IsAndroid {
		return EncoderTypeFat
	}
	return EncoderTypeNone
}
