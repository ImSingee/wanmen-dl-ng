package main

import (
	"fmt"
	"github.com/ImSingee/go-ex/exjson"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var cmdDownloadSosDoc = &cobra.Command{
	Use: "download-sos-doc",
	RunE: func(cmd *cobra.Command, args []string) error {
		downloadSosDoc(args, 0, 2)
		return nil
	},
}

func downloadSosDoc(courseIds []string, concurrency, forceLevel int) {
	wg := sync.WaitGroup{}

	queue := make(chan *DownloadTask, 256)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for _, courseId := range courseIds {
			name, ok := GetName(courseId)
			if !ok {
				redPrintf("cannot find course name for %s\n", courseId)
				continue
			}

			sosName, ok := getSosName(courseId)
			if !ok {
				redPrintf("cannot find course sosName for %s\n", courseId)
				continue
			}

			courseDir := filepath.Join(config.DownloadTo, name)

			sosPath := filepath.Join(config.SosDir, sosCleanName(sosName))
			if !isExist(sosPath) {
				redPrintf("cannot find course path at %s\n", sosPath)
				continue
			}

			metaDir := filepath.Join(courseDir, ".meta")
			_ = os.MkdirAll(metaDir, 0755)

			sosInfoPath := filepath.Join(sosPath, "info.json")

			courseInfo := &CourseInfo{}
			err := exjson.Read(sosInfoPath, courseInfo)
			if err != nil {
				redPrintf("cannot load info meta file %s: %v", sosInfoPath, err)
				continue
			}

			// 将 meta 写入文件中
			metaInfoPath := filepath.Join(metaDir, "info.json")
			_ = CopyFile(sosInfoPath, metaInfoPath)

			courseDocuments := courseInfo.Documents

			for i, doc := range courseDocuments {
				doc.Index = i + 1

				// 万门的某些课程 ext 会出现两次
				doc.Name = strings.TrimSuffix(doc.Name, "."+doc.Ext)

				queue <- &DownloadTask{
					MetaDir: metaDir,
					Doc: &DocDownloadTask{
						Document:     doc,
						DocumentPath: filepath.Join(courseDir, "资料", cleanName(fmt.Sprintf("%d - %s.%s", i+1, doc.Name, doc.Ext))),
					},
				}
			}
		}

		close(queue)
	}()

	if concurrency == 0 {
		concurrency = runtime.NumCPU() * 4
	}

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(workerId int) {
			defer wg.Done()
			defer fmt.Println("quit", workerId)

			for {
				task, ok := <-queue // 从队列中取出一个下载任务
				if !ok {
					break
				}

				fmt.Println("start", workerId, task)

				if forceLevel != 2 && !task.ForceReDownload {
					if isExist(task.Path()) {
						fmt.Println("skip-task", workerId, task)
						continue
					}
				}

				// task.Doc
				f := func(a string, v ...interface{}) {
					fmt.Println("doc", workerId, task, a, v)
				}

				err := downloadDoc(task.Doc.Document, task.Path(), task.MetaPrefix(), f)
				if err != nil {
					redPrintf("%s %v %v %v\n", "error", workerId, task, err)
					continue
				}

				fmt.Println("done", workerId, task)
			}
		}(i + 1)
	}

	wg.Wait()
}

func init() {
	app.AddCommand(cmdDownloadSosDoc)
}
