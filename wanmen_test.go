package main

import (
	"github.com/ImSingee/tt"
	"testing"
)

func TestGetWanmenLectureInfo(t *testing.T) {
	info, err := apiGetWanmenLectureInfo("5aded0c1b6917f44d5121710")
	tt.AssertIsNotError(t, err)
	tt.AssertNotEqual(t, info.Name, "")
	tt.AssertIsNotNil(t, info.VideoStream)
	tt.AssertNotEqual(t, info.VideoStream.PcHigh, "")
	tt.AssertNotEqual(t, info.VideoStream.PcMid, "")
	tt.AssertNotEqual(t, info.VideoStream.PcLow, "")
	tt.AssertNotEqual(t, info.VideoStream.MobileMid, "")
	tt.AssertNotEqual(t, info.VideoStream.MobileLow, "")
	tt.AssertNotEqual(t, len(info.RawJsonBody), 0)
}

func TestGetWanmenCourseInfo(t *testing.T) {
	_, err := apiGetWanmenCourseInfo("59df20a60dcf357a8bc0000c")
	tt.AssertIsNotError(t, err)
}
