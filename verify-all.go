package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var cmdVerifyAll = &cobra.Command{
	Use:  "verify-all [<filename>]",
	Args: cobraParseList,
	RunE: func(cmd *cobra.Command, args []string) error {
		anyError := false

		for _, courseId := range list {
			ok := verify(courseId, "", flagOffline, flagUpdateMeta)
			if !ok {
				anyError = true
				fmt.Printf("Course ID %s verified fail\n", courseId)
			}
		}

		if anyError {
			return fmt.Errorf("some errors occurred")
		}

		return nil
	},
}

func init() {
	app.AddCommand(cmdVerifyAll)

	cmdVerifyAll.Flags().BoolVarP(&flagOffline, "offline", "o", false, "offline mode (won't request wanmen api again)")
	cmdVerifyAll.Flags().BoolVar(&flagUpdateMeta, "update-meta", true, "also update exist meta")
}
