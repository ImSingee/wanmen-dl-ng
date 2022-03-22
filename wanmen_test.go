package main

import (
	"github.com/ImSingee/tt"
	"testing"
)

func TestGetWanmenLectureInfo(t *testing.T) {
	info, err := apiGetWanmenLectureInfo("5aded0c1b6917f44d5121710")
	tt.AssertIsNotError(t, err)
	tt.AssertIsNotNil(t, info.VideoStream)
	tt.AssertNotEqual(t, info.VideoStream.PcHigh, "")
	tt.AssertNotEqual(t, info.VideoStream.PcMid, "")
	tt.AssertNotEqual(t, info.VideoStream.PcLow, "")
}
