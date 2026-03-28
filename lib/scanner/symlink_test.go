// Copyright (C) 2026 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package scanner

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSymlinkTargetRelativeToRoot(t *testing.T) {
	tests := []struct {
		name    string
		root    string
		target  string
		want    string
		wantErr bool
	}{
		{
			name:   "relative target parent ref is preserved",
			root:   "/",
			target: "../dir",
			want:   "../dir",
		},
		{
			name:   "relative target child is preserved",
			root:   "/",
			target: "child",
			want:   "child",
		},
		{
			name:   "relative nested target is preserved",
			root:   "/",
			target: "child/grandchild",
			want:   "child/grandchild",
		},
		{
			name:   "relative target with dot is preserved",
			root:   "/",
			target: "./child",
			want:   "./child",
		},
		{
			name:   "relative target with many dotdots is preserved",
			root:   "/",
			target: "../../../x/y",
			want:   "../../../x/y",
		},
		{
			name:   "empty relative target stays empty",
			root:   "/",
			target: "",
			want:   "",
		},
		{
			name:   "single dot target stays dot",
			root:   "/",
			target: ".",
			want:   ".",
		},
		{
			name:   "relative target backslashes become slashes on windows input",
			root:   "/",
			target: "a/b/c",
			want:   "a/b/c",
		},
		{
			name:   "absolute target equal to root becomes dot",
			root:   "/",
			target: "/",
			want:   ".",
		},
		{
			name:   "absolute target directly under root becomes relative",
			root:   "/",
			target: "/x",
			want:   "x",
		},
		{
			name:   "absolute nested target under root becomes relative",
			root:   "/",
			target: "/x/y/z",
			want:   "x/y/z",
		},
		{
			name:   "absolute target with trailing separator is cleaned",
			root:   "/",
			target: "/x/y/",
			want:   "x/y",
		},
		{
			name:   "absolute target with dot segments is cleaned",
			root:   "/",
			target: "/a/./c",
			want:   "a/c",
		},
		{
			name:   "absolute target with dotdot segments is cleaned",
			root:   "/",
			target: "/a/b/../c",
			want:   "a/c",
		},
		{
			name:   "absolute target under nested root becomes relative",
			root:   "/a",
			target: "/a/b/c",
			want:   "b/c",
		},
		{
			name:   "absolute target equal to nested root becomes dot",
			root:   "/a/b",
			target: "/a/b",
			want:   ".",
		},
		{
			name:   "root with dot segment is cleaned",
			root:   "/.",
			target: "/x/y",
			want:   "x/y",
		},
		{
			name:   "nested root with trailing separator is cleaned",
			root:   "/a/b/",
			target: "/a/b/c/d",
			want:   "c/d",
		},
		{
			name:    "empty root is error",
			root:    "",
			target:  "/x/y",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := hostifyPath(tt.root)
			target := hostifyPath(tt.target)

			got, err := SymlinkTargetRelativeToRoot(root, target)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got result %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSymlinkTargetRelativeToRoot_WindowsOutsideRoot(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}

	tests := []struct {
		name   string
		root   string
		target string
	}{
		{
			name:   "same drive but outside root",
			root:   `C:\root`,
			target: `C:\other\x`,
		},
		{
			name:   "different drive",
			root:   `C:\root`,
			target: `D:\other\x`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SymlinkTargetRelativeToRoot(tt.root, tt.target)
			if err == nil {
				t.Fatalf("expected error, got result %q", got)
			}
		})
	}
}

func platformRoot() string {
	if runtime.GOOS == "windows" {
		return `C:\root`
	}
	return `/`
}

// Define tests in POSIX form first.
// On Windows:
//
//	/      -> C:\root
//	/a/b   -> C:\root\a\b
//	a/b    -> a\b
func hostifyPath(s string) string {
	if s == "" {
		return ""
	}

	if runtime.GOOS == "windows" {
		if s == "/" {
			return platformRoot()
		}
		if strings.HasPrefix(s, "/") {
			rest := strings.TrimPrefix(s, "/")
			if rest == "" {
				return platformRoot()
			}
			return filepath.Join(platformRoot(), filepath.FromSlash(rest))
		}
		return filepath.FromSlash(s)
	}

	return s
}
