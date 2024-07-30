// Copyright (C) 2024 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package integration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
	"testing"
	"time"
	"unicode"

	"github.com/syncthing/syncthing/lib/build"
	"github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/encoding/fat/consts"
	"github.com/syncthing/syncthing/lib/fs"
	"github.com/syncthing/syncthing/lib/logger"
	"github.com/syncthing/syncthing/lib/rand"
	"github.com/syncthing/syncthing/lib/rc"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	numberOfSubTests   = 2 // Currently just OneSide & MergeTwo
	secondsBeforePanic = 1 // seconds to deduct from deadline to avoid panic
	alnum              = "0123456789abcdefghijklmnopqrstuvwxyz"
)

var (
	srcTypes   = []srcType{srcTypeDecoded, srcTypeEncoded}
	srcTypeMap = map[srcType]string{
		srcTypeDecoded: "decoded",
		srcTypeEncoded: "encoded",
	}
)

var dstTypeMap = map[dstType]string{
	dstTypeDecoded:       "decoded",
	dstTypeEncoded:       "encoded",
	dstTypeRejectEncoded: "non-encoded",
	dstTypeSkipped:       "<skipped>", // not used
}

/*
Test matrix summary

Test  src     dst     Created Created
Nos.  Encoder Encoder on src  on dst  comments
----- ------- ------- ------- ------- ------------------------------------
1-2   Fat     None    Decoded Skipped src Fat encoders don't decode decoded filenames
3-4   Fat     None    Encoded Decoded src Fat encoder decodes filenames to send

5-6   Fat     Fat     Decoded Skipped src Fat encoders don't decode decoded filenames
7-8   Fat     Fat     Encoded Encoded src Fat encoder decodes filenames to send
"                                     dst Fat encoder encodes filenames before saving

9-10  None    None    Decoded Decoded
11-12 None    None    Encoded Encoded

13-14 None    Fat     Decoded Encoded dst Fat encoder encodes filenames before saving
15-16 None    Fat     Encoded Reject* dst Fat encoder saves encoded filenames,
"                                     but rejects encoded filenames on wire.
*/
var testResultMatrix = map[fs.EncoderType]map[fs.EncoderType]map[srcType]dstType{
	fs.EncoderTypeFat: { // src
		fs.EncoderTypeNone: { // dst
			srcTypeDecoded: dstTypeSkipped,
			srcTypeEncoded: dstTypeDecoded,
		},
		fs.EncoderTypeFat: {
			srcTypeDecoded: dstTypeSkipped,
			srcTypeEncoded: dstTypeEncoded,
		},
	},
	fs.EncoderTypeNone: {
		fs.EncoderTypeNone: {
			srcTypeDecoded: dstTypeDecoded,
			srcTypeEncoded: dstTypeEncoded,
		},
		fs.EncoderTypeFat: {
			srcTypeDecoded: dstTypeEncoded,
			srcTypeEncoded: dstTypeRejectEncoded, // do this one last as it times out
		},
	},
}

var (
	// filesToSync must be at least 4 and a multiple of 2 to run all tests.
	filesToSync  = 128
	testNumber   = 0
	skippedTests = 0
	// Exclude EncoderTypeUnset
	totalTests = (len(fs.EncoderType_name) - 1) * 2 * len(srcTypes) *
		numberOfSubTests
	maxSecondsPerTest = 600 / totalTests
	startTime         = time.Now()
)

