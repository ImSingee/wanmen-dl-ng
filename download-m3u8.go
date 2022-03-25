package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"path/filepath"
)

var cmdDownloadM3U8 = &cobra.Command{
	Use: "download-m3u8",
	Run: func(cmd *cobra.Command, args []string) {
		for _, task := range args {
			fmt.Println("start", task)

			// stream done to
			doneTo := task + ".stream.mp4"

			if isExist(doneTo) {
				fmt.Println("skip-task", task)
				continue
			}

			f := func(a string, v ...interface{}) {
				fmt.Println("lecture", task, a, v)
			}

			err := downloadSosM3U8(task, doneTo, false, f)

			if err != nil {
				fmt.Println("error", task, err)
				continue
			}

			fmt.Println("done", task)
		}

		for _, courseId := range args {
			name, ok := GetName(courseId)
			if !ok {
				fmt.Println("cannot find course name for", courseId)
				continue
			}

			sosPath := filepath.Join(config.SosDir, sosCleanName(name))

			fmt.Println("Download", courseId, sosPath)
			sosDownload(sosPath, 0)
			fmt.Println("Download", courseId, "DONE", sosPath)
		}
	},
}

func init() {
	app.AddCommand(cmdDownloadM3U8)
}
