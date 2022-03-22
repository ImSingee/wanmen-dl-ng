package main

import (
	"bytes"
	"fmt"
	"github.com/rclone/rclone/lib/terminal"
	"strings"
	"time"
)

type Dashboard struct {
	courseID   string
	courseName string

	//chaptersDone  int
	//chaptersTotal int
	lecturesDone  int
	lecturesTotal int

	documentsDone  int
	documentsTotal int

	errorsCount int64

	workers []workerStat

	startAt   time.Time
	lastLines int
	actionCh  chan DashboardAction
	logCh     chan LogMessage

	nextRoundFirst int

	closeCh chan struct{}
}

type workerStat struct {
	id     int
	desc   string
	status string // 当前的状态，可能为 空 / prepare / downloading / converting / error / done / quit
	// 下面两个状态只有 downloading 有意义
	done  int
	total int
}

func NewDashboard(courseId, courseName string, workersCount int) *Dashboard {
	workers := make([]workerStat, workersCount)
	for i := range workers {
		workers[i].id = i + 1
	}

	return &Dashboard{
		courseID:   courseId,
		courseName: courseName,
		workers:    workers,
		startAt:    time.Now(),
		actionCh:   make(chan DashboardAction, 256),
		logCh:      make(chan LogMessage, 256),
		closeCh:    make(chan struct{}),
	}
}

type DashboardAction struct {
	Method string
	Args   []interface{}
}

type LogMessage struct {
	Time    time.Time
	Level   int
	Message string
}

const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

func (m *LogMessage) String() string {
	b := strings.Builder{}
	b.Grow(16 + len(m.Message))

	switch m.Level {
	case LogLevelDebug:
		b.WriteString("[DEBUG] ")
	case LogLevelInfo:
		b.WriteString(terminal.BlueFg)
		b.WriteString("[INFO] ")
		b.WriteString(terminal.Reset)
	case LogLevelWarning:
		b.WriteString(terminal.YellowFg)
		b.WriteString("[WARNING] ")
		b.WriteString(terminal.Reset)
	case LogLevelError:
		b.WriteString(terminal.RedFg)
		b.WriteString("[ERROR] ")
		b.WriteString(terminal.Reset)
	}

	b.WriteString(m.Message)

	return b.String()
}

func (d *Dashboard) Start() chan<- DashboardAction {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		d.printHeader()

		for {
			select {
			case action, ok := <-d.actionCh:
				if !ok { // stop
					d.rePaint(true)
					return
				}

				d.actionHandler(action.Method, action.Args...)
			case <-ticker.C:
				d.rePaint(false)
			}
		}
	}()

	return d.actionCh
}

func (d *Dashboard) Close() {
	close(d.actionCh)
	<-d.closeCh
}

