// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

// Command stgenencoded generates files encoded by the FAT encoder. The tool has
// reduced functionality when run in Windows, or WSL, as it cannot create
// pre-encoded (decoded) filenames in these environments.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"unicode"

	_ "github.com/syncthing/syncthing/lib/automaxprocs"
	"github.com/syncthing/syncthing/lib/build"
	"github.com/syncthing/syncthing/lib/encoding/fat/consts"
	"github.com/syncthing/syncthing/lib/osutil/wsl"
)

var modeMap = map[string]string{
	"encoded": `Generate only encoded filenames (containing \uf0xx)`,
	"decoded": `Generate only pre-encoded filenames (containing "*:<>?|)`,
	"both":    "Generate both encoded and pre-encoded (decoded) filenames",
}

const inWSLMsg = "You appear to be running in a WSL environment"
const inWindowMsg = "You appear to be running in a Windows environment"
const noCanDoMsg = ",\nso it's very unlikely you can use this tool to\n" +
	"generate pre-encoded (decoded) filenames"
const butHeyMsg = ",\nbut hey, you asked, so maybe you know more than I do, so I'll try..."

func main() {
	log.SetFlags(0)
	var mode string
	var controls bool
	var backslash bool

	flag.Usage = usage
	flag.StringVar(&mode, "mode", "",
		"Select action: encoded, decoded, both")
	flag.BoolVar(&controls, "controls", false,
		`Generate files containing control characters (\x01-\x1f), too`)
	flag.BoolVar(&backslash, "backslash", false,
		`Generate files containing backslash character (\x5c), too`)

	flag.Parse()
	root := flag.Arg(0)
	if root == "" {
		root = "."
	}

	var err error
	switch mode {
	case "encoded":
		err = genEncodeds(root, controls, backslash)
	case "decoded":
		err = genDecodeds(root, controls, backslash)
	case "both":
		err = genEncodeds(root, controls, backslash)
		if err == nil {
			err = genDecodeds(root, controls, backslash)
		}
	default:
		usage()
	}
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func usage() {
	usageText := `
Usage: %s [options] [dir] ...

Options:

  --help
        Print this help text
`
	moreHelp := `
Mode option:

`

	for key, value := range modeMap {
		moreHelp += fmt.Sprintf("  %-8s: %s\n", key, value)
	}

	_, arg0 := filepath.Split(os.Args[0])
	usageText = fmt.Sprintf(usageText, arg0)
	fmt.Fprintln(os.Stderr, usageText)
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, moreHelp)
	os.Exit(1)
}

func genEncodeds(root string, controls bool, backslash bool) error {
	for _, r := range consts.Encodes + `\` {
		if !controls && unicode.IsControl(r) {
			continue
		}
		if !backslash && r == '\\' {
			continue
		}
		encoded := r | consts.BaseRune
		err := genFiles(root, r, encoded)
		if err != nil {
			log.Println(err)

			return err
		}
	}
	if build.IsWindows || wsl.IsWSL() {
		log.Println(
			"Encoded files were generated, but they will look like pre-encoded (decoded)\n" +
				"filenames inside Cygwin/GitBash/Msys2/WSL environments.")
	}
	return nil
}

func genDecodeds(root string, controls bool, backslash bool) error {
	if build.IsWindows {
		log.Println(inWindowMsg + noCanDoMsg + butHeyMsg)
	}
	if wsl.IsWSL() {
		log.Println(inWSLMsg + noCanDoMsg + butHeyMsg)
	}

	for _, r := range consts.Encodes + `\` {
		if !controls && unicode.IsControl(r) {
			continue
		}
		if !backslash && r == '\\' {
			continue
		}
		err := genFiles(root, r, r)
		if err != nil {
			log.Println(err)

			return err
		}
	}

	return nil
}

func genFiles(root string, arune rune, encoded rune) error {
	name := fmt.Sprintf("0x%04x-%c-reg.tmp", arune, encoded)
	err := genFile(root, name)
	if err != nil {
		return err
	}
	name = fmt.Sprintf("0x%04x-%c-dir.tmp", arune, encoded)
	err = genDir(root, name)
	if err != nil {
		return err
	}

	oldname := fmt.Sprintf("0x%04x-tar.tmp", arune)
	newname := fmt.Sprintf("0x%04x-%c-lnk.tmp", arune, encoded)
	err = genSymlink(root, oldname, newname)
	if err != nil {
		return err
	}

	return nil
}

func genFile(root string, name string) error {
	path := path.Join(root, name)
	log.Printf("Creating %s\n", path)
	hnd, err := os.Create(path)
	if err != nil {
		return err
	}
	defer hnd.Close()
	_, err = hnd.WriteString("same content for all files")

	return err
}

func genDir(root string, name string) error {
	path := path.Join(root, name)
	log.Printf("Creating %s\n", path)

	return os.MkdirAll(path, os.FileMode(0o775))
}

func genSymlink(root string, oldName, newName string) error {
	oldPath := path.Join(root, oldName)
	newPath := path.Join(root, newName)
	err := genFile(root, oldName)
	if err != nil {
		return err
	}

	log.Printf("Symlinking %s to %s\n", oldPath, newPath)

	return os.Symlink(oldPath, newPath)
}
