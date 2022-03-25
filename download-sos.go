package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

func parseSosM3U8(metapath string) ([]string, error) {
	all, err := os.ReadFile(metapath)
	if err != nil {
		return nil, fmt.Errorf("cannot read m3u8 file download request: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(all)), "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		result = append(result, line)
	}

	return result, nil
}

// target 为下载目标的绝对路径
// full 为 true 代表不会去除首尾万门的「广告」
func downloadSosM3U8(metapath string, target string, full bool, reportProgress updateProgressFunc) error {
	partDonePath := target + ".stream.mp4"

	if !isExist(partDonePath) {
		// 下载 m3u8 文件，返回 ts 列表
		tsList, err := parseSosM3U8(metapath)
		if err != nil {
			return err
		}

		if !full { // 忽略首尾片段
			tsList = tsList[1 : len(tsList)-1]
		}

		// 临时的 ts 拼接下载路径
		partFile := target + ".part"
		f, err := os.Create(partFile)
		if err != nil {
			return fmt.Errorf("cannot create and open part file: %v", err)
		}
		defer f.Close()

		N := len(tsList)

		baseHost := "https://media.wanmen.org/"

		for i, ts := range tsList {
			reportProgress("downloading", i, N)

			req, err := http.NewRequest("GET", urljoin(baseHost, ts), nil)
			if err != nil {
				return fmt.Errorf("invalid ts %d download request: %w", i, err)
			}
			req.Header = getMediaHeaders()

			response, err := httpRequestWithAutoRetry(req)
			if err != nil {
				return fmt.Errorf("cannot download ts %d: %w", i, err)
			}

			_, err = io.Copy(f, response.Body)
			if err != nil {
				return fmt.Errorf("cannot write ts %d to part file: %w", i, err)
			}
		}
		reportProgress("downloading", N, N)

		// 将 part file 关闭以保证保存（可以让下面的 ffmpeg 转换看到）
		err = f.Close()
		if err != nil {
			return fmt.Errorf("cannot close part file: %w", err)
		}

		err = os.Rename(partFile, partDonePath)
		if err != nil {
			return fmt.Errorf("cannot rename part file: %w", err)
		}
	}

	return nil
}

var cmdDownloadSos = &cobra.Command{
	Use: "download-sos",
	Run: func(cmd *cobra.Command, args []string) {
		for _, courseId := range args {
			name, ok := GetName(courseId)
			if !ok {
				fmt.Println("cannot find course name for", courseId)
				continue
			}

			sosPath := filepath.Join(config.SosDir, sosCleanName(name))

			fmt.Println("Download", courseId, sosPath)
			sosDownload(sosPath, 0)
			fmt.Println("Download", courseId, "DONE", sosPath)
		}
	},
}

func sosDownload(sosPath string, concurrency int) {
	queue := make(chan string, 256)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := filepath.WalkDir(sosPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			if strings.HasSuffix(d.Name(), ".m3u8") {
				queue <- path
			}

			return nil
		})
		if err != nil {
			fmt.Printf("Error when walk %s: %v\n", sosPath, err)
		}

		close(queue)
	}()

	if concurrency <= 0 {
		concurrency = runtime.NumCPU() * 4
	}

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(workerId int) {
			defer wg.Done()
			defer fmt.Println("worker shutdown", workerId)

			for {
				task, ok := <-queue // 从队列中取出一个下载任务
				if !ok {
					break
				}

				fmt.Println("start", workerId, task)

				downloadTo := task + ".part"
				doneTo := task + ".stream.mp4"

				if isExist(doneTo) {
					fmt.Println("skip-task", workerId, task)
					continue
				}

				if true { // is done
					f := func(a string, v ...interface{}) {
						fmt.Println("lecture", workerId, task, a, v)
					}

					err := downloadSosM3U8(task, downloadTo, false, f)

					if err != nil {
						fmt.Println("error", workerId, task, err)
						continue
					}
				} else { // task.Doc
					// doc
				}

				fmt.Println("done", workerId, task)
			}
		}(i + 1)
	}

	wg.Wait()
}

func init() {
	app.AddCommand(cmdDownloadSos)
}
