// Skill 自动化评估 —— 调 Claude API 跑测试用例
//
// 核心：把当前已安装的 SKILL.md + references 当作 system prompt，
//       把 TestCase.ViolationCode 作为 user 消息，看 AI 是否命中 Keywords
//
// prompt caching：system 块加 cache_control=ephemeral，20 个用例共享一份大 system，
//                 只首次 cache_creation_input_tokens 付全价，之后都是 cache_read（90% 折扣）
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---------- Anthropic API 类型 ----------

type anthropicReq struct {
	Model       string        `json:"model"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature *float64      `json:"temperature,omitempty"`
	System      []systemBlock `json:"system"`
	Messages    []anthMsg     `json:"messages"`
}

type systemBlock struct {
	Type         string        `json:"type"`
	Text         string        `json:"text"`
	CacheControl *cacheControl `json:"cache_control,omitempty"`
}

type cacheControl struct {
	Type string `json:"type"` // ephemeral
}

type anthMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResp struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ---------- 业务类型 ----------

type EvalRequest struct {
	Provider    string   `json:"provider"` // v1.7.10 起固定 clawnova
	APIKey      string   `json:"api_key"`
	APIBase     string   `json:"api_base,omitempty"` // clawnova 默认 https://clawnova.ai/api/v1
	Model       string   `json:"model"`
	Temperature *float64 `json:"temperature,omitempty"`
	CaseIDs     []string `json:"case_ids,omitempty"` // 空表示跑全部
}

type CaseResult struct {
	ID              string   `json:"id"`
	Rule            string   `json:"rule"`
	Category        string   `json:"category"`
	Pass            bool     `json:"pass"`
	MatchedKeywords []string `json:"matched_keywords"`
	MissedKeywords  []string `json:"missed_keywords"`
	HitCount        int      `json:"hit_count"`
	Required        int      `json:"required"`
	AIResponse      string   `json:"ai_response"`
	LatencyMS       int64    `json:"latency_ms"`
	CacheRead       int      `json:"cache_read_tokens"`
	CacheCreate     int      `json:"cache_create_tokens"`
	InputTokens     int      `json:"input_tokens"`
	OutputTokens    int      `json:"output_tokens"`
	Error           string   `json:"error,omitempty"`
}

type EvalResult struct {
	OK          bool         `json:"ok"`
	Total       int          `json:"total"`
	Passed      int          `json:"passed"`
	Cases       []CaseResult `json:"cases"`
	StartedAt   string       `json:"started_at"`
	TotalTimeMS int64        `json:"total_time_ms"`
	Error       string       `json:"error,omitempty"`
}

// ---------- Handlers ----------

func handleEvalCases(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"cases":     testCases,
		"models":    modelChoices, // 兼容旧前端
		"providers": listProviders(),
	})
}

func handleEvalRun(w http.ResponseWriter, r *http.Request) {
	var req EvalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request", err)
		return
	}
	if req.Provider == "" {
		req.Provider = "clawnova"
	}

	// v1.7.10: clawnova 走 ~/Library/Application Support/TeamStandards/eval-config.json
	// 前端不传 key（避免 key 在 HTTP 请求体里反复传输 / 落 console 日志）
	if req.Provider == "clawnova" {
		saved := loadEvalConfig()
		if saved.APIKey == "" {
			writeJSON(w, http.StatusOK, EvalResult{Error: "缺少 API Key（先在「AI 配置」段保存配置）"})
			return
		}
		req.APIKey = saved.APIKey
		if req.APIBase == "" {
			req.APIBase = saved.APIBase
		}
		if req.Model == "" {
			req.Model = saved.Model
		}
	} else if req.APIKey == "" {
		writeJSON(w, http.StatusOK, EvalResult{Error: "缺少 API Key"})
		return
	}

	if req.Model == "" {
		req.Model = "qwen3.5-27b-local"
	}

	// 加载当前已安装的 SKILL.md (+ references) 作为 system prompt。
	// clawnova 公司内部 16K 上下文模型 → 用 compact 模式（仅 SKILL.md），避免超长。
	var systemText string
	var err error
	if req.Provider == "clawnova" {
		systemText, err = buildSystemPromptCompact()
	} else {
		systemText, err = buildSystemPrompt()
	}
	if err != nil {
		writeJSON(w, http.StatusOK, EvalResult{Error: "读取 SKILL: " + err.Error()})
		return
	}

	// 过滤要跑的用例
	cases := testCases
	if len(req.CaseIDs) > 0 {
		set := map[string]bool{}
		for _, id := range req.CaseIDs {
			set[id] = true
		}
		var pick []TestCase
		for _, c := range testCases {
			if set[c.ID] {
				pick = append(pick, c)
			}
		}
		cases = pick
	}

	result := EvalResult{
		OK:        true,
		Total:     len(cases),
		Cases:     make([]CaseResult, 0, len(cases)),
		StartedAt: time.Now().Format(time.RFC3339),
	}

	start := time.Now()
	cfg := ProviderConfig{
		Provider: req.Provider, APIKey: req.APIKey,
		Model: req.Model, Temperature: req.Temperature,
		APIBase: req.APIBase,
	}
	// 16K 上下文模型 → 限制 max_tokens 给输出留够空间，又不撑爆窗口
	if req.Provider == "clawnova" {
		cfg.MaxTokens = 600
	}
	for _, tc := range cases {
		r := evalCaseWithProvider(cfg, systemText, tc)
		if r.Pass {
			result.Passed++
		}
		result.Cases = append(result.Cases, r)
	}
	result.TotalTimeMS = time.Since(start).Milliseconds()

	writeJSON(w, http.StatusOK, result)
}

// ---------- 评估核心 ----------

// evalCaseWithProvider 用统一 provider 抽象跑一条用例
func evalCaseWithProvider(cfg ProviderConfig, systemText string, tc TestCase) CaseResult {
	userPrompt := "请 review 以下 Go 代码，指出任何违反团队编码规范的地方。请按规范要求的三段式回复（规则 + 为什么 + 怎么改）：\n\n```go\n" + tc.ViolationCode + "\n```"

	start := time.Now()
	res := callByProvider(cfg, systemText, userPrompt)
	latency := time.Since(start).Milliseconds()

	if res.RawError != nil {
		return CaseResult{
			ID: tc.ID, Rule: tc.Rule, Category: tc.Category,
			Required: tc.MinMatch, Pass: false, Error: res.RawError.Error(),
			LatencyMS: latency,
		}
	}

	// 关键词匹配（小写、子串）
	aiLower := strings.ToLower(res.Text)
	var matched, missed []string
	for _, kw := range tc.Keywords {
		if strings.Contains(aiLower, strings.ToLower(kw)) {
			matched = append(matched, kw)
		} else {
			missed = append(missed, kw)
		}
	}

	return CaseResult{
		ID: tc.ID, Rule: tc.Rule, Category: tc.Category,
		Pass:            len(matched) >= tc.MinMatch,
		MatchedKeywords: matched, MissedKeywords: missed,
		HitCount: len(matched), Required: tc.MinMatch,
		AIResponse: res.Text, LatencyMS: latency,
		CacheRead: res.CacheReadTokens, CacheCreate: res.CacheCreateTokens,
		InputTokens: res.InputTokens, OutputTokens: res.OutputTokens,
	}
}

// callAnthropic 纯 HTTP 层；日志由上层 callByProvider 统一打，这里不重复
func callAnthropic(apiKey string, body anthropicReq) (*anthropicResp, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var out anthropicResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("bad response (HTTP %d): %s", resp.StatusCode, trimStr(string(raw), 500))
	}
	return &out, nil
}

// buildSystemPromptCompact 仅 SKILL.md，给 16K 上下文小模型用（如 clawnova qwen3.5-27b-local）
// 不拼接 references，避免超出 context window。
// 测试结果可能略弱于 full 模式（AI 看不到 references 细节），但 SKILL.md 里的铁律表已经够测主要规则。
func buildSystemPromptCompact() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	skillRoot := filepath.Join(home, ".claude", "skills", "go-team-standards")
	b, err := os.ReadFile(filepath.Join(skillRoot, "SKILL.md"))
	if err != nil {
		return "", fmt.Errorf("未安装 go-team-standards Skill，请先到 ⚡ 安装 Tab 装一次：%w", err)
	}
	return string(b), nil
}

// buildSystemPrompt 读取已安装的 SKILL.md + 所有 references（含 custom-*）
// 作为完整 system prompt
func buildSystemPrompt() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	skillRoot := filepath.Join(home, ".claude", "skills", "go-team-standards")

	var sb strings.Builder
	// SKILL.md
	b, err := os.ReadFile(filepath.Join(skillRoot, "SKILL.md"))
	if err != nil {
		return "", fmt.Errorf("未安装 go-team-standards Skill，请先到 ⚡ 安装 Tab 装一次：%w", err)
	}
	sb.Write(b)
	sb.WriteString("\n\n---\n\n# references 全文（评估用）\n\n")

	// 把所有 references/*.md 拼上，让 AI 像真实场景一样看到完整上下文
	refDir := filepath.Join(skillRoot, "references")
	entries, _ := os.ReadDir(refDir)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(refDir, e.Name()))
		if err != nil {
			continue
		}
		sb.WriteString("\n## references/" + e.Name() + "\n\n")
		sb.Write(data)
		sb.WriteString("\n")
	}

	return sb.String(), nil
}
