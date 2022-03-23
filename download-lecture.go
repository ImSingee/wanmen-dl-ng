package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// target 为下载目标的绝对路径
// full 为 true 代表不会去除首尾万门的「广告」
// f 用来返回当前的下载进度
// 返回值的第一个参数代表下载状况，0 代表正常，1-4 代表下载到了非超清版本，-1 代表无法下载
// 当且仅当第一个返回值为 -1 时，会带有 error 参数
func downloadLecture(lectureID string, lecturePath string, metaPrefix string, noConvert bool, full bool, f updateProgressFunc) (int, error) {
	info, err := apiGetWanmenLectureInfo(lectureID)
	if err != nil {
		err = fmt.Errorf("cannot get lecture info: %v", err)
		_ = appendJSON(metaPrefix+".error.jsonl", map[string]interface{}{
			"op":  "fail",
			"err": err.Error(),
		})
		return -1, err
	}

	_ = os.MkdirAll(filepath.Dir(lecturePath), 0755)
	_ = os.WriteFile(metaPrefix+".json", info.RawJsonBody, 0644)

	target := lecturePath

	var latestError error
	for i, url := range info.VideoStream.ToDownload() {
		err := tryDownloadLectureM3U8(url, target, noConvert, full, f)
		if err == nil {
			return i, nil
		}

		f("retry", i, err)
		_ = appendJSON(metaPrefix+".error.jsonl", map[string]interface{}{
			"op":   "retry",
			"mode": i,
			"url":  url,
			"err":  err.Error(),
		})
		latestError = err
	}

	if latestError == nil {
		latestError = errors.New("no url can be downloaded")
	}

	_ = appendJSON(metaPrefix+".error.jsonl", map[string]interface{}{
		"op":  "fail",
		"err": latestError.Error(),
	})
	return -1, latestError

}