// TestEncoderSync tests the encoderFS using the testResultMatrix above. The
// testResultMatrix has eight entries, and each entry runs two tests: 1. syncing
// one peer to another (testEncoderSyncOneSideToOther), and 2. syncing two peers
// with each other (testEncoderSyncMergeTwoDevices). This results in 16 total
// tests, but two tests are "skipped", as they can't really be performed, as
// FAT filesystems reject pre-encoded (decoded) filenames.
func TestEncoderSync(tttt *testing.T) {
	tttt.Parallel()

	dl, _ := tttt.Deadline()
	maxSecondsPerTest = int(dl.Sub(time.Now()).Seconds()) / totalTests

	if os.Getenv("STTRACE") != "" {
		logger.DefaultLogger.SetFlags(logger.DebugFlags)
	}

	if testing.Short() {
		filesToSync = 4
		maxSecondsPerTest = 20
	}

	// Sort the srcEncoderTypeIDs in descending order so the test that times out
	// runs last.
	srcEncoderTypeIDs := make([]int32, 0, len(fs.EncoderType_name))
	for id := range fs.EncoderType_name {
		if fs.EncoderType(id) == fs.EncoderTypeUnset {
			continue
		}
		srcEncoderTypeIDs = append(srcEncoderTypeIDs, id)
	}
	sort.Slice(srcEncoderTypeIDs, func(i, j int) bool {
		return srcEncoderTypeIDs[i] > srcEncoderTypeIDs[j]
	})

	for _, srcEncoderTypeID := range srcEncoderTypeIDs {
		srcEncoderType := fs.EncoderType(srcEncoderTypeID)
		tttName := "Src" + title(srcEncoderType.String()) + "Encoder"
		tttt.Run(tttName, func(ttt *testing.T) {
			for dstEncoderTypeID := range fs.EncoderType_name {
				dstEncoderType := fs.EncoderType(dstEncoderTypeID)
				if dstEncoderType == fs.EncoderTypeUnset {
					continue
				}
				ttName := "Dst" + title(dstEncoderType.String()) + "Encoder"
				ttt.Run(ttName, func(tt *testing.T) {
					for _, srcType := range srcTypes {
						tName := title(srcTypeMap[srcType])
						tt.Run(tName, func(t *testing.T) {
							testEncoderSyncAll(t, srcEncoderType, dstEncoderType, srcType)
						})
					}
				})
			}
		})
	}
}

// testEncoderSyncAll checks if the tests should be run and if so, runs the
// testEncoderSyncOneSideToOther and testEncoderSyncMergeTwoDevices tests.
func testEncoderSyncAll(t *testing.T, srcEncoder, dstEncoder fs.EncoderType, srcType srcType) {
	t.Helper()

	dstType, ok := testResultMatrix[srcEncoder][dstEncoder][srcType]
	if !ok {
		t.Fatalf("bug: missing entry in testResultMatrix for %v/%v/%v", srcEncoder, dstEncoder, srcType)
	}

	if dstType == dstTypeSkipped {
		skipSubTests(t, "Test %d of %d: Skipping as FAT encoders can't decode decoded filenames%s", "")
	}

	if build.IsWindows {
		key := ""
		if srcEncoder == fs.EncoderTypeNone && srcType == srcTypeDecoded {
			key = "src"
		}
		if dstEncoder == fs.EncoderTypeNone && dstType == dstTypeDecoded {
			key = "dst"
		}
		if key != "" {
			skipSubTests(t, "Test %d of %d: Skipping as the %v None encoder can't create decoded filenames on Windows", key)
		}
	}
	testNumber++
	t.Run("OneSide", func(t *testing.T) {
		testEncoderSyncOneSideToOther(t, srcEncoder, dstEncoder, srcType, dstType)
	})
	testNumber++
	t.Run("MergeTwo", func(t *testing.T) {
		if filesToSync%2 != 0 {
			skippedTests++
			t.Skipf("Test %d of %d: Skipping as this test requires filesToSync to be even", testNumber, totalTests)
		}
		if filesToSync < 4 {
			skippedTests++
			t.Skipf("Test %d of %d: Skipping as this test requires filesToSync to be 4 or greater", testNumber, totalTests)
		}
		testEncoderSyncMergeTwoDevices(t, srcEncoder, dstEncoder, srcType, dstType)
	})
}

