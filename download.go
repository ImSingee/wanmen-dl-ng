package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"runtime"
)

var flagSos bool
var flagSkipFFMpeg bool
var flagForce int
var flagConcurrency int
var flagFull bool
var flagDownloadTo string

var cmdDownload = &cobra.Command{
	Use:     "download <course-id>",
	Aliases: []string{"d"},
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

		if flagSos {
			return sosDownload(courseId, flagDownloadTo, flagForce, flagFull, flagSkipFFMpeg, flagConcurrency)
		} else {
			return download(courseId, flagDownloadTo, flagForce, flagFull, flagSkipFFMpeg, flagOffline, flagConcurrency)
		}

	},
}

func download(courseId string, downloadTo string, forceLevel int, full bool, noConvert bool, offline bool, concurrency int) error {
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
		downloadTo = filepath.Join(config.DownloadTo, cleanName(courseName))

		legacyDownloadTo := filepath.Join(config.DownloadTo, courseName)
		if downloadTo != legacyDownloadTo {
			_ = os.Rename(legacyDownloadTo, downloadTo)
		}
	}

	updateProgress := func(state string, params ...interface{}) {
		actionHandler <- DashboardAction{state, params}
	}

	return DownloadCourse(courseId, downloadTo, forceLevel, full, concurrency, noConvert, offline, updateProgress)
}

func sosDownload(courseId string, downloadTo string, forceLevel int, full bool, noConvert bool, concurrency int) error {
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
		downloadTo = filepath.Join(config.DownloadTo, cleanName(courseName))

		legacyDownloadTo := filepath.Join(config.DownloadTo, courseName)
		if downloadTo != legacyDownloadTo {
			_ = os.Rename(legacyDownloadTo, downloadTo)
		}
	}

	updateProgress := func(state string, params ...interface{}) {
		actionHandler <- DashboardAction{state, params}
	}

	return SosDownloadCourse(courseId, downloadTo, forceLevel, full, concurrency, noConvert, updateProgress)
}

func init() {
	app.AddCommand(cmdDownload)

	cmdDownload.Flags().BoolVar(&flagSos, "sos", false, "enable sos mode")
	cmdDownload.Flags().BoolVar(&flagOffline, "offline", false, "")
	cmdDownload.Flags().BoolVarP(&flagSkipFFMpeg, "skip-ffmpeg", "m", false, "")
	cmdDownload.Flags().IntVarP(&flagForce, "force", "f", 0, "???????????????0-?????????, 1-??????????????????, 2-??????????????????)")
	cmdDownload.Flags().BoolVar(&flagFull, "full", false, "?????????????????????")
	cmdDownload.Flags().StringVarP(&flagDownloadTo, "to", "t", "", "???????????????????????????????????????????????????+????????????")
	cmdDownload.Flags().IntVarP(&flagConcurrency, "concurrency", "c", runtime.NumCPU()*4, "????????????????????? CPU ?????? * 4")
}
