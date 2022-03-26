package main

import (
	"fmt"
	"github.com/ImSingee/go-ex/exjson"
	"github.com/rclone/rclone/lib/terminal"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var flagCoursePath string
var flagOffline bool
var flagUpdateMeta bool

var cmdVerify = &cobra.Command{
	Use:     "verify <course-id> ...",
	Aliases: []string{"v"},
	Short:   "Check course's integrity",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no course specified")
		}

		if flagCoursePath != "" && len(args) > 1 {
			return fmt.Errorf("cannot specify course path and more than one course id")
		}

		totalCount := len(args)
		errorCount := 0
		bypassCount := 0

		for _, courseId := range args {
			state := verify(courseId, flagCoursePath, flagSkipFFMpeg, flagOffline, flagUpdateMeta)
			switch state {
			case 0: // success
			case 1: // error
				errorCount++
				fmt.Printf("Course ID %s verified fail\n", courseId)
			case 2:
				bypassCount++
				fmt.Printf("Course ID %s bypassed\n", courseId)
			}
		}

		if errorCount == 0 {
			if bypassCount == 0 {
				fmt.Printf("All %d courses verified successfully\n", totalCount)
			} else {
				fmt.Printf("All %d courses verified successfully, %d courses bypassed\n", totalCount-bypassCount, bypassCount)
			}
		} else {
			if bypassCount == 0 {
				redPrintf("%d courses verified failed\n", errorCount)
			} else {
				redPrintf("%d courses verified failed, %d courses bypassed\n", errorCount, bypassCount)
			}
		}

		return nil
	},
}

