package main

import (
	"fmt"
	"path/filepath"
)

type DownloadTask struct { // 多选一
	MetaDir         string
	ForceReDownload bool

	Course *CourseDownloadTask
	Doc    *DocDownloadTask
}

func (task *DownloadTask) Path() string {
	if task.Course != nil {
		return task.Course.LecturePath
	} else {
		return task.Doc.DocumentPath
	}
}

func (task *DownloadTask) MetaPrefix() string {
	if task.Course != nil {
		return filepath.Join(task.MetaDir, fmt.Sprintf("%s-%s", task.Course.Chapter.ID, task.Course.Lecture.ID))
	} else {
		return filepath.Join(task.MetaDir, fmt.Sprintf("doc-%s", task.Doc.Document.Key))
	}
}

func (task *DownloadTask) Desc() string {
	if task.Course != nil {
		td := task.Course
		return fmt.Sprintf("Ch%d/%d-%s", td.Chapter.Index, td.Lecture.Index, td.Lecture.Name)
	} else {
		return fmt.Sprintf("Doc%d-%s", task.Doc.Document.Index, task.Doc.Document.Name)
	}
}

type CourseDownloadTask struct {
	Chapter     *CourseInfo_Chapter
	ChapterDir  string
	Lecture     *CourseInfo_Lecture
	LecturePath string

	// 以下只有 sos 模式会用到
	SosLecturePath string
}

type DocDownloadTask struct {
	Document     *CourseInfo_Document
	DocumentPath string
}
