// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

// Command stfindencoded finds files encoded by the FAT encoder. The tool has
// reduced functionality when run in Windows, or WSL, as you cannot create
// pre-encoded (decoded) filenames in these environments.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"

	_ "github.com/syncthing/syncthing/lib/automaxprocs"
	"github.com/syncthing/syncthing/lib/build"
	"github.com/syncthing/syncthing/lib/encoding/fat"
	"github.com/syncthing/syncthing/lib/encoding/fat/consts"
	"github.com/syncthing/syncthing/lib/fs"
	"github.com/syncthing/syncthing/lib/osutil/wsl"
)

// Mode is the mode selected by the user, via the --mode option.
type Mode int

const (
	modeEncoded Mode = iota
	modeDecoded
	modeBoth
	modeDuplicates
	modeFix
)

// FixMode is the mode selected by the user, via the --default option.
type FixMode int

const (
	fixModeManual FixMode = iota
	fixModeEncoded
	fixModeDecoded
	fixModeNewer
	fixModeOlder
	fixModeListOnly
)

// Action is the action to take to resolve duplicates.
type Action int

const (
	actionSkip Action = iota
	actionKeepEncoded
	actionKeepDecoded
)

type stats struct {
	regular    int // regular (non-encoded/decoded)
	encodes    int // encodes without a decode
	decodes    int // decodes without an encode
	duplicates int // file pairs found
	fixed      int
	skipped    int
}

var fixModeMap = map[string]string{
	"manual":  "Both filenames will be shown, and you can choose which one to keep (default)",
	"decoded": "The original (pre-encoded) filename will be kept, the encoded filename will be removed",
	"encoded": "The encoded filename will be kept, the original (pre-encoded) filename will be removed",
	"newer":   "The newer file will be kept, the older file will be removed",
	"older":   "The older file will be kept, the newer file will be removed",
}

const (
	inWSLMsg     = "You appear to be running in a WSL environment"
	inWindowsMsg = "You appear to be running in a Windows environment"
	noEncodeMsg  = ",\nso you almost certainly cannot use this tool to find\n" +
		"encoded filenames, as they will appear as pre-encoded (decoded) filenames."
	noDecodedMsg = ",\nso you almost certainly can't find duplicates, as encoded filenames will\n" +
		"appear as pre-encoded (decoded) filenames,"
	noDuplicatesMsg = ",\nso you almost certainly cannot use this tool to find\n" +
		"encoded filenames, as they will appear as pre-encoded (decoded) filenames."
	butHeyMsg = ",\nbut hey, you asked, so maybe you know more than I do, so I'll try..."
)

func main() {
	log.SetFlags(0)
	var mode string
	var defFixMode string
	var long bool

	flag.Usage = usage
	flag.StringVar(&mode, "mode", "encoded",
		"Set action: encoded, decoded, both, duplicates, fix")
	flag.StringVar(&defFixMode, "default", "manual",
		"Set default fix action: manual, decoded, encoded, older, newer")
	flag.BoolVar(&long, "long", false,
		"Use a long listing format")

	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}

	var fixMode FixMode
	switch defFixMode {
	case "manual":
		fixMode = fixModeManual
	case "decoded":
		fixMode = fixModeDecoded
	case "encoded":
		fixMode = fixModeEncoded
	case "newer":
		fixMode = fixModeNewer
	case "older":
		fixMode = fixModeOlder
	default:
		usage()
	}

	switch mode {
	case "encoded":
		for _, arg := range args {
			find(modeEncoded, arg, long)
		}
	case "decoded":
		for _, arg := range args {
			find(modeDecoded, arg, long)
		}
	case "both":
		for _, arg := range args {
			find(modeBoth, arg, long)
		}
	case "duplicates":
		for _, arg := range args {
			findDuplicates(arg, fixModeListOnly)
		}
	case "fix":
		stdout("Mode %s: %s", defFixMode, fixModeMap[defFixMode])
		for _, arg := range args {
			findDuplicates(arg, fixMode)
		}
	default:
		usage()
	}
}

