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
	app.PersistentFlags().StringP("config", "C", "", "配置文件路径，默认为 ./config.json")

	err := app.Execute()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
}
