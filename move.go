package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var flagMoveTo string

var cmdMoveAll = &cobra.Command{
	Use:     "move-all [<filename>]",
	Aliases: []string{"ma"},
	Args:    cobraParseList,
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagMoveTo == "" {
			return errors.New("-t is required")
		}

		for _, id := range list {
			err := move(id, flagMoveTo)
			if err != nil {
				fmt.Println(id, "not moved, error:", err)
			} else {
				fmt.Println(id, "moved")
			}
		}

		return nil
	},
}

var cmdMove = &cobra.Command{
	Use:     "move <course-id>",
	Aliases: []string{"m"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("requires course-id")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagMoveTo == "" {
			return errors.New("-t is required")
		}

		for _, id := range args {
			err := move(id, flagMoveTo)
			if err != nil {
				fmt.Println(id, "not moved, error:", err)
			} else {
				fmt.Println(id, "moved")
			}
		}

		return nil
	},
}

func move(courseId string, base string) error {
	courseName, ok := GetName(courseId)

	const tmpl = "%-30s %-12s    %s\n"

	if !ok {
		return fmt.Errorf("unknown course")
	}

	d := filepath.Join(config.DownloadTo, cleanName(courseName))

	if !isExist(d) {
		return fmt.Errorf("path %s is not exist", d)
	}

	to := filepath.Join(base, cleanName(courseName))
	if isExist(to) {
		return fmt.Errorf("path %s is already exist", to)
	}

	return os.Rename(d, to)
}

func init() {
	app.AddCommand(cmdMove)
	app.AddCommand(cmdMoveAll)

	cmdMove.Flags().StringVarP(&flagMoveTo, "to", "t", "", "move to")
	cmdMoveAll.Flags().StringVarP(&flagMoveTo, "to", "t", "", "move to")
}
