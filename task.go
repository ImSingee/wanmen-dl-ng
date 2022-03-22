package main

import (
	"fmt"
	"path"
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
		return "" // TODO
	}
}

func (task *DownloadTask) MetaPrefix() string {
	if task.Course != nil {
		return path.Join(task.MetaDir, fmt.Sprintf("%s:%s", task.Course.Chapter.ID, task.Course.Lecture.ID))
	} else {
		return "" // TODO
	}
}

func (task *DownloadTask) Desc() string {
	if task.Course != nil {
		td := task.Course
		return fmt.Sprintf("Ch%d/%d-%s", td.Chapter.Index, td.Lecture.Index, td.Lecture.Name)
	} else {
		return "" // TODO
	}
}

type CourseDownloadTask struct {
	Chapter     *CourseInfo_Chapter
	ChapterDir  string
	Lecture     *CourseInfo_Lecture
	LecturePath string
}

type DocDownloadTask struct {
}
