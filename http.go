package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

var client = http.DefaultClient

// 自动重试，暂不考虑 body
func httpRequestWithAutoRetry(request *http.Request) (*http.Response, error) {
	// 最多重试 5 次
	var latestErr error
	for i := 0; i < 5; i++ {
		response, err := client.Do(request.Clone(context.Background()))
		if err != nil {
			// 可重试错误
			latestErr = err
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
			// 先行读取 response body
			body, err := io.ReadAll(response.Body)
			if err != nil {
				response.Body.Close()
				latestErr = fmt.Errorf("cannot read response body: %v", err)
				continue
			}
			response.Body = io.NopCloser(bytes.NewReader(body))

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

func httpRequestWithAutoRetryAndCustomHandleResponse(request *http.Request, handler func(response *http.Response) error) (*http.Response, error) {
	// 最多重试 5 次
	var latestErr error
	for i := 0; i < 5; i++ {
		response, err := client.Do(request.Clone(context.Background()))
		if err != nil {
			// 可重试错误
			latestErr = err
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
			err := handler(response)
			if err != nil {
				response.Body.Close()
				latestErr = fmt.Errorf("cannot read response body: %v", err)
				continue
			}

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