func usage() {
	usageText := `
Usage: %s [options] [dir] [dir2] ...

Options:

  --help
        Print this help text
`
	moreHelp := `
Mode option:

  encoded:    Display filenames that are encoded (default if no mode specified)
  decoded:    Display filenames that would be encoded, if synced
  both:       Display filenames that are, or would be encoded, if synced
  duplicates: Display file pairs that have both encoded and original (pre-encoded) filenames
  fix:        Display duplicates and optionally delete one of the duplicates

Default option (when -mode=fix is selected):

`

	for key, value := range fixModeMap {
		moreHelp += fmt.Sprintf("  %-8s: %s\n", key, value)
	}

	_, arg0 := filepath.Split(os.Args[0])
	usageText = fmt.Sprintf(usageText, arg0)
	fmt.Fprintln(os.Stderr, usageText)
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, moreHelp)
	os.Exit(1)
}

func find(mode Mode, root string, long bool) {
	msg := ""
	switch mode {
	case modeEncoded:
		msg = "Scanning %s for encoded filenames"
		if wsl.IsWSL() {
			log.Println(inWSLMsg + noEncodeMsg + butHeyMsg)
		}
	case modeDecoded:
		msg = "Scanning %s for filenames that would be encoded"
		if wsl.IsWSL() {
			log.Println(inWSLMsg + noDecodedMsg + butHeyMsg)
		}
	case modeBoth:
		msg = "Scanning %s for encoded filenames and those that would be encoded"
		if wsl.IsWSL() {
			log.Println(inWSLMsg + noDecodedMsg + butHeyMsg)
		}
	}

	stdout(msg, root)
	found := 0

	vfs := fs.NewWalkFilesystem(fs.NewFilesystem(fs.FilesystemTypeBasic, root))

	regex := "[" + regexp.QuoteMeta(consts.Nevers) + "]"
	nevers := regexp.MustCompile(regex)
	_ = vfs.Walk(".", func(name string, _ fs.FileInfo, err error) error {
		path := filepath.Join(root, name)
		if err != nil {
			log.Printf("Warning: %s: %v\n", path, err.Error())

			return fs.SkipDir
		}
		switch mode {
		case modeDecoded, modeBoth:
			if nevers.MatchString(name) {
				fi, err := stat(path)
				if err != nil {
					return err
				}
				out := fi.Name()
				if long {
					out = fi.String()
				}
				stdout("%s %s", out, "(unencodeable)")

				return nil
			}
		}

		switch mode {
		case modeEncoded:
			if !fat.IsEncoded(name) {
				return nil
			}
		case modeDecoded:
			if !fat.IsDecoded(path) {
				return nil
			}
		case modeBoth:
			if !fat.IsEncoded(name) && !fat.IsDecoded(path) {
				return nil
			}
		}
		found++

		fi, err := stat(path)
		if err != nil {
			return err
		}
		out := fi.Name()
		if long {
			out = fi.String()
		}
		switch mode {
		case modeEncoded:
			decoded := fat.MustDecode(name)
			stdout("%s (%s decoded)", out, decoded)
		case modeDecoded, modeBoth:
			stdout("%s", out)
		}

		return nil
	})
	stdout("Found %d encoded/encodable files", found)
}

