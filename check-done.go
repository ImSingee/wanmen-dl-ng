package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"path"
)

var cmdCheck = &cobra.Command{
	Use: "check <course-id>",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("requires course-id")
		} else if len(args) > 1 {
			return fmt.Errorf("usage: check <course-id>")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		courseId := args[0]
		checkDone(courseId)
	},
}

func checkDone(courseId string) {
	courseName, ok := GetName(courseId)

	const tmpl = "%-30s %-10s    %s\n"

	if !ok {
		fmt.Printf(tmpl, courseId, "PREPARE", "UNKNOWN")
		return
	}

	d := path.Join(config.DownloadTo, courseName)

	f1 := path.Join(d, ".done")
	f2 := path.Join(d, ".meta", "DONE")

	if isExist(f1) || isExist(f2) {
		fmt.Printf(tmpl, courseId, "DONE", courseName)
		return
	}

	if isExist(d) {
		fmt.Printf(tmpl, courseId, "DOWNLOADING", courseName)
		return
	}

	fmt.Printf(tmpl, courseId, "PREPARE", courseName)
}

func init() {
	app.AddCommand(cmdCheck)
}
