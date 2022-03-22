package main

type DownloadTask struct {
	MetaDir         string
	Chapter         *CourseInfo_Chapter
	ChapterDir      string
	Lecture         *CourseInfo_Lecture
	LecturePath     string
	ForceReDownload bool
}
