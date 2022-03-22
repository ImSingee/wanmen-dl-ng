package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Authorization string
	UserAgent     string
	DownloadTo    string
	NumProcess    int
	NameMap       map[string]string
}

var config = &Config{
	Authorization: "Bearer xxx",
	UserAgent:     `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36`,
	DownloadTo:    "/data/万门",
	NumProcess:    32,
	NameMap:       map[string]string{},
}

func init() {
	data, err := os.ReadFile("config.json")
	if err != nil {
		if os.IsNotExist(err) {
			return
		}

		panic("Cannot open config: " + err.Error())
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		panic("Cannot load config: " + err.Error())
	}
}
