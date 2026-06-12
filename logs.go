// 内存日志环形缓冲 —— 记录最近 200 条操作用于 UI 控制台排错
//
// 三种事件：
//   - kind=http  所有 /api/* 请求（中间件自动记）
//   - kind=shell 所有 exec.Command 执行（runIn 手动埋点）
//   - kind=api   Anthropic API 调用（callAnthropic 手动埋点）
//   - kind=info  业务关键步骤（手动埋点）
package main

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

func jsonDecoderPool(r io.Reader) *json.Decoder { return json.NewDecoder(r) }

type LogEntry struct {
	Time     time.Time `json:"time"`
	Kind     string    `json:"kind"`   // http | shell | api | info
	Op       string    `json:"op"`     // 简短动作名
	Detail   string    `json:"detail"` // 参数 / URL / 命令行
	Duration int64     `json:"duration_ms"`
	Status   string    `json:"status"` // ok | err
	Error    string    `json:"error,omitempty"`
}

type logRing struct {
	mu      sync.Mutex
	entries []LogEntry
	max     int
	pos     int
	size    int
}

var logBuffer = &logRing{max: 200, entries: make([]LogEntry, 0, 200)}

func (r *logRing) add(e LogEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.size < r.max {
		r.entries = append(r.entries, e)
		r.size++
	} else {
		r.entries[r.pos] = e
		r.pos = (r.pos + 1) % r.max
	}
}

func (r *logRing) snapshot() []LogEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]LogEntry, 0, r.size)
	if r.size < r.max {
		out = append(out, r.entries...)
	} else {
		for i := 0; i < r.max; i++ {
			out = append(out, r.entries[(r.pos+i)%r.max])
		}
	}
	return out
}

func (r *logRing) clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = r.entries[:0]
	r.pos = 0
	r.size = 0
}

// jsonFromReader 通用 JSON 解码（给 marketplace 等用）
func jsonFromReader(r io.Reader, out any) error {
	return jsonDecoderPool(r).Decode(out)
}

// logOp 供业务埋点
func logOp(kind, op, detail string, start time.Time, err error) {
	e := LogEntry{
		Time:     time.Now(),
		Kind:     kind,
		Op:       op,
		Detail:   trimStr(detail, 400),
		Duration: time.Since(start).Milliseconds(),
		Status:   "ok",
	}
	if err != nil {
		e.Status = "err"
		e.Error = trimStr(err.Error(), 400)
	}
	logBuffer.add(e)
}

// logInfo 不带耗时的简单记录
func logInfo(op, detail string) {
	logBuffer.add(LogEntry{
		Time: time.Now(), Kind: "info", Op: op, Detail: trimStr(detail, 400), Status: "ok",
	})
}

// ---------- HTTP 中间件 ----------

type statusWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}
func (sw *statusWriter) Write(b []byte) (int, error) {
	n, err := sw.ResponseWriter.Write(b)
	sw.size += n
	return n, err
}

// logMiddleware 包装任意 HTTP handler，自动记日志
func logMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 不记日志相关的端点，避免 UI 自刷新产生递归日志
		if r.URL.Path == "/api/logs" {
			h(w, r)
			return
		}
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: 200}
		h(sw, r)
		var err error
		if sw.status >= 400 {
			err = errorFromStatus(sw.status)
		}
		logOp("http", r.Method+" "+r.URL.Path, "", start, err)
	}
}

func errorFromStatus(code int) error {
	return &httpStatusErr{code: code}
}

type httpStatusErr struct{ code int }

func (e *httpStatusErr) Error() string { return http.StatusText(e.code) }

// ---------- handlers ----------

func handleLogs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"entries": logBuffer.snapshot(),
		"count":   logBuffer.size,
		"max":     logBuffer.max,
	})
}

func handleLogsClear(w http.ResponseWriter, r *http.Request) {
	logBuffer.clear()
	logInfo("clear", "日志已清空")
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
