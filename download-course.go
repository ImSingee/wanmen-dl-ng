package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"sync"
)

type toDownload struct {
	Chapter     *CourseInfo_Chapter
	ChapterDir  string
	Lecture     *CourseInfo_Lecture
	LecturePath string
}

type updateCourseProgressFunc func(state string, params ...interface{})

func downloadCourse(courseId, courseDir string, full bool, concurrency int, updateProgress updateCourseProgressFunc) error {
	courseInfo, err := apiGetWanmenCourseInfo(courseId)
	if err != nil {
		return fmt.Errorf("apiGetWanmenCourseInfo error: %v", err)
	}

	updateProgress("fetch_course_info", courseInfo)

	wg := sync.WaitGroup{}

	// 生成下载队列
	queue := make(chan *toDownload, 64)
	wg.Add(1)
	go func() {
		defer wg.Done()

		for i, chapter := range courseInfo {
			chapterdir := path.Join(courseDir, fmt.Sprintf("%d - %s", i+1, cleanName(chapter.Name)))

			for j, lecture := range chapter.Children {
				queue <- &toDownload{
					Chapter:     chapter,
					ChapterDir:  chapterdir,
					Lecture:     lecture,
					LecturePath: path.Join(chapterdir, fmt.Sprintf("%d - %s", j+1, cleanName(lecture.Name))),
				}
			}
		}

		close(queue)
	}()

	// 默认并发数为当前系统 CPU 核数 * 4
	if concurrency <= 0 {
		concurrency = runtime.NumCPU() * 4
	}

	wg.Add(concurrency)
	// 开始下载
	for i := 0; i < concurrency; i++ {
		go func(workerId int) {
			defer wg.Done()

			for {
				select {
				case toDownload := <-queue: // 从队列中取出一个下载任务
					updateProgress("start", workerId, toDownload)

					f := func(a string, v ...interface{}) {
						updateProgress("sub", workerId, a, v)
					}

					_, err := downloadLecture(toDownload.Lecture.ID, toDownload.LecturePath, full, f)
					if err != nil {
						updateProgress("error", workerId, toDownload, err)
						continue
					}

					updateProgress("done", workerId, toDownload)
				default:
					return
				}
			}
		}(i + 1)
	}

	wg.Done()
	return nil
}

type updateLectureProgressFunc func(state string, params ...interface{})

// target 为下载目标的绝对路径
// full 为 true 代表不会去除首尾万门的「广告」
// f 用来返回当前的下载进度
// 返回值的第一个参数代表下载状况，0 代表正常，1-4 代表下载到了非超清版本，-1 代表无法下载
// 当且仅当第一个返回值为 -1 时，会带有 error 参数
func downloadLecture(lectureID string, lecturePath string, full bool, f updateLectureProgressFunc) (int, error) {
	info, err := apiGetWanmenLectureInfo(lectureID)
	if err != nil {
		return -1, fmt.Errorf("cannot get lecture info: %v", err)
	}

	_ = os.MkdirAll(path.Dir(lecturePath), 0755)

	target := lecturePath

	var latestError error
	for i, url := range info.VideoStream.ToDownload() {
		err := tryDownloadLectureM3U8(url, target, full, f)
		if err == nil {
			return i, nil
		}
		f("retry", i, err)
		latestError = err
	}

	if latestError == nil {
		return -1, errors.New("no url can be downloaded")
	} else {
		return -1, latestError
	}
}
