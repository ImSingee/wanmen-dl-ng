package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"path/filepath"
	"runtime"
)

var flagSkipFFMpeg bool
var flagForce int
var flagConcurrency int
var flagFull bool
var flagDownloadTo string

var cmdDownload = &cobra.Command{
	Use: "download <course-id>",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("requires course-id")
		} else if len(args) > 1 {
			return fmt.Errorf("usage: download <course-id>")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		courseId := args[0]
		return download(courseId, flagDownloadTo, flagForce, flagFull, flagSkipFFMpeg, flagConcurrency)
	},
}

func download(courseId string, downloadTo string, forceLevel int, full bool, noConvert bool, concurrency int) error {
	if !noConvert && ffmpegPath == "" {
		return errors.New("ffmpeg is not installed")
	}

	courseName, ok := GetName(courseId)
	if !ok {
		return errors.New("unknown course, please register first")
	}

	dashboard := NewDashboard(courseId, courseName, concurrency)

	actionHandler := dashboard.Start()
	defer dashboard.Close()

	if downloadTo == "" {
		downloadTo = filepath.Join(config.DownloadTo, courseName)
	}

	updateProgress := func(state string, params ...interface{}) {
		actionHandler <- DashboardAction{state, params}
	}

	return DownloadCourse(courseId, downloadTo, forceLevel, full, concurrency, noConvert, updateProgress)
}

func init() {
	app.AddCommand(cmdDownload)

	cmdDownload.Flags().BoolVarP(&flagSkipFFMpeg, "skip-ffmpeg", "m", false, "")
	cmdDownload.Flags().IntVarP(&flagForce, "force", "f", 0, "跳过去重（0-不跳过, 1-跳过课程检测, 2-跳过文件检测)")
	cmdDownload.Flags().BoolVar(&flagFull, "full", false, "不去除万门广告")
	cmdDownload.Flags().StringVarP(&flagDownloadTo, "to", "t", "", "下载到的路径，留空代表配置中的路径+课程名称")
	cmdDownload.Flags().IntVarP(&flagConcurrency, "concurrency", "c", runtime.NumCPU()*4, "并发数，默认为 CPU 数量 * 4")
}
