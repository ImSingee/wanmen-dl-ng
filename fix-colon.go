package main

import (
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

//var flagReverse bool
var flagGlob bool

var fixColonCmd = &cobra.Command{
	Use: "fix-colon",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, arg := range args {
			if flagGlob {
				files, err := filepath.Glob(arg)
				if err != nil {
					return err
				}
				for _, file := range files {
					err = fixColon(file)
					if err != nil {
						return err
					}
				}
			} else {
				err := fixColon(arg)
				if err != nil {
					return err
				}
			}
		}

		return nil
	},
}

func fixColon(filename string) error {
	if !isExist(filename) {
		return nil
	}

	newFilename := strings.ReplaceAll(filename, ":", "：")
	if newFilename == filename {
		return nil
	}

	return os.Rename(filename, newFilename)
}

func init() {
	app.AddCommand(fixColonCmd)

	//fixColonCmd.Flags().BoolVarP(&flagReverse, "reverse", "r", false, "")
	fixColonCmd.Flags().BoolVarP(&flagGlob, "glob", "g", false, "")

}
