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