// testEncoderSyncOneSideToOther verifies that files on one side get synced to the
// other. The test creates actual files on disk in a temp directory, so that
// the data can be compared after syncing. It was patterned after the
// TestSyncOneSideToOther test in sync_2dev_test.go.
func testEncoderSyncOneSideToOther(t *testing.T, srcEncoder, dstEncoder fs.EncoderType, srcType srcType, dstType dstType) {
	t.Helper()

	ctx, cancel := contextWithDeadline(t)
	defer cancel()

	// Create a source folder with some data in it.
	prefix := fmt.Sprintf("%02d-src-fold", testNumber)
	srcDir := getTempDir(t, prefix)
	srcPrefixes := srcPrefixes(srcType)
	created := generateTreeWithPrefixes(t, srcDir, filesToSync, srcPrefixes, "s")

	// Create an empty destination folder to hold the synced data.
	prefix = fmt.Sprintf("%02d-dst-fold", testNumber)
	dstDir := getTempDir(t, prefix)

	// Spin up two instances to sync the data.
	err := testEncoderSyncTwoDevicesFolders(ctx, t, srcDir, dstDir, srcEncoder, dstEncoder)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the destination folder now contains the same files as the source folder.
	walkResults := compareTreesByType(t, srcDir, dstDir, dstType)
	got := walkResults.found
	// The number of encoded/decoded filenames is only half of all filenames synced.
	synced := got / 2
	want := wanted(dstType, created, 0)

	if got != want {
		// Skip cleaning up, and progress to the next subtest.
		t.Fatalf("FAIL: dst %v encoder: got %d files (%d regular and %d %v filenames), wanted %d files",
			dstEncoder, got, synced, synced, dstTypeMap[dstType], want)
	}

	rejected := created - want
	suffix := ""
	if rejected != 0 {
		suffix = fmt.Sprintf(", and rejected %d encoded filenames received on the wire", rejected)
	}
	t.Logf("dst %v encoder synced %d files (%d regular and %d %v filenames)%v",
		dstEncoder, got, synced, synced, dstTypeMap[dstType], suffix)
	cleanup([]string{srcDir, dstDir})
}

// testEncoderSyncMergeTwoDevices verifies that two sets of files, one from each
// device, get merged into a coherent total. The test creates actual files
// on disk in a temp directory, so that the data can be compared after
// syncing. It is patterned after the TestSyncMergeTwoDevices test in
// sync_2dev_test.go.
func testEncoderSyncMergeTwoDevices(t *testing.T, srcEncoder, dstEncoder fs.EncoderType, srcType srcType, dstType dstType) {
	t.Helper()

	ctx, cancel := contextWithDeadline(t)
	defer cancel()

	filesPerPeer := filesToSync / 2

	// Create a source folder with some data in it.
	prefix := fmt.Sprintf("%02d-src-fold", testNumber)
	srcDir := getTempDir(t, prefix)
	srcPrefixes := srcPrefixes(srcType)
	srcCreated := generateTreeWithPrefixes(t, srcDir, filesPerPeer, srcPrefixes, "s")

	// Create an empty destination folder to hold the synced data.
	prefix = fmt.Sprintf("%02d-dst-fold", testNumber)
	dstDir := getTempDir(t, prefix)
	dstPrefixes := dstPrefixes(dstType)
	dstCreated := generateTreeWithPrefixes(t, dstDir, filesPerPeer, dstPrefixes, "d")

	// Spin up two instances to sync the data.
	err := testEncoderSyncTwoDevicesFolders(ctx, t, srcDir, dstDir, srcEncoder, dstEncoder)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the destination folder now contains the same files as the source folder.
	walkResults := compareTreesByType(t, srcDir, dstDir, dstType)

	got := walkResults.found
	// The number of encoded/decoded filenames is only half of all files synced.
	synced := got / 2
	want := wanted(dstType, srcCreated, dstCreated)

	if got != want {
		// Skip cleaning up, and progress to the next subtest.
		t.Fatalf("FAIL: dst %v encoder: got %d files (%d regular and %d %v filenames), wanted %d files",
			dstEncoder, got, synced, synced, dstTypeMap[dstType], want)
	}

	rejected := srcCreated + dstCreated - want
	suffix := ""
	if rejected != 0 {
		suffix = fmt.Sprintf(", and rejected %d encoded filenames received on the wire", rejected)
	}
	suffix2 := ""
	if build.IsWindows {
		suffix2 = fmt.Sprintf(", and Windows couldn't save %d decoded filenames", dstCreated/2)
	}
	t.Logf("dst %v encoder synced %d files (%d regular and %d %v filenames)%v%v",
		dstEncoder, got, synced, synced, dstTypeMap[dstType], suffix, suffix2)

	cleanup([]string{srcDir, dstDir})
}

