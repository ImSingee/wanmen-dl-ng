package main

import (
	"fmt"
	"github.com/ImSingee/tt"
	"testing"
)

func TestDownloadM3U8(t *testing.T) {
	info, err := apiGetWanmenLectureInfo("5aded0c1b6917f44d5121710")
	tt.AssertIsNotError(t, err)

	target := "/tmp/wanmen-dl-ng-test-download-m3u8.mp4"

	updateProgress := func(state string, params ...interface{}) {
		fmt.Printf("%v %v\n", state, params)
	}

	code, err := downloadLectureM3U8(info, target, false, updateProgress)
	tt.AssertIsNotError(t, err)
	tt.AssertEqual(t, code, 0)
}
