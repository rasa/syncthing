// Copyright (C) 2026 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package rclone_test

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"github.com/syncthing/syncthing/lib/encoder/rclone"
	"github.com/syncthing/syncthing/lib/encoder/rclone/consts"
)

const (
	// https://github.com/golang/text/blob/9c2f3a21/transform/transform.go#L130
	defaultBufSize = 4096
)

var (
	repl  = string(unicode.ReplacementChar) // \uFFFD
	repl2 = repl + repl
	repl3 = repl2 + repl
	repl4 = repl3 + repl
	repl5 = repl4 + repl
)

type decodeTest struct {
	in      string
	out     string
	decoded bool
}

var decodeTests = []decodeTest{
	{"", "", false},
	{"ab", "ab", false},
	{"abc", "abc", false},
	{".", ".", false},
	{"..", "..", false},
	{"\uff1a\uff1a", "::", true},
	{"[\uff1a", "[:", true},
	{"\uff1a[", ":[", true},
	{"/", "/", false},
	{`\`, `\`, false},
	{"\x00", "\x00", false},
	{"\uf001", "\uf001", false},
	{"c\uff1a", "c:", true},
	{"c\uff1a\\", `c:\`, true},
	{"\\\\c\uff1a\\", `\\c:\`, true},
	{repl, repl, false},
}

type encodeTest = struct {
	in      string
	out     string
	encoded bool
}

var encodeTests = []encodeTest{
	{"", "", false},
	{"", "", false},
	{"ab", "ab", false},
	{"abc", "abc", false},
	{".", ".", false},
	{"..", "..", false},
	{"::", "\uff1a\uff1a", true},
	{"[:", "[\uff1a", true},
	{":[", "\uff1a[", true},
	{"/", "/", false},
	{`\`, `\`, false},
	{"\x00", "\x00", false},
	{"\x01", "␁", true},

	// Unicode range
	{"\u0000", "\u0000", false},
	{"\u0001", "␁", true},
	{"\n", "␊", true},
	{"\r", "␍", true},
	{"\u001F", "␟", true},
	{"\u0020", " ", false},
	{"\u0025", "%", false},
	{"\u0026", "&", false},
	{"\u0027", "'", false},
	{"\u007E", "~", false},
	{"\u0080", "\u0080", false},
	{"\u00FF", "\u00FF", false},
	{"\u07FF", "\u07FF", false},
	{"\u0800", "\u0800", false},
	{"\uEFFF", "\uEFFF", false},
	{"\uF100", "\uF100", false},
	{"\uFFEF", "\uFFEF", false},
	{"\uFFFF", "\uFFFF", false},
	{"\U00010000", "\U00010000", false},
	{"\U0010FFFF", "\U0010FFFF", false},
	// Invalid UTF-8 (bad bytes are converted to U+FFFD)
	{"\xC0\x80", repl2, false},                   // U+0000
	{"\xF4\x90\x80\x80", repl4, false},           // U+110000
	{"\xF7\xBF\xBF\xBF", repl4, false},           // U+1FFFFF
	{"\xF8\x88\x80\x80\x80", repl5, false},       // U+200000
	{"\xF4\x8F\xBF\x3E", repl3 + "\uff1E", true}, // U+10FFFF (bad byte)
	{"\xF4\x8F\xBF", repl3, false},               // U+10FFFF (short)
	{"\xF4\x8F", repl2, false},
	{"\xF4", repl, false},
	{"\x00\xF4\x00", "\x00" + repl + "\x00", false},

	{"\xC0\x80:", repl2 + "\uff1a", true},             // U+0000
	{"\xF4\x90\x80\x80:", repl4 + "\uff1a", true},     // U+110000
	{"\xF7\xBF\xBF\xBF:", repl4 + "\uff1a", true},     // U+1FFFFF
	{"\xF8\x88\x80\x80\x80:", repl5 + "\uff1a", true}, // U+200000
	{"\xF4\x8F\xBF:", repl3 + "\uff1a", true},         // U+10FFFF (short)
	{"\xF4\x8F:", repl2 + "\uff1a", true},
	{"\xF4:", repl + "\uff1a", true},
	{"\x00\xF4\x00:", "\x00" + repl + "\x00" + "\uff1a", true},
}

var patternEncodeTests = []encodeTest{
	{"*", "*", false},
	{"?", "?", false},
	{`\*`, `\*`, false},
	{`\?`, `\?`, false},
	{"*:", "*\uff1a", true},
	{":*", "\uff1a*", true},
	{"*:*", "*\uff1a*", true},
	{":*:", "\uff1a*\uff1a", true},
	{"?:", "?\uff1a", true},
	{":?", "\uff1a?", true},
	{"?:?", "?\uff1a?", true},
	{":?:", "\uff1a?\uff1a", true},
}

func init() {
	// Android encodes '\x7F', but Windows doesn't
	if strings.ContainsRune(consts.Encodes, '\x7F') {
		encodeTests = append(encodeTests, encodeTest{"\u007F", "␡", true})
	} else {
		encodeTests = append(encodeTests, encodeTest{"\u007F", "␡", false})
	}
}

func TestRcloneDecoderDecodes(t *testing.T) {
	t.Parallel()
	testRcloneDecoder(t, decodeTests)
}

func TestRcloneDecoderLatin1Chars(t *testing.T) {
	t.Parallel()
	tests := make([]decodeTest, 0, consts.NumChars)
	for r := rune(0); r < rune(consts.NumChars); r++ {
		ins := string(r)
		out := ins
		decoded := strings.ContainsRune(consts.Encodes, r)
		test := decodeTest{
			in:      ins,
			out:     out,
			decoded: decoded,
		}
		tests = append(tests, test)
	}
	testRcloneDecoder(t, tests)
}

func TestRcloneDecoderRcloneChars(t *testing.T) {
	t.Parallel()
	tests := make([]decodeTest, 0, consts.NumChars)
	for r := rune(0); r < rune(consts.NumChars); r++ {
		// ins := string(r | consts.BaseRune)
		enc, ok := consts.EncodeMap[r]
		if ok {
			enc = r
		}
		ins := string(enc)
		out := ins
		decoded := strings.ContainsRune(consts.Encodes, r)
		if decoded {
			out = string(r)
		}
		test := decodeTest{
			in:      ins,
			out:     out,
			decoded: decoded,
		}
		tests = append(tests, test)
	}
	testRcloneDecoder(t, tests)
}

func TestRcloneDecoderEncodes(t *testing.T) {
	t.Parallel()
	tests := make([]decodeTest, 0, len(encodeTests))

	for _, test := range encodeTests {
		// Can't decode Unicode ReplacementChar (\ufffd) so skip the test.
		if hasReplacementChar(test.out) {
			continue
		}
		test := decodeTest{
			// Reversal is intentional
			in:      test.out,
			out:     test.in,
			decoded: test.encoded,
		}
		tests = append(tests, test)
	}
	testRcloneDecoder(t, tests)
}

func TestRcloneEncoderEncodes(t *testing.T) {
	t.Parallel()
	testRcloneEncoder(t, encodeTests)
}

func TestRcloneEncoderLatin1Chars(t *testing.T) {
	t.Parallel()
	tests := make([]encodeTest, 0, consts.NumChars)

	for r := rune(0); r < rune(consts.NumChars); r++ {
		ins := string(r)
		out := ins
		encoded := strings.ContainsRune(consts.Encodes, r)
		if encoded {
			// out = string(r | consts.BaseRune)
			enc, ok := consts.EncodeMap[r]
			if ok {
				out = string(enc)
			}
		}
		test := encodeTest{
			in:      ins,
			out:     out,
			encoded: encoded,
		}
		tests = append(tests, test)
	}
	testRcloneEncoder(t, tests)
}

func TestRcloneEncoderRcloneChars(t *testing.T) {
	t.Parallel()
	tests := make([]encodeTest, 0, consts.NumChars)

	for r := rune(0); r < rune(consts.NumChars); r++ {
		// ins := string(r | consts.BaseRune)
		var ins string
		enc, ok := consts.EncodeMap[r]
		if ok {
			ins = string(enc)
		} else {
			ins = string(r)
		}
		out := ins
		encoded := strings.ContainsRune(consts.Encodes, r)
		test := encodeTest{
			in:      ins,
			out:     out,
			encoded: encoded,
		}
		tests = append(tests, test)
	}
	testRcloneEncoder(t, tests)
}

func TestRcloneEncoderDecodes(t *testing.T) {
	t.Parallel()
	tests := make([]encodeTest, 0, len(decodeTests))

	for _, d := range decodeTests {
		test := encodeTest{
			// reversal intentional
			in:      d.out,
			out:     d.in,
			encoded: d.decoded,
		}
		tests = append(tests, test)
	}
	testRcloneEncoder(t, tests)
}

// Test the pattern encoder with the encodeTests tests.
func TestRclonePatternEncoder(t *testing.T) {
	t.Parallel()
	enc := rclone.RclonePattern.NewEncoder()

	if testing.Verbose() {
		t.Logf("Running %d tests", len(encodeTests)*len(getLengths()))
	}

	for i, test := range encodeTests {
		j := i + 1
		for _, length := range getLengths() {
			ins := strings.Repeat(test.in, length)
			want := strings.Repeat(test.out, length)
			got, err := enc.String(ins)
			if err != nil {
				t.Errorf("Test %d: RclonePattern.Encode(%+q) unexpected error; %v", j, ins, err)
			}
			if got != want {
				t.Errorf("Test %d: RclonePattern.Encode(%+q) got %+q; want %+q (%d vs %d bytes)", j, ins, got, want, len(want), len(want))
			}

			got, err = rclone.EncodePattern(ins)
			if err != nil {
				t.Errorf("Test %d: EncodePattern(%+q, true) unexpected error; %v", j, ins, err)
			}
			if got != want {
				t.Errorf("Test %d: EncodePattern(%+q, true) got %v; want %v (%d vs %d bytes)", j, ins, got, want, len(got), len(want))
			}

			got = rclone.MustEncodePattern(ins)
			if got != want {
				t.Errorf("Test %d: MustEncodePattern(%+q, true) got %v; want %v (%d vs %d bytes)", j, ins, got, want, len(got), len(want))
			}
		}
	}
}

// Test the pattern encoder with the patternEncodeTests tests.
func TestRclonePatternEncoderSpecific(t *testing.T) {
	t.Parallel()
	enc := rclone.RclonePattern.NewEncoder()

	if testing.Verbose() {
		t.Logf("Running %d tests", len(patternEncodeTests)*len(getLengths()))
	}

	for i, test := range patternEncodeTests {
		j := i + 1
		for _, length := range getLengths() {
			ins := strings.Repeat(test.in, length)
			out := strings.Repeat(test.out, length)
			got, err := enc.String(ins)
			if err != nil {
				t.Errorf("Test %d: FATPattern.Encode(%+q) unexpected error; %v", j, ins, err)
			}
			if got != out {
				t.Errorf("Test %d: FATPattern.Encode(%+q) got %+q; want %+q (%d vs %d bytes)", j, ins, got, out, len(got), len(out))
			}
		}
	}
}

func TestRcloneIsDecoded(t *testing.T) {
	t.Parallel()
	testRcloneDecoder(t, decodeTests)
}

func testRcloneDecoder(t *testing.T, tests []decodeTest) {
	t.Helper()

	if testing.Verbose() {
		t.Logf("Running %d tests", len(tests)*len(getLengths()))
	}

	dec := rclone.Rclone.NewDecoder()

	for i, test := range tests {
		j := i + 1
		for _, length := range getLengths() {
			ins := strings.Repeat(test.in, length)
			want := strings.Repeat(test.out, length)
			got, err := dec.String(ins)
			if err != nil {
				t.Errorf("Test %d: Rclone.Decode(%q) got %v, want %v", j, ins, err, nil)
			}
			if got != want {
				t.Errorf("Test %d: Rclone.Decode(%q) got %+q; want %+q (%d vs %d bytes)", j, ins, got, want, len(got), len(want))
			}

			got, err = rclone.Decode(ins)
			if err != nil {
				t.Errorf("Test %d: Decode(%+q) unexpected error; %v", j, ins, err)
			}
			if got != want {
				t.Errorf("Test %d: Decode(%+q) got %+q; want %+q (%d vs %d bytes)", j, ins, got, want, len(got), len(want))
			}

			got = rclone.MustDecode(ins)
			if got != want {
				t.Errorf("Test %d: MustDecode(%+q) got %+q; want %+q (%d vs %d bytes)", j, ins, got, want, len(got), len(want))
			}

			got2 := rclone.IsDecoded(want)
			want2 := test.decoded
			if got2 != want2 {
				t.Errorf("Test %d: IsDecoded(%q) got %v; want %v", j, ins, got2, want2)
			}
		}
	}
}

func testRcloneEncoder(t *testing.T, tests []encodeTest) {
	t.Helper()

	enc := rclone.Rclone.NewEncoder()

	if testing.Verbose() {
		t.Logf("Running %d tests", len(tests)*len(getLengths()))
	}

	for i, test := range tests {
		j := i + 1
		for _, length := range getLengths() {
			ins := strings.Repeat(test.in, length)
			want := strings.Repeat(test.out, length)
			got, err := enc.String(ins)
			if err != nil {
				t.Errorf("Test %d: Rclone.Encode(%+q) got %v, want %v", j, ins, err, nil)
			}
			if got != want {
				t.Errorf("Test %d: Rclone.Encode(%q) got %+q; want %+q (%d vs %d bytes)", j, ins, got, want, len(got), len(want))
			}

			got, err = rclone.Encode(ins)
			if err != nil {
				t.Errorf("Test %d: Encode(%+q) unexpected error; %v", j, ins, err)
			}
			if got != want {
				t.Errorf("Test %d: Encode(%+q) got %+q; want %+q (%d vs %d bytes)", j, ins, got, want, len(got), len(want))
			}

			got = rclone.MustEncode(ins)
			if got != want {
				t.Errorf("Test %d: MustEncode(%+q) got %+q; want %+q (%d vs %d bytes)", j, ins, got, want, len(got), len(want))
			}

			got2 := rclone.IsEncoded(want)
			want2 := test.encoded
			if got2 != want2 {
				t.Fatalf("Test %d: IsEncoded(%q) got %v; want %v", j, want, got2, want2)
			}
		}
	}
}

func getLengths() []int {
	lengths := make([]int, 0)
	maxLenRaw := os.Getenv("STTESTFATMAXLEN")
	if maxLenRaw != "" {
		maxLen, err := strconv.Atoi(maxLenRaw)
		if err == nil {
			for i := 1; i < maxLen; i++ {
				lengths = append(lengths, i)
			}

			return lengths
		}
	}
	// fibonacci sequence:
	// 1 2 3 5 8 13 21 34 55 89 144 233 377 610 987 1597 2584 4181 6765
	a, b := 1, 2
	for a < defaultBufSize*2 {
		lengths = append(lengths, a)
		a, b = b, a+b
	}
	lengths = append(lengths, defaultBufSize-1)
	lengths = append(lengths, defaultBufSize)
	lengths = append(lengths, defaultBufSize+1)

	return lengths
}

func hasReplacementChar(name string) bool {
	return strings.ContainsRune(name, unicode.ReplacementChar)
}

func BenchmarkFatDecoder(b *testing.B) {
	b.ReportAllocs()

	dec := rclone.Rclone.NewDecoder()
	for i := 0; i < b.N; i++ {
		for _, d := range decodeTests {
			_, _ = dec.String(d.in)
		}
		for _, d := range encodeTests {
			if !hasReplacementChar(d.out) {
				_, _ = dec.String(d.out)
			}
		}
	}
}

func BenchmarkFatEncoder(b *testing.B) {
	b.ReportAllocs()

	enc := rclone.Rclone.NewEncoder()
	for i := 0; i < b.N; i++ {
		for _, d := range decodeTests {
			_, _ = enc.String(d.out)
		}
		for _, d := range encodeTests {
			_, _ = enc.String(d.in)
		}
	}
}

func BenchmarkFatPatternEncoder(b *testing.B) {
	b.ReportAllocs()

	enc := rclone.RclonePattern.NewEncoder()
	for i := 0; i < b.N; i++ {
		for _, d := range decodeTests {
			_, _ = enc.String(d.out)
		}
		for _, d := range encodeTests {
			_, _ = enc.String(d.in)
		}
	}
}

/*
Benchmark results for various defaultDecodeTransformer functions in rclone.go:

// BenchmarkFatDecoder-4            1973989              6487 ns/op
//
var defaultDecodeTransformer = runes.Map(func(r rune) rune {
	if r < consts.BaseRune || r >= (consts.BaseRune+consts.NumChars) {
		return r
	}
	if puaEncodes[r&^consts.BaseRune] >= consts.BaseRune {
		return r &^ consts.BaseRune
	}

	return r
})

// BenchmarkFatDecoder-4            1861136              6589 ns/op
//
var defaultDecodeTransformer = runes.Map(func(r rune) rune {
	if r < consts.BaseRune || r >= (consts.BaseRune+consts.NumChars) {
		return r
	}
	if puaEncodes[r-consts.BaseRune] >= consts.BaseRune {
		return r &^ consts.BaseRune
	}
	return r
})

// BenchmarkFatDecoder-4            1648190              6779 ns/op
//
var defaultDecodeTransformer = runes.Map(func(r rune) rune {
	if r < consts.BaseRune || r >= (consts.BaseRune+consts.NumChars) {
		return r
	}
	if puaEncodes[r-consts.BaseRune] >= consts.BaseRune {
		return r - consts.BaseRune
	}
	return r
})

// BenchmarkFatDecoder-4            1665566              7235 ns/op
//
var decodeMap map[rune]rune
func init() {
	decodeMap = make(map[rune]rune, len(puaEncodes))
	for _, r := range defaultDecodes {
		decodeMap[r] = r &^ consts.BaseRune
	}
}
var defaultDecodeTransformer = runes.Map(func(r rune) rune {
	r, ok := decodeMap[r]
	if ok {
		return r
	}
	return r
})

// BenchmarkFatDecoder-4            1371726              8911 ns/op
//
var defaultDecodeTransformer = runes.Map(func(r rune) rune {
	if strings.ContainsRune(puaEncodes, r) {
		return r &^ consts.BaseRune
	}
})

*/
