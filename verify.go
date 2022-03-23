package main

import (
	"fmt"
	"github.com/ImSingee/go-ex/exjson"
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
	Use:   "verify",
	Short: "Check course's integrity",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no course specified")
		}

		if flagCoursePath != "" && len(args) > 1 {
			return fmt.Errorf("cannot specify course path and more than one course id")
		}

		anyError := false

		for _, courseId := range args {
			ok := verify(courseId, flagCoursePath, flagOffline, flagUpdateMeta)
			if !ok {
				anyError = true
				fmt.Printf("Course ID %s verified fail\n", courseId)
			}
		}

		if anyError {
			return fmt.Errorf("some errors occurred")
		}

		return nil
	},
}

func verify(courseId string, courseDir string, offline, updateMeta bool) bool {
	courseName, ok := GetName(courseId)
	if !ok {
		fmt.Printf("Unknown course id %s, please register first", courseId)
		return false
	}

	fmt.Printf(">>> Verify %s (%s)\n", courseName, courseId)

	if courseDir == "" {
		courseDir = filepath.Join(config.DownloadTo, courseName)
	}

	if !isExist(courseDir) {
		fmt.Printf("Course path %s not exist\n", courseDir)
		return false
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
			fmt.Printf("Cannot load lectures meta file %s: %v\n", lecturesMetaPath, err)
			return false
		}

		err = exjson.Read(infoMetaPath, courseInfo)
		if err != nil {
			fmt.Printf("Cannot load course info meta file %s: %v\n", infoMetaPath, err)
			return false
		}
	} else {
		courseLectures, err = apiGetWanmenCourseLectures(courseId)
		if err != nil {
			fmt.Printf("Failed to get course lectures: %s\n", err)
			return false
		}

		courseInfo, err = apiGetWanmenCourseInfo(courseId)
		if err != nil {
			fmt.Printf("Failed to get course info: %s\n", err)
			return false
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

		for j, lecture := range chapter.Children {
			lecturePath := filepath.Join(chapterdir, fmt.Sprintf("%d-%d %s.mp4", i+1, j+1, cleanName(lecture.Name)))

			if !isExist(lecturePath) {
				fmt.Printf("Lecture %s not exist\n", lecturePath)
				pass = false
			}
		}
	}

	// 检查文档
	for i, doc := range courseInfo.Documents {
		// 万门的某些课程 ext 会出现两次
		doc.Name = strings.TrimSuffix(doc.Name, "."+doc.Ext)
		docPath := filepath.Join(courseDir, "资料", cleanName(fmt.Sprintf("%d - %s.%s", i+1, doc.Name, doc.Ext)))

		if !isExist(docPath) {
			fmt.Printf("Document %s not exist\n", docPath)
			pass = false
		}
	}

	// 检查 DONE
	donePath := filepath.Join(metaDir, "DONE")
	donePathLegacy := filepath.Join(courseDir, ".done")
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
	} else {
		// 应该不存在 DONE，存在则删除
		if isExist(donePath) {
			fmt.Println("DONE marker is removed from course automatically (due to not-pass-verify)")
			_ = os.Remove(donePath)
		}
	}

	return pass
}

func init() {
	app.AddCommand(cmdVerify)

	cmdVerify.Flags().BoolVarP(&flagOffline, "offline", "o", false, "offline mode (won't request wanmen api again)")
	cmdVerify.Flags().BoolVar(&flagUpdateMeta, "update-meta", true, "also update exist meta")
	cmdVerify.Flags().StringVarP(&flagCoursePath, "path", "p", "", "course path")
}
