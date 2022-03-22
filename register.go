package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var cmdRegister = &cobra.Command{
	Use:   "register <course-id> [<course-name>]",
	Short: "记录 ID 对应的名称，用于解决内置数据库不全的问题",
	RunE: func(cmd *cobra.Command, args []string) error {
		var courseId, courseName string

		if len(args) == 1 {
			courseId = args[0]
		} else if len(args) == 2 {
			courseId = args[0]
			courseName = args[1]
		} else {
			return fmt.Errorf("register <course-id> [<course-name>]")
		}

		if courseName == "" {
			info, err := apiGetWanmenCourseInfo(courseId)
			if err != nil {
				return fmt.Errorf("cannot get course info: %v", err)
			}
			courseName = info.Name
			fmt.Println(courseName)
		}

		config.NameMap[courseId] = courseName

		p, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("cannot marshal config: %v", err)
		}

		err = os.WriteFile("config.json", p, 0644)
		if err != nil {
			return fmt.Errorf("cannot write config: %v", err)
		}

		return nil
	},
}

func init() {
	app.AddCommand(cmdRegister)
}
