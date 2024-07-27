// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

import (
	"os"

	"golang.org/x/text/encoding"

	"github.com/syncthing/syncthing/lib/encoding/fat"
)

// The "FAT" encoder encodes characters reserved on vFAT/exFAT/NTFS/reFS
// filesystems. It does not encode filenames ending with a space or period,
// which are accepted on Android, but rejected on Windows. It also does not
// encode Windows' reserved filenames, such as `NUL` or `CON.txt`.
// These can be addressed later, with a "Windows" encoder, if desired.
type fatEncoderFS struct {
	encoderFS
	decoder        *encoding.Decoder
	encoder        *encoding.Encoder
	patternEncoder *encoding.Encoder
}

type OptionFatEncoder struct{}

func (*OptionFatEncoder) apply(fs Filesystem) Filesystem {
	ffs := new(fatEncoderFS)
	ffs.Filesystem = fs
	ffs.Encoder = ffs
	ffs.encoderType = EncoderTypeFat
	ffs.decoder = fat.PUA.NewDecoder()
	ffs.encoder = fat.PUA.NewEncoder()
	ffs.patternEncoder = fat.PUAPattern.NewEncoder()
	ffs.SetRooter(ffs)
	return ffs
}

func (*OptionFatEncoder) String() string {
	return "fatEncoder"
}

// decode returns the original pre-encoded filename, if the filename is encoded.
func (f *fatEncoderFS) decode(name string) string {
	if !fat.IsEncoded(name) {
		return name
	}
	decoded, err := f.decoder.String(name)
	if err != nil {
		panic("bug: fat.decode: " + err.Error())
	}
	if decoded != name && l.ShouldDebug("encoder") {
		l.Debugf("FAT encoder decoded %q as %q", name, decoded)
	}
	return decoded
}

// encode returns the encoded filename, if the filename needs encoding.
func (f *fatEncoderFS) encode(name string, pattern bool) (string, error) {
	if fat.IsEncoded(name) {
		// The FAT encoder rejects encoded filenames, regardless of the
		// underlying filesystem.
		l.Warnf("FAT encoder ignoring encoded filename %q", name)
		return "", &os.PathError{Op: "encode", Path: name, Err: os.ErrNotExist}
	}
	if !fat.IsDecoded(name) {
		return name, nil
	}
	var encoded string
	var err error
	if f.pattern {
		encoded, err = f.patternEncoder.String(name)
	} else {
		encoded, err = f.encoder.String(name)
	}
	// The encoder has never failed in testing, but since we can return an error,
	// we might as well.
	if err != nil {
		return "", err
	}
	if encoded != name && l.ShouldDebug("encoder") {
		l.Debugf("FAT encoder encoded %q as %q", name, encoded)
	}
	return encoded, nil
}
