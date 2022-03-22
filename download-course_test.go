package main

import (
	"fmt"
	"github.com/ImSingee/tt"
	"testing"
)

func TestDownloadCourse(t *testing.T) {
	updateProgress := func(state string, params ...interface{}) {
		fmt.Printf("%v %v\n", state, params)
	}

	err := DownloadCourse("6182392abc669300bdd6bc89", "/tmp/wanmen-dl-test-download-course", false, 0, updateProgress)
	tt.AssertIsNotError(t, err)
}

func TestDownloadLecture(t *testing.T) {
	target := "/tmp/wanmen-dl-ng-test-download-m3u8.mp4"
	metaPath := "/tmp/wanmen-dl-ng-test-download-m3u8-meta.json"

	updateProgress := func(state string, params ...interface{}) {
		fmt.Printf("%v %v\n", state, params)
	}

	code, err := downloadLecture("5aded0c1b6917f44d5121710", target, metaPath, false, updateProgress)
	tt.AssertIsNotError(t, err)
	tt.AssertEqual(t, code, 0)
}
