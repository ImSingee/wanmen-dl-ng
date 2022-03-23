package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func downloadDoc(docInfo *CourseInfo_Document, saveTo string, metaPrefix string, updateProgress updateProgressFunc) (err error) {
	defer func() {
		if err != nil {
			_ = appendJSON(metaPrefix+".error.jsonl", map[string]interface{}{
				"op":  "fail",
				"err": err,
			})
		}
	}()

	_ = os.MkdirAll(filepath.Dir(saveTo), 0755)

	req, err := http.NewRequest("GET", docInfo.URL, nil)
	if err != nil {
		return fmt.Errorf("doc url is invalid: %w", err)
	}
	req.Header = getMediaHeaders()

	tempSaveTo := saveTo + ".tmp"

	f, err := os.Create(tempSaveTo)
	if err != nil {
		return fmt.Errorf("cannot open or create file: %v", err)
	}
	defer f.Close()

	_, err = httpRequestWithAutoRetryAndCustomHandleResponse(req, func(response *http.Response) error {
		updateProgress("downloading", 0, int(response.ContentLength))

		_, err := io.Copy(f, response.Body)
		if err != nil { // reset file
			f.Truncate(0)
			f.Seek(0, 0)
		} else {
			updateProgress("downloading", int(response.ContentLength), int(response.ContentLength))
		}
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to download doc: %w", err)
	}

	err = os.Rename(tempSaveTo, saveTo)
	if err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