func findDuplicates(root string, fixMode FixMode) {
	if wsl.IsWSL() {
		log.Println(inWSLMsg + noDuplicatesMsg + butHeyMsg)
	}
	if build.IsWindows {
		log.Println(inWindowsMsg + noDuplicatesMsg + butHeyMsg)
	}

	stats := stats{}

	stdout("Scanning %s for duplicate filenames (both encoded and pre-encoded (decoded) versions)", root)

	vfs := fs.NewWalkFilesystem(fs.NewFilesystem(fs.FilesystemTypeBasic, root))

	err := vfs.Walk(".", func(name string, info fs.FileInfo, err error) error {
		path := filepath.Join(root, name)
		if err != nil {
			log.Printf("Warning: %s: %v\n", path, err.Error())

			return fs.SkipDir
		}
		if !info.IsDir() {
			return nil
		}

		decoded := quoted(name)
		if decoded != name {
			decoded = " (" + decoded + " decoded)"
		} else {
			decoded = ""
		}
		stdout("In directory %s%s", name, decoded)
		files, err := vfs.DirNames(name)
		if err != nil {
			return err
		}
		decodes := make(map[string]bool)

		// Creates map of decoded files.
		for _, file := range files {
			if !fat.IsDecoded(file) {
				continue
			}
			key := fat.MustDecode(file)
			decodes[key] = true
		}

		slices.Sort(files)

		for _, eFile := range files {
			dFile := fat.MustDecode(eFile)
			if !fat.IsEncoded(eFile) {
				if fat.IsDecoded(eFile) {
					if !decodes[dFile] {
						stats.decodes++
					}
				} else {
					stats.regular++
				}
				continue
			}
			// Does the encoded filename's decoded cousin exist?
			if !decodes[dFile] {
				stats.encodes++
				continue
			}

			stats.duplicates++

			fmt.Println("")
			dPath := filepath.Join(root, dFile)
			ePath := filepath.Join(root, eFile)

			dfi, err := stat(dPath)
			if err != nil {
				log.Print(err)

				continue
			}
			efi, err := stat(ePath)
			if err != nil {
				log.Print(err)

				continue
			}

			diffs := make(map[string]string)

			if efi.IsRegular() && dfi.IsRegular() {
				dSum := sha256sum(dPath)
				eSum := sha256sum(ePath)
				if dSum != eSum {
					diffs["hashes"] = "hashes"
				}
				if dfi.Size() != efi.Size() {
					diffs["sizes"] = strconv.FormatInt(dfi.Size()-efi.Size(), 10)
				}
				dTime := dfi.ModTime().Round(time.Second)
				eTime := efi.ModTime().Round(time.Second)
				if dTime != eTime {
					diffs["times"] = dTime.Sub(eTime).String()
				}
				if dfi.Mode() != efi.Mode() {
					diffs["attributes"] = ""
					dMode := dfi.Mode().Perm().String()
					eMode := efi.Mode().Perm().String()
					for i := range dMode {
						if dMode[i] == eMode[i] {
							diffs["attributes"] += " "
						} else {
							diffs["attributes"] += string(dMode[i])
						}
					}
				}
			}
			stdout("D: %s", dfi.String())
			stdout("E: %s", efi.String())

			// if len(diffs) > 0 {
			//          1         2         3         4         5
			// 123456789012345678901234567890123456789012345678901234567890
			// 2: -rw-rw-r--        12 2024-05-31 16:18:37 0x3f-.tmp
			// Δ: 1234567890 123456789 1234567890123456789 1
			extra := ""
			times := diffs["times"]
			if times != "" {
				if times[0] == '-' {
					extra = "older"
				} else {
					extra = "newer"
				}
			}
			sizes := diffs["sizes"]
			if sizes != "" {
				if extra != "" {
					extra += "/"
				}
				if sizes[0] == '-' {
					extra += "smaller"
				} else {
					extra += "bigger"
				}
			}
			if extra != "" {
				extra = " (D. is " + extra + ")"
			}
			stdout("Δ: %10s %9s %19s %6s%s",
				diffs["attributes"], diffs["sizes"], diffs["times"], diffs["hashes"], extra)
			//}

			action := actionSkip
			why := ""
			switch fixMode {
			case fixModeManual:
				for {
					fmt.Printf("Keep: (D)ecoded, (E)ncoded, (N)ewer, (O)lder, (S)kip, (Q)uit? ")
					c, err := getch()
					if err != nil {
						log.Fatal("\n" + err.Error())
					}
					fmt.Printf("%c\n", c)
					switch c {
					case 'd', 'D', 'p', 'P': // (P)re-encoded
						action = actionKeepDecoded
					case 'e', 'E':
						action = actionKeepEncoded
					case 'n', 'N':
						action = getNewerAction(dfi, efi)
						why = " (newer)"
					case 'o', 'O':
						action = getOlderAction(dfi, efi)
						why = " (older)"
					case 's', 'S':
					case 'q', 'Q', '\x03': // Ctrl-C
						os.Exit(0)
					default:
						// bad input, try again
						continue
					}

					break
				}
			case fixModeDecoded:
				action = actionKeepDecoded
			case fixModeEncoded:
				action = actionKeepEncoded
			case fixModeNewer:
				why = " (newer)"
				action = getNewerAction(dfi, efi)
			case fixModeOlder:
				why = " (older)"
				action = getOlderAction(dfi, efi)
			case fixModeListOnly:
				// noop
			}

			switch action {
			case actionKeepDecoded:
				stdout("Keeping D: %s%s", dPath, why)
				err := os.RemoveAll(ePath)
				if err != nil {
					log.Fatalf("Failed to remove %s: %s\n", ePath, err)
				}
				stats.fixed++
			case actionKeepEncoded:
				stdout("Keeping E: %s%s", ePath, why)
				if dfi.IsDir() {
					err := os.RemoveAll(dPath)
					if err != nil {
						log.Fatalf("Failed to remove directory %s: %s\n", dPath, err)
					}
				}
				err := os.Rename(ePath, dPath)
				if err != nil {
					log.Fatalf("Failed to rename %q to %q: %s\n", ePath, dPath, err)
				}
				stats.fixed++
			case actionSkip:
				stats.skipped++
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	total := stats.duplicates + stats.decodes + stats.encodes + stats.regular

	stdout("\nTotal found:")
	stdout("Duplicates: %8d (matching encoded/decodes name pairs)", stats.duplicates)
	stdout("Encoded:    %8d (lone encoded file/directory names)", stats.encodes)
	stdout("Decoded:    %8d (lone decoded file/directory names)", stats.decodes)
	stdout("Regular:    %8d (non-encoded/decoded names)", stats.regular)
	stdout("Total:      %8d (files and directories)", total)
	stdout("Actions taken:")
	stdout("Fixed:      %8d", stats.fixed)
	stdout("Skipped:    %8d", stats.skipped)
}

type encFileInfo struct {
	os.FileInfo
	target string
}

func (e encFileInfo) IsSymlink() bool {
	return e.Mode()&os.ModeSymlink != 0
}

func (e encFileInfo) IsRegular() bool {
	return e.Mode()&os.ModeType == 0
}

func (e encFileInfo) Target() string {
	return e.target
}

func (e encFileInfo) String() string {
	mode := e.Mode().Perm().String()
	if e.IsDir() {
		mode = "d" + mode[1:]
	}

	time := fmt.Sprintf("%v", e.ModTime())
	parts := strings.Split(time, " ")
	parts2 := strings.Split(parts[1], ".")
	time = parts[0] + " " + parts2[0]

	target := ""
	if e.target != "" {
		mode = "l" + mode[1:]
		target = " -> " + e.target
	}

	return fmt.Sprintf("%s %9d %s %s%s",
		mode,
		e.Size(),
		time,
		e.Name(),
		target)
}

func stat(path string) (encFileInfo, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return encFileInfo{fileInfo, ""}, err
	}
	target := ""
	lstat, err := os.Lstat(path)
	if err != nil {
		return encFileInfo{fileInfo, ""}, err
	}
	if (lstat.Mode() & os.ModeSymlink) == os.ModeSymlink {
		target, err = os.Readlink(path)
		if err != nil {
			return encFileInfo{fileInfo, ""}, err
		}
	}

	return encFileInfo{fileInfo, target}, nil
}

func getNewerAction(dfi encFileInfo, efi encFileInfo) Action {
	if dfi.ModTime().Compare(efi.ModTime()) >= 0 {
		return actionKeepDecoded
	}

	return actionKeepEncoded
}

func getOlderAction(dfi encFileInfo, efi encFileInfo) Action {
	if dfi.ModTime().Compare(efi.ModTime()) <= 0 {
		return actionKeepDecoded
	}

	return actionKeepEncoded
}

func sha256sum(file string) string {
	hasher := sha256.New()
	s, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	hasher.Write(s)
	if err != nil {
		log.Fatal(err)
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

func quoted(path string) string {
	quoted := fmt.Sprintf("%q", path)
	quoted = strings.Trim(quoted, `"`)
	re := regexp.MustCompile(`\\\\`)
	quoted = re.ReplaceAllString(quoted, `\`)
	re = regexp.MustCompile(`\\\"`)
	quoted = re.ReplaceAllString(quoted, `"`)

	return quoted
}

func getch() (rune, error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return -1, err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	b := make([]byte, 1)
	_, err = os.Stdin.Read(b)
	if err != nil {
		return -1, err
	}
	return rune(b[0]), nil
}

func stdout(format string, a ...any) {
	fmt.Printf(format+"\n", a...)
}
