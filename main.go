package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var app = &cobra.Command{
	Use:           "wanmen-dl",
	Short:         "直接下载某门课程",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func main() {
	err := app.Execute()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
}
