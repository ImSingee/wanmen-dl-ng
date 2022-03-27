package main

import (
	"encoding/json"
	"fmt"
	"github.com/ImSingee/go-ex/exjson"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

func SosDownloadCourse(courseId, courseDir string, forceLevel int, full bool, concurrency int, noConvert bool, updateProgress updateProgressFunc) error {
	metaDir := filepath.Join(courseDir, ".meta")
	_ = os.MkdirAll(metaDir, 0755)

	// 全课程自动跳过
	if forceLevel == 0 && isExist(filepath.Join(metaDir, "DONE")) || isExist(filepath.Join(courseDir, ".done")) {
		updateProgress("skip")
		return nil
	}

	sosCourseName, ok := getSosName(courseId)
	if !ok {
		return fmt.Errorf("courseId %s not found", courseId)
	}
	sosPath := filepath.Join(config.SosDir, sosCleanName(sosCourseName))
	if !isExist(sosPath) {
		return fmt.Errorf("sosPath %s not found", sosPath)
	}

	var courseLectures *CourseLectures
	var courseInfo *CourseInfo
	var err error

	courseLectures = &CourseLectures{}
	courseInfo = &CourseInfo{}

	lecturesMetaPath := filepath.Join(metaDir, "lectures.json")
	infoMetaPath := filepath.Join(metaDir, "info.json")

	//err = exjson.Read(lecturesMetaPath, &courseLectures.Chapters)
	if true { // err != nil {
		// 利用 sos 路径恢复
		sosLecturesMetaPath := filepath.Join(sosPath, "lectures.json")

		tempmap := map[string]interface{}{}

		err = exjson.Read(sosLecturesMetaPath, &tempmap)
		if err != nil {
			return fmt.Errorf("cannot load lectures sos meta file %s: %v", sosLecturesMetaPath, err)
		}

		if _, ok := tempmap["lectures"]; !ok {
			return fmt.Errorf("cannot load lectures sos meta file %s: no lectures key", sosLecturesMetaPath)
		}

		// remarshal
		origin, err := json.Marshal(tempmap["lectures"])
		if err != nil {
			return fmt.Errorf("cannot re-marshal lectures meta file %s: %v", sosLecturesMetaPath, err)
		}

		// re-unmarshal
		err = json.Unmarshal(origin, &courseLectures.Chapters)
		if err != nil {
			return fmt.Errorf("cannot re-unmarshal lectures meta file %s: %v", sosLecturesMetaPath, err)
		}

		// save if need
		_ = os.WriteFile(filepath.Join(metaDir, "lectures.sos.json"), origin, 0644)
		if !isExist(lecturesMetaPath) {
			_ = os.WriteFile(lecturesMetaPath, origin, 0644)
		}

	}

	//err = exjson.Read(infoMetaPath, courseInfo)
	if true { // err != nil {
		// 利用 sos 路径恢复
		sosInfoPath := filepath.Join(sosPath, "info.json")

		err = exjson.Read(sosInfoPath, &courseInfo)
		if err != nil {
			return fmt.Errorf("cannot load lectures sos meta file %s: %v", sosInfoPath, err)
		}

		// save if need
		_ = CopyFile(sosInfoPath, filepath.Join(metaDir, "info.sos.json"))
		if !isExist(infoMetaPath) {
			_ = CopyFile(sosInfoPath, infoMetaPath)

		}

	}

	updateProgress("init-lectures", courseLectures)

	courseDocuments := courseInfo.Documents
	updateProgress("init-documents", courseDocuments)

	wg := sync.WaitGroup{}

	// 生成下载队列
	queue := make(chan *DownloadTask, 64)
	wg.Add(1)
	go func() {
		defer wg.Done()

		for i, chapter := range courseLectures.Chapters {
			chapter.Index = i + 1
			chapterdir := filepath.Join(courseDir, fmt.Sprintf("%d - %s", i+1, cleanName(chapter.Name)))
			chapterSosDir := filepath.Join(sosPath, fmt.Sprintf("第%s讲_%s", chapter.Prefix, sosCleanName(chapter.Name)))

			for j, lecture := range chapter.Children {
				lecture.Index = j + 1
				queue <- &DownloadTask{
					MetaDir: metaDir,
					Course: &CourseDownloadTask{
						Chapter:        chapter,
						ChapterDir:     chapterdir,
						Lecture:        lecture,
						LecturePath:    filepath.Join(chapterdir, fmt.Sprintf("%d-%d %s.mp4", i+1, j+1, cleanName(lecture.Name))),
						SosLecturePath: filepath.Join(chapterSosDir, fmt.Sprintf("%s_%s", lecture.Prefix, sosCleanName(lecture.Name))),
					},
				}
			}
		}

		for i, doc := range courseDocuments {
			doc.Index = i + 1

			// 万门的某些课程 ext 会出现两次
			doc.Name = strings.TrimSuffix(doc.Name, "."+doc.Ext)

			queue <- &DownloadTask{
				MetaDir: metaDir,
				Doc: &DocDownloadTask{
					Document:     doc,
					DocumentPath: filepath.Join(courseDir, "资料", cleanName(fmt.Sprintf("%d - %s.%s", i+1, doc.Name, doc.Ext))),
				},
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

				if forceLevel != 2 && !task.ForceReDownload {
					if isExist(task.Path()) {
						updateProgress("skip-task", workerId, task)
						continue
					}
					if task.Course != nil {
						partDonePath := task.Course.LecturePath + ".stream.mp4"
						if noConvert && isExist(partDonePath) {
							updateProgress("skip-task", workerId, task)
							continue
						}
					}
				}

				if task.Course != nil {
					f := func(a string, v ...interface{}) {
						updateProgress("lecture", workerId, task, a, v)
					}

					_, err := sosDownloadLecture(task.Course.Lecture.ID, task.Path(), task.Course.SosLecturePath, task.MetaPrefix(), noConvert, full, f)
					if err != nil {
						updateProgress("error", workerId, task, err)
						continue
					}
				} else { // task.Doc
					f := func(a string, v ...interface{}) {
						updateProgress("doc", workerId, task, a, v)
					}

					err := sosDownloadDoc(task.Doc.Document, task.Path(), sosPath, task.MetaPrefix(), f)
					if err != nil {
						updateProgress("error", workerId, task, err)
						continue
					}
				}

				updateProgress("done", workerId, task)
			}
		}(i + 1)
	}

	wg.Wait()

	_ = os.WriteFile(filepath.Join(metaDir, "DONE"), []byte(time.Now().Format(time.RFC3339)), 0644)

	return nil
}
