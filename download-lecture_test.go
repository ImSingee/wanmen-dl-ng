package main

import (
	"fmt"
	"github.com/ImSingee/tt"
	"testing"
)

func TestDownloadLecture(t *testing.T) {
	target := "/tmp/wanmen-dl-ng-test-download-m3u8.mp4"
	metaPrefix := "/tmp/wanmen-dl-ng-test-download-m3u8-meta"

	updateProgress := func(state string, params ...interface{}) {
		fmt.Printf("%v %v\n", state, params)
	}

	code, err := downloadLecture("5aded0c1b6917f44d5121710", target, metaPrefix, false, updateProgress)
	tt.AssertIsNotError(t, err)
	tt.AssertEqual(t, code, 0)
}
