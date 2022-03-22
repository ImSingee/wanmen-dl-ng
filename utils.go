package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
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
