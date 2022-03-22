package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var ffmpegName string
var ffmpegPath string

// target 为下载目标的绝对路径
// full 为 true 代表不会去除首尾万门的「广告」
func tryDownloadLectureM3U8(url string, target string, full bool, reportProgress updateLectureProgressFunc) error {
	// 下载 m3u8 文件，返回 ts 列表
	tsList, err := downloadAndParseM3U8(url)
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

	for i, ts := range tsList {
		reportProgress("downloading", i, N)

		req, err := http.NewRequest("GET", urljoin(url, ts), nil)
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

	reportProgress("ffmpeg_start")

	ffmpegTarget := target + ".ffmpeg.mp4"

	// 运行 ffmpeg 将 ts 集合转换为 mp4
	ffmpegOutput, err := exec.Command(ffmpegPath, "-y", "-loglevel", "error",
		"-i", partFile,
		"-bsf:a", "aac_adtstoasc",
		"-vcodec", "copy",
		"-acodec", "copy",
		ffmpegTarget,
	).CombinedOutput()
	if err != nil {
		reportProgress("ffmpeg_error", string(ffmpegOutput))
		return err
	}
	err = os.Rename(ffmpegTarget, target)
	if err != nil {
		return fmt.Errorf("cannot rename ffmpeg middle target to final target: %w", err)
	}
	reportProgress("ffmpeg_done")

	_ = os.Remove(partFile)

	return nil
}

func downloadAndParseM3U8(url string) ([]string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot init request: %v", err)
	}
	req.Header = getMediaHeaders()

	response, err := httpRequestWithAutoRetry(req)
	if err != nil {
		return nil, fmt.Errorf("cannot request m3u8 file download: %v", err)
	}
	defer response.Body.Close()

	all, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response for m3u8 file download request: %v", err)
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

func init() {
	// Ensure ffmpeg exist

	if runtime.GOOS == "Windows" {
		ffmpegName = "ffmpeg.exe"
	} else {
		ffmpegName = "ffmpeg"
	}

	var err error
	ffmpegPath, err = exec.LookPath(ffmpegName)
	if err != nil {
		panic("Cannot found ffmpeg!")
	}
}
