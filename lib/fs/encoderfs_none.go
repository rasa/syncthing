// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fs

// The "None" encoder does not encode, and it's only instantiated in our
// test suite. They are not used in non-test code.
type noneEncoderFS struct {
	encoderFS
}

type OptionNoneEncoder struct{}

func (*OptionNoneEncoder) apply(fs Filesystem) Filesystem {
	// Only used in test suite, as we don't instantiate None encoders otherwise.
	ffs := new(noneEncoderFS)
	ffs.Filesystem = fs
	ffs.Encoder = ffs
	ffs.encoderType = EncoderTypeNone
	ffs.SetRooter(ffs)
	return ffs
}

func (*OptionNoneEncoder) String() string {
	return "noneEncoder"
}

func (f *noneEncoderFS) decode(name string) string {
	return name
}

func (f *noneEncoderFS) encode(name string, pattern bool) (string, error) {
	return name, nil
}
