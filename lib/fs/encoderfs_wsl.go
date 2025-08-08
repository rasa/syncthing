// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

import (
	"log/slog"
	"os"

	"github.com/syncthing/syncthing/internal/slogutil"
	"github.com/syncthing/syncthing/lib/encoder/wsl"
	"golang.org/x/text/encoding"
)

// The "WSL" encoder encodes characters reserved on vFAT/exFAT/NTFS/reFS
// filesystems. It does not encode filenames ending with a space or period,
// which are accepted on Android, but rejected on Windows. It also does not
// encode Windows' reserved filenames, such as `NUL` or `CON.txt`.
// These reserved filenames are discussed in
// https://github.com/syncthing/syncthing/issues/9623
// and a proposed solution using the config setting reservedFilenames is in
// https://github.com/rasa/syncthing/tree/feature/9623-allow-reserved .
// We could also implement this as a "Windows" encoder, if desired.
type wslEncoderFS struct {
	encoderFS
	decoder        *encoding.Decoder
	encoder        *encoding.Encoder
	patternEncoder *encoding.Encoder
}

type OptionWSLEncoder struct{}

func (*OptionWSLEncoder) apply(fs Filesystem) Filesystem {
	ffs := new(wslEncoderFS)
	ffs.Filesystem = fs
	ffs.Encoder = ffs
	ffs.encoderType = EncoderTypeWSL
	ffs.decoder = wsl.WSL.NewDecoder()
	ffs.encoder = wsl.WSL.NewEncoder()
	ffs.patternEncoder = wsl.WSLPattern.NewEncoder()
	ffs.SetRooter(ffs)
	return ffs
}

func (*OptionWSLEncoder) String() string {
	return "wslEncoder"
}

// decode returns the original pre-encoded filename, if the filename is encoded.
func (f *wslEncoderFS) decode(name string) string {
	if !wsl.IsEncoded(name) {
		return name
	}
	decoded, err := f.decoder.String(name)
	if err != nil {
		panic("bug: wsl.decode: " + err.Error())
	}
	if decoded != name && debugEncoder {
		slog.Debug("WSL encoder: decoded", slogutil.FilePath(name), slog.Any("result", decoded))
	}
	return decoded
}

// encode returns the encoded filename, if the filename needs encoding.
func (f *wslEncoderFS) encode(name string, pattern bool) (string, error) {
	if wsl.IsEncoded(name) {
		// The WSL encoder rejects encoded filenames, regardless of the
		// underlying filesystem.
		slog.Warn("WSL encoder: ignoring encoded filename", slogutil.FilePath(name))
		return "", &os.PathError{Op: "encode", Path: name, Err: os.ErrNotExist}
	}
	if !wsl.IsDecoded(name) {
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
	if encoded != name && debugEncoder {
		slog.Debug("WSL encoder: encoded", slogutil.FilePath(name), slog.Any("result", encoded))
	}
	return encoded, nil
}
