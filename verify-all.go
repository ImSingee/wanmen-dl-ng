package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var cmdVerifyAll = &cobra.Command{
	Use:     "verify-all [<filename>]",
	Aliases: []string{"va"},
	Args:    cobraParseList,
	RunE: func(cmd *cobra.Command, args []string) error {
		totalCount := len(list)
		errorCount := 0
		bypassCount := 0

		for _, courseId := range list {
			state := verify(courseId, flagCoursePath, flagSkipFFMpeg, flagOffline, flagUpdateMeta)
			switch state {
			case 0: // success
			case 1: // error
				errorCount++
				fmt.Printf("Course ID %s verified fail\n", courseId)
			case 2:
				bypassCount++
				fmt.Printf("Course ID %s bypassed\n", courseId)
			}
		}

		if errorCount == 0 {
			if bypassCount == 0 {
				fmt.Printf("%d/%d courses verified successfully\n", totalCount, totalCount)
			} else {
				fmt.Printf("%d/%d courses verified successfully, %d courses bypassed\n", totalCount-bypassCount, totalCount, bypassCount)
			}
		} else {
			if bypassCount == 0 {
				redPrintf("%d/%d courses verified failed\n", errorCount, totalCount)
			} else {
				redPrintf("%d/%d courses verified failed, %d courses bypassed\n", errorCount, totalCount, bypassCount)
			}
		}

		return nil
	},
}

func init() {
	app.AddCommand(cmdVerifyAll)

	cmdVerifyAll.Flags().BoolVarP(&flagSkipFFMpeg, "skip-ffmpeg", "m", false, "")
	cmdVerifyAll.Flags().BoolVarP(&flagOffline, "offline", "o", true, "offline mode (won't request wanmen api again)")
	cmdVerifyAll.Flags().BoolVar(&flagUpdateMeta, "update-meta", true, "also update exist meta")
}