func (d *Dashboard) actionHandler(method string, params ...interface{}) {
	switch method {
	case "init-lectures": // courseInfo
		info := params[0].(*CourseLectures)
		//d.chaptersTotal = len(info.Chapters)
		c := 0
		for _, chapter := range info.Chapters {
			c += len(chapter.Children)
		}
		d.lecturesTotal = c
	case "init-documents": // []*CourseInfo_Document
		documents := params[0].([]*CourseInfo_Document)
		d.documentsTotal = len(documents)
	case "skip":
		// 整个课程被跳过
		d.logCh <- LogMessage{Level: LogLevelInfo, Message: "Course already saved, auto skip (You can delete .meta/DONE to re-download the course)"}
	case "skip-task": // workerId, DownloadTask
		// 单个任务被跳过
		task := params[1].(*DownloadTask)
		if task.Course != nil {
			d.lecturesDone++
		} else {
			d.documentsDone++
		}
	case "start": // workerId, DownloadTask
		// 任务开始
		workerId := params[0].(int)
		task := params[1].(*DownloadTask)
		d.workers[workerId-1].status = "prepare"
		d.workers[workerId-1].desc = task.Desc()
	case "lecture": // workerId, task,  subMethod, subParams
		// 课程任务下载进度汇报
		workerId := params[0].(int)
		//task := params[1].(*DownloadTask)
		subMethod := params[2].(string)
		subParams := params[3].([]interface{})

		switch subMethod {
		case "downloading": // i, N
			done := subParams[0].(int)
			total := subParams[1].(int)

			d.workers[workerId-1].status = "downloading"
			d.workers[workerId-1].done = done
			d.workers[workerId-1].total = total
		case "ffmpeg_start":
			d.workers[workerId-1].status = "converting"
		case "ffmpeg_error": // output
			ffmpegOutput := subParams[0].(string)

			d.logCh <- LogMessage{Level: LogLevelError, Message: fmt.Sprintf("ffmpeg error on %s: %s", d.workers[workerId-1].desc, ffmpegOutput)}

			d.workers[workerId-1].status = "error"
		case "ffmpeg_done":
			d.workers[workerId-1].status = "done"
		case "retry": // i, err
			times := subParams[0].(int)
			err := subParams[1].(error)

			d.logCh <- LogMessage{Level: LogLevelWarning, Message: fmt.Sprintf("Download %s fail, retry lower level (level %d fail): %v", d.workers[workerId-1].desc, times, err)}
		default:
			d.logCh <- LogMessage{Level: LogLevelDebug, Message: fmt.Sprintf("Unknown sub method %s for course", subMethod)}
		}
	case "doc": // workerId, task,  subMethod, subParams
		// 文档任务下载进度汇报
		workerId := params[0].(int)
		//task := params[1].(*DownloadTask)
		subMethod := params[2].(string)
		subParams := params[3].([]interface{})
		switch subMethod {
		case "downloading": // i, N
			done := subParams[0].(int)
			total := subParams[1].(int)

			d.workers[workerId-1].status = "downloading"
			d.workers[workerId-1].done = done
			d.workers[workerId-1].total = total
		default:
			d.logCh <- LogMessage{Level: LogLevelDebug, Message: fmt.Sprintf("Unknown sub method %s for doc", subMethod)}
		}
	case "done": // workerId, task
		// 任务完成
		workerId := params[0].(int)
		// td := params[1].(*DownloadTask)
		d.workers[workerId-1].status = "done"
		d.lecturesDone++
	case "error": // workerId, DownloadTask, err
		// 任务异常
		workerId := params[0].(int)
		td := params[1].(*DownloadTask)
		err := params[2].(error)

		d.errorsCount++
		d.logCh <- LogMessage{Level: LogLevelError, Message: fmt.Sprintf("%v (%v, %v)", err, workerId, td.Path())}
		d.workers[workerId-1].status = "error"
		d.lecturesDone++
	case "quit": // workerId
		// worker 停止
		workerId := params[0].(int)
		d.workers[workerId-1].status = "quit"
	default: // unknown method
		d.logCh <- LogMessage{Level: LogLevelDebug, Message: fmt.Sprintf("Unknown method: %s", method)}
	}
}

func (d *Dashboard) printHeader() {
	fmt.Println("==============================")
	fmt.Println("           1MAN DL")
	fmt.Println("==============================")
	fmt.Println()
	fmt.Printf("Course: %s (%s)\n", d.courseName, d.courseID)
	fmt.Println()
}

// rePaint 非并发安全
func (d *Dashboard) rePaint(last bool) {
	w, h := terminal.GetSize()

	stats := strings.TrimSpace(d.StatsString(h))

	var log string
	if last {
		log = strings.TrimSpace(d.FullLogString())
	} else {
		log = strings.TrimSpace(d.LogString())
	}
	_ = log

	var buf bytes.Buffer
	out := func(s string) {
		_, _ = buf.WriteString(s)
	}

	if log != "" { // 存在 log，写至上方
		out("\n")
		out(terminal.MoveUp)
	}

	// 清空原有的信息
	for i := 0; i < d.lastLines-1; i++ {
		out(terminal.EraseLine)
		out(terminal.MoveUp)
	}
	out(terminal.EraseLine)
	out(terminal.MoveToStartOfLine)

	// 写入日志（如果有）
	if log != "" { // 存在 log，写至上方
		out(terminal.EraseLine)
		out(log + "\n")
	}

	// 将 stats 信息输出（定长）
	fixedLines := strings.Split(stats, "\n")
	d.lastLines = len(fixedLines)
	for i, line := range fixedLines {
		if len(line) > w {
			line = line[:w]
		}
		out(line)
		if i != d.lastLines-1 {
			out("\n")
		}
	}

	// 将内容输出
	terminal.Write(buf.Bytes())

	if last {
		close(d.closeCh)
	}
}