// contextWithDeadline returns the context and cancel functions with a deadline
// that ensures no test will panic if the deadline is reached.
func contextWithDeadline(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()

	dl, _ := t.Deadline()
	deadline := maxDeadline(dl)
	average := time.Since(startTime) / time.Duration(testNumber-skippedTests)
	t.Logf("Test %d of %d: Timeout in %v (total remaining %v) (%v average per test)",
		testNumber, totalTests, time.Until(deadline).Truncate(time.Second),
		time.Until(dl).Truncate(time.Second), average.Truncate(time.Second))
	return context.WithDeadline(context.Background(), deadline)
}

// maxDeadline sets the deadline for a single test to either
// maxSecondsPerTest, or the time left until the testing deadline, whichever is
// less.
func maxDeadline(deadline time.Time) time.Time {
	now := time.Now()
	if deadline.Sub(now).Seconds() < float64(maxSecondsPerTest) {
		// Cause a context deadline timeout to occur before the test deadline is reached.
		deadline = deadline.Add(-time.Second * time.Duration(secondsBeforePanic))
		return deadline
	}
	newDeadline := now.Add(time.Second * time.Duration(maxSecondsPerTest))
	if newDeadline.After(deadline) {
		newDeadline = newDeadline.Add(-newDeadline.Sub(deadline))
		// Cause a context deadline timeout to occur before the test deadline is reached.
		newDeadline = newDeadline.Add(-time.Second * time.Duration(secondsBeforePanic))
	}

	return newDeadline
}

// testEncoderSyncTwoDevicesFolders is patterned after the
// testSyncTwoDevicesFolders function in sync_2dev_test.go.
func testEncoderSyncTwoDevicesFolders(ctx context.Context, t *testing.T, srcDir, dstDir string, srcEncoderType, dstEncoderType fs.EncoderType) error {
	t.Helper()

	// The folder needs an ID.
	folderID := rand.String(8)

	// Start the source device.
	src := startInstance(t)
	srcAPI := rc.NewAPI(src.apiAddress, src.apiKey)

	// Start the destination device.
	dst := startInstance(t)
	dstAPI := rc.NewAPI(dst.apiAddress, dst.apiKey)

	// Add the peer device to each device. Hard code the sync addresses to
	// speed things up.
	if err := srcAPI.Post("/rest/config/devices", &config.DeviceConfiguration{
		DeviceID:  dst.deviceID,
		Addresses: []string{fmt.Sprintf("tcp://127.0.0.1:%d", dst.tcpPort)},
	}, nil); err != nil {
		t.Fatal(err)
	}
	if err := dstAPI.Post("/rest/config/devices", &config.DeviceConfiguration{
		DeviceID:  src.deviceID,
		Addresses: []string{fmt.Sprintf("tcp://127.0.0.1:%d", src.tcpPort)},
	}, nil); err != nil {
		t.Fatal(err)
	}

	// Add the folder to both devices.
	if err := srcAPI.Post("/rest/config/folders", &config.FolderConfiguration{
		ID:             folderID,
		Path:           srcDir,
		FilesystemType: fs.FilesystemTypeBasic,
		Type:           config.FolderTypeSendReceive,
		PullerPauseS:   1, // speed up testing by retrying every second
		EncoderType:    srcEncoderType,
		Devices: []config.FolderDeviceConfiguration{
			{DeviceID: src.deviceID},
			{DeviceID: dst.deviceID},
		},
	}, nil); err != nil {
		t.Fatal(err)
	}
	if err := dstAPI.Post("/rest/config/folders", &config.FolderConfiguration{
		ID:             folderID,
		Path:           dstDir,
		FilesystemType: fs.FilesystemTypeBasic,
		Type:           config.FolderTypeSendReceive,
		PullerPauseS:   1,
		EncoderType:    dstEncoderType,
		Devices: []config.FolderDeviceConfiguration{
			{DeviceID: src.deviceID},
			{DeviceID: dst.deviceID},
		},
	}, nil); err != nil {
		t.Fatal(err)
	}

	// Listen to events; we want to know when the folder is fully synced. We
	// consider the other side in sync when we've received an index update
	// from them and subsequently a completion event with percentage equal
	// to 100. At that point they should be done. Wait for both sides to do
	// their thing.

	var srcDur, dstDur map[string]time.Duration
	var srcErr, dstErr error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		srcDur, srcErr = waitForSync(ctx, folderID, srcAPI)
	}()
	go func() {
		defer wg.Done()
		dstDur, dstErr = waitForSync(ctx, folderID, dstAPI)
	}()
	wg.Wait()

	if srcErr != nil && !errors.Is(srcErr, context.DeadlineExceeded) {
		return srcErr
	}

	if dstErr != nil && !errors.Is(dstErr, context.DeadlineExceeded) {
		return dstErr
	}

	t.Log("src durations:", srcDur)
	t.Log("dst durations:", dstDur)

	return nil
}