// verify 返回值
// 0 - 成功
// 1 - 失败
// 2 - 忽略
func verify(courseId string, courseDir string, noConvert, offline, updateMeta bool) int {
	terminal.Start()

	courseName, ok := GetName(courseId)
	if !ok {
		redPrintf("Unknown course id %s, please register first\n", courseId)
		return 1
	}

	bluePrintf(">>> Verify %s (%s)\n", courseName, courseId)

	if courseDir == "" {
		courseDir = filepath.Join(config.DownloadTo, cleanName(courseName))

		legacyCourseDir := filepath.Join(config.DownloadTo, courseName)
		if courseDir != legacyCourseDir {
			_ = os.Rename(legacyCourseDir, courseDir)
		}
	}

	if !isExist(courseDir) {
		redPrintf("Course path %s not exist\n", courseDir)
		return 1
	}

	metaDir := filepath.Join(courseDir, ".meta")
	_ = os.MkdirAll(metaDir, 0755)

	// 下载最新的课程信息
	var courseLectures *CourseLectures
	var courseInfo *CourseInfo
	var err error
	if offline {
		courseLectures = &CourseLectures{}
		courseInfo = &CourseInfo{}

		lecturesMetaPath := filepath.Join(metaDir, "lectures.json")
		infoMetaPath := filepath.Join(metaDir, "info.json")

		err = exjson.Read(lecturesMetaPath, &courseLectures.Chapters)
		if err != nil {
			redPrintf("Cannot load lectures meta file %s: %v\n", lecturesMetaPath, err)
			return 1
		}

		err = exjson.Read(infoMetaPath, courseInfo)
		if err != nil {
			redPrintf("Cannot load course info meta file %s: %v\n", infoMetaPath, err)
			return 1
		}
	} else {
		courseLectures, err = apiGetWanmenCourseLectures(courseId)
		if err != nil {
			redPrintf("Failed to get course lectures: %s\n", err)
			return 1
		}

		courseInfo, err = apiGetWanmenCourseInfo(courseId)
		if err != nil {
			redPrintf("Failed to get course info: %s\n", err)
			return 1
		}

		if updateMeta { // 更新存储最新的 meta
			lecturesMetaPath := filepath.Join(metaDir, "lectures.json")
			lecturesMetaPathOld := filepath.Join(metaDir, "lectures.json.bak")
			infoMetaPath := filepath.Join(metaDir, "info.json")
			infoMetaPathOld := filepath.Join(metaDir, "info.json.bak")

			_ = os.Rename(lecturesMetaPath, lecturesMetaPathOld)
			_ = os.Rename(infoMetaPath, infoMetaPathOld)
			_ = os.WriteFile(lecturesMetaPath, courseLectures.Raw, 0644)
			_ = os.WriteFile(infoMetaPath, courseInfo.Raw, 0644)
		}
	}

	pass := true

	// 检查课程
	for i, chapter := range courseLectures.Chapters {
		chapterdir := filepath.Join(courseDir, fmt.Sprintf("%d - %s", i+1, cleanName(chapter.Name)))

		// fix oldChapterDir
		oldChapterDir := filepath.Join(courseDir, fmt.Sprintf("%d - %s", i+1, oldCleanName(chapter.Name)))
		if oldChapterDir != chapterdir {
			_ = os.Rename(oldChapterDir, chapterdir)
		}

		for j, lecture := range chapter.Children {
			lecturePath := filepath.Join(chapterdir, fmt.Sprintf("%d-%d %s.mp4", i+1, j+1, cleanName(lecture.Name)))
			lecturePartDonePath := lecturePath + ".stream.mp4"

			// fix oldLecturePath
			oldLecturePath := filepath.Join(chapterdir, fmt.Sprintf("%d-%d %s.mp4", i+1, j+1, oldCleanName(lecture.Name)))
			if lecturePath != oldLecturePath {
				_ = os.Rename(oldLecturePath, lecturePath)
				oldLecturePartDonePath := oldLecturePath + ".stream.mp4"
				_ = os.Rename(oldLecturePartDonePath, lecturePartDonePath)
			}

			if noConvert {
				if !isExist(lecturePath) && !isExist(lecturePartDonePath) {
					redPrintf("Lecture %s not exist\n", lecturePartDonePath)
					pass = false
				}
			} else {
				if !isExist(lecturePath) {
					redPrintf("Lecture %s not exist\n", lecturePath)
					pass = false
				}
			}
		}
	}

	// 检查文档
	for i, doc := range courseInfo.Documents {
		// 万门的某些课程 ext 会出现两次
		originDocName := doc.Name
		doc.Name = strings.TrimSuffix(doc.Name, "."+doc.Ext)
		docPath := filepath.Join(courseDir, "资料", cleanName(fmt.Sprintf("%d - %s.%s", i+1, doc.Name, doc.Ext)))

		// 修正 doc 名称
		oldDocPath := filepath.Join(courseDir, "资料", oldCleanName(fmt.Sprintf("%d - %s.%s", i+1, doc.Name, doc.Ext)))
		if oldDocPath != docPath {
			_ = os.Rename(oldDocPath, docPath)
		}

		// 修正之前 ext 可能两次的名称
		veryOldDocPath := filepath.Join(courseDir, "资料", oldCleanName(fmt.Sprintf("%d - %s.%s", i+1, originDocName, doc.Ext)))
		if veryOldDocPath != docPath {
			if isExist(docPath) {
				_ = os.Remove(veryOldDocPath)
			} else {
				_ = os.Rename(veryOldDocPath, docPath)
			}
		}

		if !isExist(docPath) {
			redPrintf("Document %s not exist\n", docPath)
			pass = false
		}
	}

	// 检查 DONE
	donePath := filepath.Join(metaDir, "DONE")
	donePathLegacy := filepath.Join(courseDir, ".done")
	forceDonePath := filepath.Join(metaDir, "FORCE-DONE")
	if isExist(donePathLegacy) {
		if isExist(donePath) {
			_ = os.Remove(donePathLegacy)
		} else {
			_ = os.Rename(donePathLegacy, donePath)
		}
	}

	if pass {
		// 应该存在 DONE，不存在则添加
		if !isExist(donePath) {
			_ = os.WriteFile(donePath, []byte(time.Now().Format(time.RFC3339)), 0644)
		}

		return 0
	} else {
		// 应该不存在 DONE，存在则删除
		if isExist(donePath) {
			if isExist(forceDonePath) {
				fmt.Println("DONE marker is NOT removed due to force-done")
				fmt.Println("Run following commands to remove the flag")
				fmt.Println("> rm", donePath)
				fmt.Println("> rm", forceDonePath)
			} else {
				fmt.Println("DONE marker is removed from course automatically (due to not-pass-verify)")
				_ = os.Remove(donePath)
			}
		}

		if isExist(forceDonePath) {
			return 2
		} else {
			return 1
		}
	}
}

func init() {
	app.AddCommand(cmdVerify)

	cmdVerify.Flags().BoolVarP(&flagSkipFFMpeg, "skip-ffmpeg", "m", false, "")
	cmdVerify.Flags().BoolVarP(&flagOffline, "offline", "o", false, "offline mode (won't request wanmen api again)")
	cmdVerify.Flags().BoolVar(&flagUpdateMeta, "update-meta", true, "also update exist meta")
	cmdVerify.Flags().StringVarP(&flagCoursePath, "path", "p", "", "course path")
}