// FullLogString 输出所有剩余的日志
// 这一方法只能被调用一次
func (d *Dashboard) FullLogString() string {
	close(d.logCh)

	b := strings.Builder{}
	for l := range d.logCh {
		b.WriteString(l.String())
		b.WriteString("\n")
	}
	return b.String()
}

func (d *Dashboard) LogString() string {
	b := strings.Builder{}

	// 最多输出 50 条（防止输出过多）
outer:
	for i := 0; i < 50; i++ {
		select {
		case l, ok := <-d.logCh:
			if !ok {
				break outer
			}

			b.WriteString(l.String())
			b.WriteString("\n")
		default: // not found now
			break outer
		}
	}

	return b.String()
}

func (d *Dashboard) StatsString(maxLines int) string {
	b := strings.Builder{}

	b.WriteString(fmt.Sprintf("Download: %d/%d Lectures %d/%d Docs\n",
		d.lecturesDone, d.lecturesTotal, d.documentsDone, d.documentsTotal))
	b.WriteString(fmt.Sprintf("Time: %s elaspsed\n", time.Since(d.startAt)))

	if d.errorsCount != 0 {
		b.WriteString(fmt.Sprintf("Errors: %d\n", d.errorsCount))
	}

	// 上面使用了 3 行，留给 worker 的可用行数为 maxLines - 3
	// 为了防止奇奇怪怪的问题，这里用 -4

	b.WriteString("\n")
	for _, stat := range d.showWorkers(maxLines - 4) {
		id := stat.id
		switch stat.status {
		case "", "error", "done":
			b.WriteString(fmt.Sprintf("* [%d] LEISURE\n", id))
		case "prepare":
			b.WriteString(fmt.Sprintf("* [%d] PREPARE %s\n", id, stat.desc))
		case "downloading":
			b.WriteString(fmt.Sprintf("* [%d] %d/%d %s\n", id, stat.done, stat.total, stat.desc))
		case "converting":
			b.WriteString(fmt.Sprintf("* [%d] CONVERTING %s\n", id, stat.desc))
		case "quit":
			b.WriteString(fmt.Sprintf("* [%d] SHUTDOWN\n", id))
		default: // unknown
			b.WriteString(fmt.Sprintf("* [%d] UNKNOWN %s %s\n", id, stat.status, stat.desc))
		}
	}

	return b.String()
}

func (d *Dashboard) showWorkers(max int) []workerStat {
	if max <= 0 {
		max = 1
	}
	if max >= len(d.workers) {
		return d.workers
	}

	// 每次只展示一定数量的 worker，防止溢出屏幕
	thisRoundFirst := d.nextRoundFirst
	nextRoundFirst := thisRoundFirst + max

	if nextRoundFirst < len(d.workers) {
		d.nextRoundFirst = nextRoundFirst
		return d.workers[thisRoundFirst:nextRoundFirst]
	} else if nextRoundFirst == len(d.workers) {
		d.nextRoundFirst = 0
		return d.workers[thisRoundFirst:]
	} else { // 出现了溢出
		d.nextRoundFirst = nextRoundFirst - len(d.workers)
		result := make([]workerStat, 0, max)
		result = append(result, d.workers[thisRoundFirst:]...)
		result = append(result, d.workers[:d.nextRoundFirst]...)
		return result
	}
}