// srcPrefixes returns a string of filename prefix characters for the specified
// srcType.
func srcPrefixes(srcType srcType) string {
	switch srcType {
	case srcTypeDecoded:
		return fatDecodes()
	case srcTypeEncoded:
		return fatEncodes()
	}
	panic(fmt.Sprintf("bug: unexpected srcType %v", srcType))
}

// dstPrefixes returns a string of filename prefix characters for the specified
// dstType.
func dstPrefixes(dstType dstType) string {
	switch dstType {
	case dstTypeDecoded:
		return fatDecodes()
	case dstTypeEncoded, dstTypeRejectEncoded:
		return fatEncodes()
	case dstTypeSkipped:
	}
	panic(fmt.Sprintf("bug: unexpected dstType %v", dstType))
}

// fatDecodes returns a string where 50% of the characters are encodable
// by the FAT encoder.
func fatDecodes() string {
	chars := ""
	index := 0
	for _, r := range consts.Encodes {
		// The tests pass with unprintable characters, but why bother?
		if unicode.IsControl(r) {
			continue
		}
		// Avoid Alternate Data Streams
		if build.IsWindows && r == ':' {
			continue
		}
		chars += string(rune(alnum[index%len(alnum)]))
		index++
		chars += string(r)
	}

	return chars
}

// fatEncodes returns a string where 50% of the characters have been encoded
// by the FAT encoder.
func fatEncodes() string {
	chars := ""
	index := 0
	for _, r := range consts.Encodes {
		if unicode.IsControl(r) {
			continue
		}
		if build.IsWindows && r == ':' {
			continue
		}
		chars += string(rune(alnum[index%len(alnum)]))
		index++
		chars += string(r | consts.BaseRune)
	}

	return chars
}

// wanted returns the number of filenames we want to find for the given dstType.
func wanted(dstType dstType, srcCreated, dstCreated int) int {
	var want int
	switch dstType {
	case dstTypeDecoded, dstTypeEncoded:
		want = srcCreated + dstCreated
	case dstTypeRejectEncoded:
		// The encoded filenames generated on the src None encoder instance will
		// be rejected by the dst Fat encoder, so we cut srcCreated in half, as
		// only half the filenames it generated were encoded.
		want = (srcCreated / 2) + dstCreated
		// On Windows, the dst Fat encoder will send decoded filenames over the
		// wire, but the src None encoder will reject them as its underlying
		// FAT filesystem rejects decoded filenames, so we have to subtract
		// those (expected) write failures. Again, we cut the result in half,
		// as only half the filenames generated were encoded.
		if build.IsWindows {
			want -= dstCreated / 2
		}
	case dstTypeSkipped:
		want = 0
	}

	return want
}

// skipSubTests skips all subtests for a specific testResultMatrix entry.
func skipSubTests(t *testing.T, msg, extra string) {
	t.Helper()

	for i := 0; i < numberOfSubTests-1; i++ {
		testNumber++
		skippedTests++
		t.Logf(msg, testNumber, totalTests, extra)
	}
	testNumber++
	skippedTests++
	t.Skipf(msg, testNumber, totalTests, extra)
}

// title upper cases the first letter of s. We use it instead of
// strings.Title() as that function is marked as deprecated.
func title(s string) string {
	return cases.Title(language.English).String(s)
}
