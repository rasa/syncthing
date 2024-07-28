// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package integration

type srcType int

const (
	// Generate pre-encoded filenames on the src encoder.
	srcTypeDecoded srcType = iota
	// Generate encoded filenames on the src encoder.
	srcTypeEncoded
)

type dstType int

const (
	// The dst encoder will save pre-encoded filenames.
	dstTypeDecoded dstType = iota
	// The dst encoder will save encoded filenames.
	dstTypeEncoded
	// The dst encoder will save encoded filenames, but reject encode filenames
	// on the wire.
	dstTypeRejectEncoded
	// dstTypeSkipped indicates a skipped test as FAT filesystems cannot save
	// pre-encoded filenames.
	dstTypeSkipped
)
