package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var client = http.DefaultClient

// 自动重试，暂不考虑 body
func httpRequestWithAutoRetry(request *http.Request) (*http.Response, error) {
	// 最多重试 3 次
	var latestErr error
	for i := 0; i < 3; i++ {
		response, err := client.Do(request.Clone(context.Background()))
		if err != nil {
			// 可重试错误
			latestErr = err
			_ = response.Body.Close()
			continue
		}

		// 没有错误
		latestErr = nil
		// 万门土豆服务器又崩了，重试
		if response.StatusCode >= 500 {
			latestErr = fmt.Errorf("API Status Code = %d", response.StatusCode)
			_ = response.Body.Close()
			continue
		}

		if response.StatusCode >= 200 && response.StatusCode < 400 { // 成功
			return response, nil
		}

		// StatusCode 错误，因为参数或授权等因素导致的
		_ = response.Body.Close()
		return nil, fmt.Errorf("API Status Code = %d", response.StatusCode)
	}

	if latestErr != nil {
		return nil, fmt.Errorf("error too many times: %w", latestErr)
	}

	// impossible
	return nil, fmt.Errorf("something strange happened")
}

type LectureInfo struct {
	Name        string
	VideoStream *VideoStream
	RawJsonBody []byte
}

func apiGetWanmenLectureInfo(lectureId string) (*LectureInfo, error) {
	url := fmt.Sprintf("https://api.wanmen.org/4.0/content/lectures/%s?routeId=main&debug=1", lectureId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot build lecture info api request: %w", err)
	}
	req.Header = getHeaders()

	response, err := httpRequestWithAutoRetry(req)
	if err != nil {
		return nil, fmt.Errorf("cannot request lecture info api: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot request lecture info api (read body): %w", err)
	}

	var jsonBody map[string]interface{}

	err = json.Unmarshal(body, &jsonBody)
	if err != nil {
		return nil, fmt.Errorf("cannot request lecture info api (unmarshal json): %w", err)
	}

	// get name
	name, _ := jsonBody["name"].(string)

	// get video hls
	var hls *VideoStream
	if v, ok := jsonBody["video"]; ok {
		hls = tryGetHls(v)
	} else {
		hls = tryGetHls(jsonBody)
	}

	if hls == nil {
		return nil, errors.New("no hls found")
	}

	return &LectureInfo{
		Name:        name,
		VideoStream: hls,
		RawJsonBody: body,
	}, nil
}

type VideoStream struct {
	PcHigh string `json:"pcHigh"`
	PcMid  string `json:"pcMid"`
	PcLow  string `json:"pcLow"`
}

func tryGetHls(v interface{}) *VideoStream {
	// v: {"hls": { ... }}
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	hls := m["hls"]
	if hls == nil {
		return nil
	}
	// hls: { "pcHigh": "...", ...}

	hlsM, ok := hls.(map[string]interface{})
	if !ok {
		return nil
	}

	pcHigh, _ := hlsM["pcHigh"].(string)
	pcMid, _ := hlsM["pcMid"].(string)
	pcLow, _ := hlsM["pcLow"].(string)

	return &VideoStream{
		PcHigh: pcHigh,
		PcMid:  pcMid,
		PcLow:  pcLow,
	}
}
