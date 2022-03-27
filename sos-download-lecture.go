package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type SosLectureInfo struct {
	Path    string // m3u8 path
	Content []byte // m3u8 content
}

func sosGetWanmenLectureInfo(sosLectureDir string) (*SosLectureInfo, error) {
	matches, err := filepath.Glob(filepath.Join(sosLectureDir, "*.m3u8"))
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no m3u8 file found in %s", sosLectureDir)
	}
	if len(matches) > 0 {
		match := matches[0]
		for _, m := range matches {
			if strings.HasSuffix(m, "_pc_high.m3u8") {
				match = m
				break
			}
		}
		matches[0] = match
	}

	match := matches[0]

	f, err := os.ReadFile(match)
	if err != nil {
		return nil, fmt.Errorf("cannot read lecture m3u8: %w", err)
	}

	return &SosLectureInfo{match, f}, nil
}

func parseSosM3U8Data(data []byte) ([]string, error) {
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
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
func sosDownloadLectureM3U8(info *SosLectureInfo, target string, noConvert bool, full bool, reportProgress updateProgressFunc) error {
	partDonePath := target + ".stream.mp4"

	legacyPartDonePath1 := info.Path + ".stream.mp4"
	if !isExist(partDonePath) && isExist(legacyPartDonePath1) {
		_ = os.MkdirAll(filepath.Dir(partDonePath), 0755)
		_ = os.Rename(legacyPartDonePath1, partDonePath)
	}
	legacyPartDonePath2 := info.Path + ".part.stream.mp4"
	if !isExist(partDonePath) && isExist(legacyPartDonePath2) {
		_ = os.MkdirAll(filepath.Dir(partDonePath), 0755)
		_ = os.Rename(legacyPartDonePath2, partDonePath)
	}

	if !isExist(partDonePath) {
		// 下载 m3u8 文件，返回 ts 列表
		tsList, err := parseSosM3U8Data(info.Content)
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

		// python 生成的 ts 路径
		pythonTsPrefix := strings.TrimSuffix(info.Path, ".m3u8")

		N := len(tsList)

		baseHost := "https://media.wanmen.org/"

		for i, ts := range tsList {
			reportProgress("downloading", i, N)

			pythonTsPath := fmt.Sprintf("%s%d.ts", pythonTsPrefix, i+1)

			if isExist(pythonTsPath) {
				err := func() error {
					ff, err := os.Open(pythonTsPrefix)
					if err != nil {
						return fmt.Errorf("cannot open python ts file: %v", err)
					}
					defer ff.Close()

					_, err = io.Copy(f, ff)
					if err != nil {
						return fmt.Errorf("cannot copy from python ts: %v", err)
					}

					return nil
				}()
				if err != nil {
					return err
				}
			} else {
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

	// 将 part done file 转换为 mp4

	if !noConvert { // convert to mp4
		err := convertToMp4(target, partDonePath, reportProgress)
		if err != nil {
			return fmt.Errorf("cannot convert part file to mp4: %w", err)
		}
	}

	return nil
}

// target 为下载目标的绝对路径
// full 为 true 代表不会去除首尾万门的「广告」
// f 用来返回当前的下载进度
// 返回值的第一个参数代表下载状况，0 代表正常，1-4 代表下载到了非超清版本，-1 代表无法下载
// 当且仅当第一个返回值为 -1 时，会带有 error 参数
func sosDownloadLecture(lectureID string, lecturePath string, sosLecturePath, metaPrefix string, noConvert bool, full bool, f updateProgressFunc) (int, error) {
	target := lecturePath
	partDonePath := target + ".stream.mp4"

	if isExist(partDonePath) { // 简单转换为 mp4
		if !noConvert {
			err := convertToMp4(target, partDonePath, f)
			if err != nil {
				err = fmt.Errorf("cannot convert to mp4: %w", err)

				_ = appendJSON(metaPrefix+".error.jsonl", map[string]interface{}{
					"op":  "fail",
					"err": err.Error(),
				})
				return -1, err
			}
		}

		return 0, nil
	}

	info, err := sosGetWanmenLectureInfo(sosLecturePath)
	if err != nil {
		err = fmt.Errorf("cannot get lecture info: %v", err)
		_ = appendJSON(metaPrefix+".error.jsonl", map[string]interface{}{
			"op":  "fail",
			"err": err.Error(),
		})
		return -1, err
	}

	_ = os.MkdirAll(filepath.Dir(lecturePath), 0755)

	err = sosDownloadLectureM3U8(info, target, noConvert, full, f)
	if err != nil {
		_ = appendJSON(metaPrefix+".error.jsonl", map[string]interface{}{
			"op":  "fail",
			"err": err.Error(),
		})

		return -1, err
	}

	return 0, nil
}
