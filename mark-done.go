package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

var cmdMarkDone = &cobra.Command{
	Use: "mark-done <course-id>",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no course specified")
		}

		for _, courseId := range args {
			err := markDone(courseId, flagCoursePath)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func markDone(courseId string, courseDir string) error {
	courseName, ok := GetName(courseId)
	if !ok {
		return fmt.Errorf("unknown course %s", courseId)
	}

	if courseDir == "" {
		courseDir = filepath.Join(config.DownloadTo, cleanName(courseName))
	}

	if !isExist(courseDir) {
		return fmt.Errorf("course dir %s not exist", courseDir)
	}

	metaDir := filepath.Join(courseDir, ".meta")
	_ = os.MkdirAll(metaDir, 0755)

	donePath := filepath.Join(metaDir, "DONE")
	forceDonePath := filepath.Join(metaDir, "FORCE-DONE")

	if !isExist(donePath) {
		_ = os.WriteFile(donePath, []byte(time.Now().Format(time.RFC3339)), 0644)
	}
	if !isExist(forceDonePath) {
		_ = os.WriteFile(forceDonePath, []byte(time.Now().Format(time.RFC3339)), 0644)
	}

	return nil
}

func init() {
	app.AddCommand(cmdMarkDone)

	cmdMarkDone.Flags().StringVarP(&flagCoursePath, "path", "p", "", "course path")
}
