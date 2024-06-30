// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package fat_test

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"github.com/syncthing/syncthing/lib/encoding/fat"
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
	{"\uf03a\uf03a", "::", true},
	{"[\uf03a", "[:", true},
	{"\uf03a[", ":[", true},
	{"/", "/", false},
	{`\`, `\`, false},
	{"\x00", "\x00", false},
	{"\uf001", "\x01", true},
	{"c\uf03a", "c:", true},
	{"c\uf03a\\", `c:\`, true},
	{"\\\\c\uf03a\\", `\\c:\`, true},
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
	{"::", "\uf03a\uf03a", true},
	{"[:", "[\uf03a", true},
	{":[", "\uf03a[", true},
	{"/", "/", false},
	{`\`, `\`, false},
	{"\x00", "\x00", false},
	{"\x01", "\uf001", true},

	// Unicode range
	{"\u0000", "\u0000", false},
	{"\u0001", "\uF001", true},
	{"\n", "\uf00a", true},
	{"\r", "\uf00d", true},
	{"\u001F", "\uF01F", true},
	{"\u0020", " ", false},
	{"\u0025", "%", false},
	{"\u0026", "&", false},
	{"\u0027", "'", false},
	{"\u007E", "~", false},
	{"\u007F", "\u007F", false},
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
	{"\xF4\x8F\xBF\x3E", repl3 + "\uf03E", true}, // U+10FFFF (bad byte)
	{"\xF4\x8F\xBF", repl3, false},               // U+10FFFF (short)
	{"\xF4\x8F", repl2, false},
	{"\xF4", repl, false},
	{"\x00\xF4\x00", "\x00" + repl + "\x00", false},

	{"\xC0\x80:", repl2 + "\uf03a", true},             // U+0000
	{"\xF4\x90\x80\x80:", repl4 + "\uf03a", true},     // U+110000
	{"\xF7\xBF\xBF\xBF:", repl4 + "\uf03a", true},     // U+1FFFFF
	{"\xF8\x88\x80\x80\x80:", repl5 + "\uf03a", true}, // U+200000
	{"\xF4\x8F\xBF:", repl3 + "\uf03a", true},         // U+10FFFF (short)
	{"\xF4\x8F:", repl2 + "\uf03a", true},
	{"\xF4:", repl + "\uf03a", true},
	{"\x00\xF4\x00:", "\x00" + repl + "\x00" + "\uf03a", true},
}

var patternEncodeTests = []encodeTest{
	{"*", "*", false},
	{"?", "?", false},
	{`\*`, `\*`, false},
	{`\?`, `\?`, false},
	{"*:", "*\uf03a", true},
	{":*", "\uf03a*", true},
	{"*:*", "*\uf03a*", true},
	{":*:", "\uf03a*\uf03a", true},
	{"?:", "?\uf03a", true},
	{":?", "\uf03a?", true},
	{"?:?", "?\uf03a?", true},
	{":?:", "\uf03a?\uf03a", true},
}

func TestFATDecoderDecodes(t *testing.T) {
	t.Parallel()
	testFATDecoder(t, decodeTests)
}

func TestFATDecoderLatin1Chars(t *testing.T) {
	t.Parallel()
	tests := make([]decodeTest, 0, fat.NumChars)
	for r := rune(0); r < rune(fat.NumChars); r++ {
		ins := string(r)
		out := ins
		decoded := strings.ContainsRune(fat.Encodes, r)
		test := decodeTest{
			in:      ins,
			out:     out,
			decoded: decoded,
		}
		tests = append(tests, test)
	}
	testFATDecoder(t, tests)
}

func TestFATDecoderPUAChars(t *testing.T) {
	t.Parallel()
	tests := make([]decodeTest, 0, fat.NumChars)
	for r := rune(0); r < rune(fat.NumChars); r++ {
		ins := string(r | fat.BaseRune)
		out := ins
		decoded := strings.ContainsRune(fat.Encodes, r)
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
	testFATDecoder(t, tests)
}

func TestFATDecoderEncodes(t *testing.T) {
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
	testFATDecoder(t, tests)
}

func TestFATEncoderEncodes(t *testing.T) {
	t.Parallel()
	testFATEncoder(t, encodeTests)
}

func TestFATEncoderLatin1Chars(t *testing.T) {
	t.Parallel()
	tests := make([]encodeTest, 0, fat.NumChars)

	for r := rune(0); r < rune(fat.NumChars); r++ {
		ins := string(r)
		out := ins
		encoded := strings.ContainsRune(fat.Encodes, r)
		if encoded {
			out = string(r | fat.BaseRune)
		}
		test := encodeTest{
			in:      ins,
			out:     out,
			encoded: encoded,
		}
		tests = append(tests, test)
	}
	testFATEncoder(t, tests)
}

func TestFATEncoderPUAChars(t *testing.T) {
	t.Parallel()
	tests := make([]encodeTest, 0, fat.NumChars)

	for r := rune(0); r < rune(fat.NumChars); r++ {
		ins := string(r | fat.BaseRune)
		out := ins
		encoded := strings.ContainsRune(fat.Encodes, r)
		test := encodeTest{
			in:      ins,
			out:     out,
			encoded: encoded,
		}
		tests = append(tests, test)
	}
	testFATEncoder(t, tests)
}

func TestFATEncoderDecodes(t *testing.T) {
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
	testFATEncoder(t, tests)
}

// Test the pattern encoder with the encodeTests tests.
func TestFATPatternEncoder(t *testing.T) {
	t.Parallel()
	enc := fat.PUAPattern.NewEncoder()

	if testing.Verbose() {
		t.Logf("Running %d tests", len(encodeTests)*len(getLengths()))
	}

	for i, test := range encodeTests {
		for _, length := range getLengths() {
			ins := strings.Repeat(test.in, length)
			out := strings.Repeat(test.out, length)
			got, err := enc.String(ins)
			if err != nil {
				t.Errorf("Test %d: PUAPattern.Encode(%+q) unexpected error; %v", i+1, ins, err)
			}
			if got != out {
				t.Errorf("Test %d: PUAPattern.Encode(%+q) got %+q; want %+q (%d vs %d bytes)", i+1, ins, got, out, len(got), len(out))
			}

			got2, err := fat.Encode(ins, true)
			want2 := got
			if err != nil {
				t.Errorf("Test %d: Encode(%+q, true) unexpected error; %v", i+1, ins, err)
			}
			if got2 != want2 {
				t.Errorf("Test %d: Encode(%+q, true) got %v; want %v (%d vs %d bytes)", i+1, ins, got2, want2, len(got2), len(want2))
			}

			got3 := fat.IsEncoded(out)
			want3 := test.encoded
			if got3 != want3 {
				t.Errorf("Test %d: IsEncoded(%+q) got %v; want %v", i+1, out, got3, want3)
			}
		}
	}
}

// Test the pattern encoder with the patternEncodeTests tests.
func TestFATPatternEncoderSpecific(t *testing.T) {
	t.Parallel()
	enc := fat.PUAPattern.NewEncoder()

	if testing.Verbose() {
		t.Logf("Running %d tests", len(patternEncodeTests)*len(getLengths()))
	}

	for i, test := range patternEncodeTests {
		for _, length := range getLengths() {
			ins := strings.Repeat(test.in, length)
			out := strings.Repeat(test.out, length)
			got, err := enc.String(ins)
			if err != nil {
				t.Errorf("Test %d: FATPattern.Encode(%+q) unexpected error; %v", i+1, ins, err)
			}
			if got != out {
				t.Errorf("Test %d: FATPattern.Encode(%+q) got %+q; want %+q (%d vs %d bytes)", i+1, ins, got, out, len(got), len(out))
			}
		}
	}
}

func TestFATIsDecoded(t *testing.T) {
	t.Parallel()
	testFATDecoder(t, decodeTests)
}

func testFATDecoder(t *testing.T, tests []decodeTest) {
	t.Helper()

	if testing.Verbose() {
		t.Logf("Running %d tests", len(tests)*len(getLengths()))
	}

	dec := fat.PUA.NewDecoder()

	for i, test := range tests {
		for _, length := range getLengths() {
			ins := strings.Repeat(test.in, length)
			out := strings.Repeat(test.out, length)
			got, err := dec.String(ins)
			if err != nil {
				t.Errorf("Test %d: PUA.Decode(%q) got %v, want %v", i+1, ins, err, nil)
			}
			if got != out {
				t.Errorf("Test %d: PUA.Decode(%q) got %+q; want %+q (%d vs %d bytes)", i, ins, got, out, len(got), len(out))
			}

			got2 := fat.Decode(ins)
			want2 := out
			if got2 != want2 {
				t.Errorf("Test %d: Decode(%q) got %v; want %v (%d vs %d bytes)", i, ins, got2, want2, len(got2), len(want2))
			}

			got3 := fat.IsDecoded(out)
			want3 := test.decoded
			if got3 != want3 {
				t.Errorf("Test %d: IsDecoded(%q) got %v; want %v", i, out, got3, want3)
			}
		}
	}
}

func testFATEncoder(t *testing.T, tests []encodeTest) {
	t.Helper()

	enc := fat.PUA.NewEncoder()

	if testing.Verbose() {
		t.Logf("Running %d tests", len(tests)*len(getLengths()))
	}

	for i, test := range tests {
		for _, length := range getLengths() {
			ins := strings.Repeat(test.in, length)
			out := strings.Repeat(test.out, length)
			got, err := enc.String(ins)
			if err != nil {
				t.Errorf("Test %d: PUA.Encode(%+q) got %v, want %v", i+1, ins, err, nil)
			}
			if got != out {
				t.Errorf("Test %d: PUA.Encode(%+q) got %+q; want %+q (%d vs %d bytes)", i+1, ins, got, out, len(got), len(out))
			}

			got2, err := fat.Encode(ins, false)
			want2 := out
			if err != nil {
				t.Errorf("Test %d: Encode(%+q, false) unexpected error; %v", i+1, ins, err)
			}
			if got2 != out {
				t.Errorf("Test %d: Encode(%+q, false) got %+q; want %+q (%d vs %d bytes)", i+1, ins, got2, out, len(got2), len(out))
			}
			if got2 != want2 {
				t.Errorf("Test %d: Encode(%+q, false) got %v; want %v", i+1, ins, got2, want2)
			}

			got3 := fat.IsEncoded(out)
			want3 := test.encoded
			if got3 != want3 {
				t.Errorf("Test %d: IsEncoded(%+q) got %v; want %v: length=%v", i+1, out, got3, want3, length)
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
	dec := fat.PUA.NewDecoder()
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
	enc := fat.PUA.NewEncoder()
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
	enc := fat.PUAPattern.NewEncoder()
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
Benchmark results for various puaDecodeTransformer functions in fat.go:

// BenchmarkFatDecoder-4            1973989              6487 ns/op
//
var puaDecodeTransformer = runes.Map(func(r rune) rune {
	if r < BaseRune || r >= (BaseRune+NumChars) {
		return r
	}
	if puaEncodes[r&^BaseRune] >= BaseRune {
		return r &^ BaseRune
	}

	return r
})

// BenchmarkFatDecoder-4            1861136              6589 ns/op
//
var puaDecodeTransformer = runes.Map(func(r rune) rune {
	if r < BaseRune || r >= (BaseRune+NumChars) {
		return r
	}
	if puaEncodes[r-BaseRune] >= BaseRune {
		return r &^ BaseRune
	}
	return r
})

// BenchmarkFatDecoder-4            1648190              6779 ns/op
//
var puaDecodeTransformer = runes.Map(func(r rune) rune {
	if r < BaseRune || r >= (BaseRune+NumChars) {
		return r
	}
	if puaEncodes[r-BaseRune] >= BaseRune {
		return r - BaseRune
	}
	return r
})

// BenchmarkFatDecoder-4            1665566              7235 ns/op
//
var decodeMap map[rune]rune
func init() {
	decodeMap = make(map[rune]rune, len(puaEncodes))
	for _, r := range fatDecodes {
		decodeMap[r] = r &^ BaseRune
	}
}
var puaDecodeTransformer = runes.Map(func(r rune) rune {
	r, ok := decodeMap[r]
	if ok {
		return r
	}
	return r
})

// BenchmarkFatDecoder-4            1371726              8911 ns/op
//
var puaDecodeTransformer = runes.Map(func(r rune) rune {
	if strings.ContainsRune(puaEncodes, r) {
		return r &^ BaseRune
	}
})

*/
