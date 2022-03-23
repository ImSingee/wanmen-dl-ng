package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var flagDryRun bool

var cmdClean = &cobra.Command{
	Use:   "clean [<course-id>/<course-path>]",
	Short: "Simply clean all .part/.ffmpeg.mp4 files",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("invalid: must provide at least one id or path")
		}

		for _, arg := range args {
			name, ok := GetName(arg)
			if ok {
				err := cleanByName(name, flagDryRun)
				if err != nil {
					return err
				}
			} else {
				err := cleanByPath(arg, flagDryRun)
				if err != nil {
					return err
				}
			}
		}

		return nil
	},
}

func cleanById(id string, dryRun bool) error {
	name, ok := GetName(id)
	if !ok {
		return fmt.Errorf("unknown id %s", id)
	}

	return cleanByName(name, dryRun)
}

func cleanByName(name string, dryRun bool) error {
	return cleanByPath(filepath.Join(config.DownloadTo, cleanName(name)), dryRun)
}

func cleanByPath(p string, dryRun bool) error {
	return filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		name := d.Name()

		if strings.HasSuffix(name, ".part") ||
			strings.HasSuffix(name, ".ffmpeg.mp4") ||
			strings.HasSuffix(name, ".part.mp4") ||
			strings.HasSuffix(name, ".tmp") {

			fmt.Print("Remove ", path)

			if !dryRun {
				err := os.Remove(path)
				if err != nil {
					fmt.Print(" ERROR: ", err)
				}
			}
			fmt.Println()
		}

		return nil
	})
}

func init() {
	app.AddCommand(cmdClean)

	cmdClean.Flags().BoolVar(&flagDryRun, "dry", false, "not really execute")
}
