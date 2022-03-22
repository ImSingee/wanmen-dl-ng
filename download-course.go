package main

import "errors"

func fetch_course(course_id, course_name, base_dir string) {

}

// target 为下载目标的绝对路径
// full 为 true 代表不会去除首尾万门的「广告」
// f 用来返回当前的下载进度
// 返回值的第一个参数代表下载状况，0 代表正常，1-4 代表下载到了非超清版本，-1 代表无法下载
// 当且仅当第一个返回值为 -1 时，会带有 error 参数
func downloadLecture(info *LectureInfo, target string, full bool, f updateProgressFunc) (int, error) {
	var latestError error
	for i, url := range info.VideoStream.ToDownload() {
		err := tryDownloadLectureM3U8(url, target, full, f)
		if err == nil {
			return i, nil
		}
		latestError = err
	}

	if latestError == nil {
		return -1, errors.New("no url can be downloaded")
	} else {
		return -1, latestError
	}
}
