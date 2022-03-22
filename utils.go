package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var nameReplacer = strings.NewReplacer(
	"\\", " ",
	"/", " ",
	":", " ",
	"*", " ",
	"?", " ",
	`'`, " ",
	`"`, " ",
	"<", " ",
	">", " ",
	"|", " ",
)

func cleanName(name string) string {
	return nameReplacer.Replace(name)
}

func getToken() (string, string) {
	timeStr := fmt.Sprintf("%x", time.Now().Unix())

	h := md5.New()
	h.Write([]byte("5ec029c599f7abec29ebf1c50fcc05a0"))
	h.Write([]byte(timeStr))

	token := hex.EncodeToString(h.Sum(nil))

	return timeStr, token
}

func getHeaders() http.Header {
	timeStr, token := getToken()

	return http.Header{
		"Authorization": []string{config.Authorization},
		"User-Agent":    []string{config.UserAgent},
		"x-sa":          []string{"9e2fc61b78106962a1fa5c5ba6874acaaf0cabfecb6f85ae2d4a082b672b9139f1466529572da95c36dd39a7cf9c8444"},
		"accept":        []string{"vnd.wanmen.v9+json"},
		"x-app":         []string{"uni"},
		"x-platform":    []string{"web"},
		"x-time":        []string{timeStr},
		"x-token":       []string{token},
	}
}

func getMediaHeaders() http.Header {
	return http.Header{
		"User-Agent": []string{config.UserAgent},
		"Referer":    []string{"https://www.wanmen.org/"},
	}
}

func urljoin(base, endpoint string) string {
	i := strings.LastIndex(base, "/")
	return base[:i+1] + strings.TrimPrefix(endpoint, "/")
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func appendFile(path string, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(strings.TrimSpace(content) + "\n")
	return err
}

func appendJSON(path string, content map[string]interface{}) error {
	data, err := json.Marshal(content)
	if err != nil {
		return err
	}

	return appendFile(path, string(data))
}
