package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime"
)

var cmdDownloadAll = &cobra.Command{
	Use:  "download-all [<filename>]",
	Args: cobraParseList,
	RunE: func(cmd *cobra.Command, args []string) error {
		errorsMap := map[string]error{}

		for _, courseId := range list {
			err := download(courseId, flagDownloadTo, flagForce, flagFull, flagConcurrency)
			fmt.Println()
			if err != nil {
				errorsMap[courseId] = err
				fmt.Printf("!!! %s error: %v\n", courseId, err)
			}
		}

		if len(errorsMap) != 0 {
			for id, err := range errorsMap {
				fmt.Printf("%s error: %v\n", id, err)
			}

			return fmt.Errorf("some course error")
		}

		return nil
	},
}

func init() {
	app.AddCommand(cmdDownloadAll)

	cmdDownloadAll.Flags().IntVarP(&flagForce, "force", "f", 0, "跳过去重（0-不跳过, 1-跳过课程检测, 2-跳过文件检测)")
	cmdDownloadAll.Flags().BoolVar(&flagFull, "full", false, "不去除万门广告")
	cmdDownloadAll.Flags().IntVarP(&flagConcurrency, "concurrency", "c", runtime.NumCPU()*4, "并发数，默认为 CPU 数量 * 4")

}
