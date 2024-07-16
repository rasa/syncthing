// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

// build this file on all platforms, as a user may mount a disk that he intends
// to hook up to a Windows systems in the future:

package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding"

	"github.com/syncthing/syncthing/lib/encoding/fat"
)

var _ = fmt.Printf

// The "Windows" encoder encodes characters reserved on vFAT/exFAT/NTFS/reFS
// filesystems. It also encodes filenames ending with a space or period, and
// Windows' reserved filenames, such as `NUL` or `CON.txt`.
type windowsEncoderFS struct {
	encoderFS
	decoder        *encoding.Decoder
	encoder        *encoding.Encoder
	patternEncoder *encoding.Encoder
}

type OptionWindowsEncoder struct{}

func (*OptionWindowsEncoder) apply(fs Filesystem) Filesystem {
	wfs := &windowsEncoderFS{
		encoderFS: encoderFS{
			Filesystem:  fs,
			encoderType: EncoderTypeWindows,
		},
		decoder:        fat.PUA.NewDecoder(),
		encoder:        fat.PUA.NewEncoder(),
		patternEncoder: fat.PUAPattern.NewEncoder(),
	}
	wfs.Encoder = wfs
	wfs.SetRooter(wfs)
	return wfs
}

func (*OptionWindowsEncoder) String() string {
	return "windowsEncoder"
}

// decode returns the original pre-encoded filename, if the filename is encoded.
func (f *windowsEncoderFS) decode(name string) string {
	if !fat.IsEncoded(name) {
		return name
	}
	decoded, err := f.decoder.String(name)
	if err != nil {
		panic("bug: windows.decode: " + err.Error())
	}
	if decoded != name && l.ShouldDebug("encoder") {
		l.Debugf("Windows encoder decoded %q as %q", name, decoded)
	}
	return decoded
}

const windowsReservedFilenamesRegex = "(AUX|CON|NUL|PRN|COM[1-9\u00b2\u00b3\u00b9]|LPT[1-9\u00b2\u00b3\u00b9]|CONIN\\$|CONOUT\\$)"

// var windowsReservedFilenamesRegex1 = regexp.MustCompile(
//
//	`(?i)(^|` + string(os.PathSeparator) + `)` +
//	windowsReservedFilenamesRegex +
//	`(` + string(os.PathSeparator) + `|\.[^.]*$|$)`)
var windowsReservedFilenamesRegex2 = regexp.MustCompile(`(?i)^` +
	windowsReservedFilenamesRegex + `(\.[^.]*$|$)`)

func encodeLastChar(name string) string {
	runes := []rune(name)
	runes[len(runes)-1] |= fat.BaseRune
	return string(runes)
}

// encode returns the encoded filename, if the filename needs encoding.
func (f *windowsEncoderFS) encode(name string, pattern bool) (string, error) {
	if fat.IsEncoded(name) {
		// The Windows encoder rejects encoded filenames, regardless of the
		// underlying filesystem.
		l.Warnf("Windows encoder ignoring encoded filename %q", name)
		return "", &os.PathError{Op: "encode", Path: name, Err: os.ErrNotExist}
	}

	encoded := name

	// https://go.dev/play/p/S9KtG8FjGjH
	if !f.pattern {
		separators := strings.Count(encoded, pathSeparatorString)
		encoded = filepath.ToSlash(encoded)
		parts := strings.Split(encoded, "/")
		for i, part := range parts {
			if part == "" {
				continue
			}
			if part == "." || part == ".." {
				continue
			}
			ext := filepath.Ext(part)
			if ext != "." {
				base := strings.TrimSuffix(part, ext)
				base = windowsReservedFilenamesRegex2.ReplaceAllStringFunc(base, encodeLastChar)
				part = base + ext
			}
			r, _ := utf8.DecodeLastRuneInString(part)
			if r == ' ' || r == '.' {
				part = encodeLastChar(part)
			}
			parts[i] = part
		}
		encoded = strings.Join(parts, "/")
		if separators > 0 {
			encoded = filepath.FromSlash(encoded)
		}
	}

	if fat.IsDecoded(encoded) {
		var err error
		if f.pattern {
			encoded, err = f.patternEncoder.String(encoded)
		} else {
			encoded, err = f.encoder.String(encoded)
		}
		// The encoder has never failed in testing, but since we can return an error,
		// we might as well.
		if err != nil {
			return "", err
		}
	}
	if encoded != name && l.ShouldDebug("encoder") {
		l.Debugf("Windows encoder encoded %q as %q", name, encoded)
	}

	return encoded, nil
}
