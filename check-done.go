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
	RunE: func(cmd *cobra.Command, args []string) error {
		courseId := args[0]
		courseName, ok := GetName(courseId)

		if !ok {
			fmt.Printf("%-30s %-10s    %s\n", courseId, "PREPARE", "UNKNOWN")
			return nil
		}

		d := path.Join(config.DownloadTo, courseName)

		f1 := path.Join(d, ".done")
		f2 := path.Join(d, ".meta", "DONE")

		if isExist(f1) || isExist(f2) {
			fmt.Printf("%-30s %-10s %s\n", courseId, "DONE", courseName)
			return nil
		}

		if isExist(d) {
			fmt.Printf("%-30s %-10s %s\n", courseId, "DOWNLOADING", courseName)
			return nil
		}

		fmt.Printf("%-30s %-10s %s\n", courseId, "PREPARE", courseName)
		return nil
	},
}

func init() {
	app.AddCommand(cmdCheck)
}
