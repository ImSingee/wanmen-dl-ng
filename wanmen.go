package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
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

type LectureInfo struct {
	Name        string
	VideoStream *VideoStream
	// TODO check video size?
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
	PcHigh    string `json:"pcHigh"`
	PcMid     string `json:"pcMid"`
	PcLow     string `json:"pcLow"`
	MobileMid string `json:"mobileMid"`
	MobileLow string `json:"mobileLow"`
}

func (vs *VideoStream) ToDownload() []string {
	target := make([]string, 0, 5)

	if v := vs.PcHigh; v != "" {
		target = append(target, v)
	}
	if v := vs.PcMid; v != "" {
		target = append(target, v)
	}
	if v := vs.MobileMid; v != "" {
		target = append(target, v)
	}
	if v := vs.PcLow; v != "" {
		target = append(target, v)
	}
	if v := vs.MobileLow; v != "" {
		target = append(target, v)
	}

	return target
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
	mobileMid, _ := hlsM["mobileMid"].(string)
	mobileLow, _ := hlsM["mobileLow"].(string)

	return &VideoStream{
		PcHigh:    pcHigh,
		PcMid:     pcMid,
		PcLow:     pcLow,
		MobileMid: mobileMid,
		MobileLow: mobileLow,
	}
}

type CourseLectures struct {
	Chapters CourseInfo_Chapters
	Raw      []byte
}

type CourseInfo_Chapters []*CourseInfo_Chapter

type CourseInfo_Chapter struct {
	Index    int                   // 由外部赋值, base-1
	ID       string                `json:"_id"`
	Courseid string                `json:"courseId"`
	Name     string                `json:"name"`
	Order    int                   `json:"order"`
	Hide     bool                  `json:"hide"`
	Children []*CourseInfo_Lecture `json:"children"`
}

type CourseInfo_Lecture struct {
	Index         int     // 由外部赋值, base-1
	ID            string  `json:"_id"`
	Name          string  `json:"name"`
	ParentId      string  `json:"parentId"`
	CourseId      string  `json:"courseId"`
	Order         int     `json:"order"`
	Hide          bool    `json:"hide"`
	VideoDuration float64 `json:"videoDuration"`
	VideoSize     struct {
		MobileLow int `json:"mobileLow"`
		PcHigh    int `json:"pcHigh"`
		PcLow     int `json:"pcLow"`
		MobileMid int `json:"mobileMid"`
		PcMid     int `json:"pcMid"`
	} `json:"videoSize"`
}

func apiGetWanmenCourseLectures(courseId string) (*CourseLectures, error) {
	url := fmt.Sprintf("https://api.wanmen.org/4.0/content/lectures?courseId=%s&debug=1", courseId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot build course lectures api request: %w", err)
	}
	req.Header = getHeaders()

	response, err := httpRequestWithAutoRetry(req)
	if err != nil {
		return nil, fmt.Errorf("cannot request course lectures api: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot request course lectures api (read body): %w", err)
	}

	var courseInfo CourseInfo_Chapters

	err = json.Unmarshal(body, &courseInfo)
	if err != nil {
		return nil, fmt.Errorf("cannot request course lectures api (unmarshal json): %w", err)
	}

	return &CourseLectures{
		Chapters: courseInfo,
		Raw:      body,
	}, nil
}

