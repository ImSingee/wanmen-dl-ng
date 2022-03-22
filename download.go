package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"path"
	"runtime"
)

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
		courseName, ok := GetName(courseId)
		if !ok {
			return errors.New("unknown course, please register first")
		}

		dashboard := NewDashboard(courseId, courseName, flagConcurrency)

		actionHandler := dashboard.Start()
		defer dashboard.Close()

		downloadTo := flagDownloadTo
		if downloadTo == "" {
			downloadTo = path.Join(config.DownloadTo, courseName)
		}

		updateProgress := func(state string, params ...interface{}) {
			actionHandler <- DashboardAction{state, params}
		}

		return DownloadCourse(courseId, downloadTo, flagFull, flagConcurrency, updateProgress)
	},
}

func init() {
	app.AddCommand(cmdDownload)

	cmdDownload.Flags().BoolVar(&flagFull, "full", false, "不去除万门广告")
	cmdDownload.Flags().StringVarP(&flagDownloadTo, "to", "t", "", "下载到的路径，留空代表配置中的路径+课程名称")
	cmdDownload.Flags().IntVarP(&flagConcurrency, "concurrency", "c", runtime.NumCPU()*4, "并发数，默认为 CPU 数量 * 4")
}
