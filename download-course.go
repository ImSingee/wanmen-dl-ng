package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"sync"
	"time"
)

type updateCourseProgressFunc func(state string, params ...interface{})

func DownloadCourse(courseId, courseDir string, full bool, concurrency int, updateProgress updateCourseProgressFunc) error {
	metaDir := path.Join(courseDir, ".meta")
	_ = os.MkdirAll(metaDir, 0755)

	// 开启自动跳过
	if isExist(path.Join(metaDir, "DONE")) || isExist(path.Join(courseDir, ".done")) {
		updateProgress("skip")
		return nil
	}

	courseLectures, err := apiGetWanmenCourseLectures(courseId)
	if err != nil {
		return fmt.Errorf("apiGetWanmenCourseLectures error: %v", err)
	}

	updateProgress("init", courseLectures)

	// 将原始 lectures 信息存储
	_ = os.WriteFile(path.Join(metaDir, "lectures.json"), courseLectures.Raw, 0644)

	wg := sync.WaitGroup{}

	// 生成下载队列
	queue := make(chan *DownloadTask, 64)
	wg.Add(1)
	go func() {
		defer wg.Done()

		for i, chapter := range courseLectures.Chapters {
			chapter.Index = i + 1
			chapterdir := path.Join(courseDir, fmt.Sprintf("%d - %s", i+1, cleanName(chapter.Name)))

			for j, lecture := range chapter.Children {
				lecture.Index = j + 1
				queue <- &DownloadTask{
					MetaDir: metaDir,
					Course: &CourseDownloadTask{
						Chapter:     chapter,
						ChapterDir:  chapterdir,
						Lecture:     lecture,
						LecturePath: path.Join(chapterdir, fmt.Sprintf("%d-%d %s.mp4", i+1, j+1, cleanName(lecture.Name))),
					},
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
			defer updateProgress("quit", workerId)

			for {
				task, ok := <-queue // 从队列中取出一个下载任务
				if !ok {
					break
				}

				updateProgress("start", workerId, task)

				if !task.ForceReDownload && isExist(task.Path()) {
					updateProgress("skip-task", workerId, task)
					continue
				}

				if task.Course != nil {
					f := func(a string, v ...interface{}) {
						updateProgress("lecture", workerId, task, a, v)
					}

					_, err := downloadLecture(task.Course.Lecture.ID, task.Path(), task.MetaPrefix(), full, f)
					if err != nil {
						updateProgress("error", workerId, task, err)
						continue
					}
				} else { // task.Doc
					// TODO
				}

				updateProgress("done", workerId, task)
			}
		}(i + 1)
	}

	wg.Wait()

	_ = os.WriteFile(path.Join(metaDir, "DONE"), []byte(time.Now().Format(time.RFC3339)), 0644)

	return nil
}

type updateLectureProgressFunc func(state string, params ...interface{})

// target 为下载目标的绝对路径
// full 为 true 代表不会去除首尾万门的「广告」
// f 用来返回当前的下载进度
// 返回值的第一个参数代表下载状况，0 代表正常，1-4 代表下载到了非超清版本，-1 代表无法下载
// 当且仅当第一个返回值为 -1 时，会带有 error 参数
func downloadLecture(lectureID string, lecturePath string, metaPrefix string, full bool, f updateLectureProgressFunc) (int, error) {
	info, err := apiGetWanmenLectureInfo(lectureID)
	if err != nil {
		return -1, fmt.Errorf("cannot get lecture info: %v", err)
	}

	_ = os.MkdirAll(path.Dir(lecturePath), 0755)
	_ = os.WriteFile(metaPrefix+".json", info.RawJsonBody, 0644)

	target := lecturePath

	var latestError error
	for i, url := range info.VideoStream.ToDownload() {
		err := tryDownloadLectureM3U8(url, target, full, f)
		if err == nil {
			return i, nil
		}

		f("retry", i, err)
		_ = appendJSON(metaPrefix+".error.jsonl", map[string]interface{}{
			"op":   "retry",
			"mode": i,
			"url":  url,
			"err":  err,
		})
		latestError = err
	}

	if latestError == nil {
		latestError = errors.New("no url can be downloaded")
	}

	_ = appendJSON(metaPrefix+".error.jsonl", map[string]interface{}{
		"op":  "fail",
		"err": latestError,
	})

	return -1, latestError

}
