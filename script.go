package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"path/filepath"
)

var cmdScript = &cobra.Command{
	Use:    "script",
	Short:  "Script commands",
	Hidden: true,
	RunE:   script,
}

func script(cmd *cobra.Command, args []string) error {
	l, err := getList("/Users/wangxuan/Desktop/wanmen/blocks/l/l")
	if err != nil {
		return err
	}

	_ = l

	//for _, id := range l {
	//name, ok := getSosName(id)
	//if !ok {
	//	return fmt.Errorf("invalid id: %s", id)
	//}

	//p := filepath.Join(config.SosDir, sosCleanName(name))

	// 重建 API 数据
	//courseLectures := &CourseLectures{}
	//courseInfo := &CourseInfo{}
	//
	//lecturesMetaPath := filepath.Join(metaDir, "lectures.json")
	//infoMetaPath := filepath.Join(metaDir, "info.json")
	//
	//err = exjson.Read(lecturesMetaPath, &courseLectures.Chapters)
	//if err != nil {
	//	return fmt.Errorf("cannot load lectures meta file %s: %v", lecturesMetaPath, err)
	//}
	//
	//err = exjson.Read(infoMetaPath, courseInfo)
	//if err != nil {
	//	return fmt.Errorf("cannot load info meta file %s: %v", infoMetaPath, err)
	//}
	//}

	return nil
}

func script1(cmd *cobra.Command, args []string) error {
	l, err := getList("/Users/wangxuan/Desktop/wanmen/blocks/l/l")
	if err != nil {
		return err
	}

	for _, id := range l {
		name, ok := getSosName(id)
		if !ok {
			return fmt.Errorf("invalid id: %s", id)
		}

		p := filepath.Join(config.SosDir, sosCleanName(name))
		if !isExist(p) {
			fmt.Println("Not Found ", p, "\t", id)
		} else {
			fmt.Println(p)
		}
	}

	return nil
}

func init() {
	app.AddCommand(cmdScript)
}