type CourseInfo struct {
	Price                   int                    `json:"price"`
	Hide                    bool                   `json:"hide"`
	Likes                   int                    `json:"likes"`
	Type                    string                 `json:"type"`
	Mediatype               string                 `json:"mediaType"`
	Status                  string                 `json:"status"`
	State                   string                 `json:"state"`
	Isended                 bool                   `json:"isEnded"`
	Isonlive                bool                   `json:"isOnLive"`
	Playbackstatus          string                 `json:"playbackStatus"`
	App                     string                 `json:"app"`
	Views                   int                    `json:"views"`
	Communitytype           string                 `json:"communityType"`
	Subject                 string                 `json:"subject"`
	Quizshuffled            bool                   `json:"quizShuffled"`
	Podcasttype             string                 `json:"podcastType"`
	Isregionrestricted      bool                   `json:"isRegionRestricted"`
	Namehistory             []string               `json:"nameHistory"`
	K12Hide                 bool                   `json:"k12Hide"`
	Opduration              int                    `json:"opDuration"`
	Edduration              int                    `json:"edDuration"`
	Introductionimages      []interface{}          `json:"introductionImages"`
	Schedule                []interface{}          `json:"schedule"`
	Name                    string                 `json:"name"`
	Expiration              int                    `json:"expiration"`
	Faq                     []interface{}          `json:"faq"`
	Documents               []*CourseInfo_Document `json:"documents"`
	Createdat               time.Time              `json:"createdAt"`
	Updatedat               time.Time              `json:"updatedAt"`
	Showat                  time.Time              `json:"showAt"`
	Hoursoflesson           float64                `json:"hoursOfLesson"`
	Bigimage                string                 `json:"bigImage"`
	Description             string                 `json:"description"`
	Smallimage              string                 `json:"smallImage"`
	Videocount              int                    `json:"videoCount"`
	Lastupdatetimeoflecture time.Time              `json:"lastUpdateTimeOfLecture"`
	Tag                     string                 `json:"tag"`
	Topiccountinplan        int                    `json:"topicCountInPlan"`
	Endat                   time.Time              `json:"endAt"`
	Specialprice            int                    `json:"specialPrice"`
	Incoming                bool                   `json:"incoming"`
	Bigimageurl             string                 `json:"bigImageUrl"`
	Smallimageurl           string                 `json:"smallImageUrl"`
	Teachername             string                 `json:"teacherName"`
	Teacheravatar           string                 `json:"teacherAvatar"`
	Teacherdescription      string                 `json:"teacherDescription"`
	ID                      string                 `json:"id"`
	Expirationtext          string                 `json:"expirationText"`
	Share                   struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Image   string `json:"image"`
	} `json:"share"`
	Paymentalert     string      `json:"paymentAlert"`
	Ispaid           bool        `json:"isPaid"`
	Hidedat          interface{} `json:"hidedAt"`
	Remainingdays    int         `json:"remainingDays"`
	Expiredat        interface{} `json:"expiredAt"`
	Remainingminutes int         `json:"remainingMinutes"`
	Isliked          bool        `json:"isLiked"`
	Isfaved          bool        `json:"isFaved"`
	Ispass           bool        `json:"isPass"`
	Hastopic         bool        `json:"hasTopic"`
	Hasexamticket    bool        `json:"hasExamTicket"`
	Latestlecture    interface{} `json:"latestLecture"`
	Hascontents      bool        `json:"hasContents"`
	Hasmaterial      bool        `json:"hasMaterial"`
	Expiredattext    string      `json:"expiredAtText"`
	Totalstudents    int         `json:"totalStudents"`
	Haslecture       bool        `json:"hasLecture"`

	Raw []byte // 原始数据
}

type CourseInfo_Document struct {
	Index int    // 次序 1-based 代码中添加
	ID    string `json:"_id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
	URL   string `json:"url"`
	Key   string `json:"key"`
	Ext   string `json:"ext"`
}

func apiGetWanmenCourseInfo(courseId string) (*CourseInfo, error) {
	url := fmt.Sprintf("https://api.wanmen.org/4.0/content/v2/courses/%s", courseId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot build course info api request: %w", err)
	}
	req.Header = getHeaders()

	response, err := httpRequestWithAutoRetry(req)
	if err != nil {
		return nil, fmt.Errorf("cannot request course info api: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot request course info api (read body): %w", err)
	}

	var courseInfo CourseInfo

	err = json.Unmarshal(body, &courseInfo)
	if err != nil {
		return nil, fmt.Errorf("cannot request course info api (unmarshal json): %w", err)
	}

	courseInfo.Raw = body

	return &courseInfo, nil
}
