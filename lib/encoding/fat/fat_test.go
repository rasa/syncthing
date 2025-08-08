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
	"github.com/syncthing/syncthing/lib/encoding/fat/consts"
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

func init() {
	// Android encodes '\x7F', but Windows doesn't
	if strings.ContainsRune(consts.Encodes, '\x7F') {
		encodeTests = append(encodeTests, encodeTest{"\u007F", "\uF07F", true})
	} else {
		encodeTests = append(encodeTests, encodeTest{"\u007F", "\u007F", false})
	}
}

func TestFATDecoderDecodes(t *testing.T) {
	t.Parallel()
	testFATDecoder(t, decodeTests)
}

func TestFATDecoderLatin1Chars(t *testing.T) {
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
	testFATDecoder(t, tests)
}

func TestFATDecoderPUAChars(t *testing.T) {
	t.Parallel()
	tests := make([]decodeTest, 0, consts.NumChars)
	for r := rune(0); r < rune(consts.NumChars); r++ {
		ins := string(r | consts.BaseRune)
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
	tests := make([]encodeTest, 0, consts.NumChars)

	for r := rune(0); r < rune(consts.NumChars); r++ {
		ins := string(r)
		out := ins
		encoded := strings.ContainsRune(consts.Encodes, r)
		if encoded {
			out = string(r | consts.BaseRune)
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
	tests := make([]encodeTest, 0, consts.NumChars)

	for r := rune(0); r < rune(consts.NumChars); r++ {
		ins := string(r | consts.BaseRune)
		out := ins
		encoded := strings.ContainsRune(consts.Encodes, r)
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
		j := i + 1
		for _, length := range getLengths() {
			ins := strings.Repeat(test.in, length)
			want := strings.Repeat(test.out, length)
			got, err := enc.String(ins)
			if err != nil {
				t.Errorf("Test %d: PUAPattern.Encode(%+q) unexpected error; %v", j, ins, err)
			}
			if got != want {
				t.Errorf("Test %d: PUAPattern.Encode(%+q) got %+q; want %+q (%d vs %d bytes)", j, ins, got, want, len(want), len(want))
			}

			got, err = fat.EncodePattern(ins)
			if err != nil {
				t.Errorf("Test %d: EncodePattern(%+q, true) unexpected error; %v", j, ins, err)
			}
			if got != want {
				t.Errorf("Test %d: EncodePattern(%+q, true) got %v; want %v (%d vs %d bytes)", j, ins, got, want, len(got), len(want))
			}

			got = fat.MustEncodePattern(ins)
			if got != want {
				t.Errorf("Test %d: MustEncodePattern(%+q, true) got %v; want %v (%d vs %d bytes)", j, ins, got, want, len(got), len(want))
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
		j := i + 1
		for _, length := range getLengths() {
			ins := strings.Repeat(test.in, length)
			want := strings.Repeat(test.out, length)
			got, err := dec.String(ins)
			if err != nil {
				t.Errorf("Test %d: PUA.Decode(%q) got %v, want %v", j, ins, err, nil)
			}
			if got != want {
				t.Errorf("Test %d: PUA.Decode(%q) got %+q; want %+q (%d vs %d bytes)", j, ins, got, want, len(got), len(want))
			}

			got, err = fat.Decode(ins)
			if err != nil {
				t.Errorf("Test %d: Decode(%+q) unexpected error; %v", j, ins, err)
			}
			if got != want {
				t.Errorf("Test %d: Decode(%+q) got %+q; want %+q (%d vs %d bytes)", j, ins, got, want, len(got), len(want))
			}

			got = fat.MustDecode(ins)
			if got != want {
				t.Errorf("Test %d: MustDecode(%+q) got %+q; want %+q (%d vs %d bytes)", j, ins, got, want, len(got), len(want))
			}

			got2 := fat.IsDecoded(want)
			want2 := test.decoded
			if got2 != want2 {
				t.Errorf("Test %d: IsDecoded(%q) got %v; want %v", j, ins, got2, want2)
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
		j := i + 1
		for _, length := range getLengths() {
			ins := strings.Repeat(test.in, length)
			want := strings.Repeat(test.out, length)
			got, err := enc.String(ins)
			if err != nil {
				t.Errorf("Test %d: PUA.Encode(%+q) got %v, want %v", j, ins, err, nil)
			}
			if got != want {
				t.Errorf("Test %d: PUA.Encode(%q) got %+q; want %+q (%d vs %d bytes)", j, ins, got, want, len(got), len(want))
			}

			got, err = fat.Encode(ins)
			if err != nil {
				t.Errorf("Test %d: Encode(%+q) unexpected error; %v", j, ins, err)
			}
			if got != want {
				t.Errorf("Test %d: Encode(%+q) got %+q; want %+q (%d vs %d bytes)", j, ins, got, want, len(got), len(want))
			}

			got = fat.MustEncode(ins)
			if got != want {
				t.Errorf("Test %d: MustEncode(%+q) got %+q; want %+q (%d vs %d bytes)", j, ins, got, want, len(got), len(want))
			}

			got2 := fat.IsEncoded(want)
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
	b.ReportAllocs()

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
	b.ReportAllocs()

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
Benchmark results for various defaultDecodeTransformer functions in fat.go:

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
