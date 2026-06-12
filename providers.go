// 多 Provider 支持 —— 让测试可以用免费 AI
//
// 统一接口：传入 system + user message 和配置，返回文本 + token 统计
// 支持：Claude (Anthropic) / Gemini / Groq / OpenRouter / DeepSeek / Ollama
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// APIResult 所有 provider 调用的统一返回
type APIResult struct {
	Text              string
	InputTokens       int
	OutputTokens      int
	CacheReadTokens   int
	CacheCreateTokens int
	RawError          error
}

// ProviderConfig 单次调用的配置
type ProviderConfig struct {
	Provider    string   // claude | gemini | groq | openrouter | deepseek | ollama | clawnova
	APIKey      string   // Ollama 时作 baseURL（可选）
	Model       string
	Temperature *float64
	APIBase     string // 仅 clawnova 用：自定义 base URL，空则用默认 https://clawnova.ai/api/v1
	MaxTokens   int    // 0 表示用默认（1200）。clawnova 16K 上下文模型建议 600
}

// callByProvider 统一分发 + 统一日志
// 所有 provider 的成功/失败都会进日志控制台，不再 Claude 单独记
func callByProvider(cfg ProviderConfig, systemText, userText string) APIResult {
	start := time.Now()
	var res APIResult
	switch strings.ToLower(cfg.Provider) {
	case "", "claude":
		res = callClaudeProvider(cfg, systemText, userText)
	case "gemini":
		res = callGeminiProvider(cfg, systemText, userText)
	case "groq":
		res = callOAICompat("https://api.groq.com/openai/v1", cfg, systemText, userText, true)
	case "openrouter":
		res = callOAICompat("https://openrouter.ai/api/v1", cfg, systemText, userText, true)
	case "deepseek":
		res = callOAICompat("https://api.deepseek.com/v1", cfg, systemText, userText, true)
	case "ollama":
		res = callOllamaProvider(cfg, systemText, userText)
	case "clawnova":
		// 公司内部 OpenAI 兼容服务（v1.7.10 起作为 Skill 测试唯一 provider）
		// Base URL 走配置（可被 cfg.APIBase 覆盖），默认 https://clawnova.ai/api/v1
		base := cfg.APIBase
		if base == "" {
			base = "https://clawnova.ai/api/v1"
		}
		res = callOAICompat(base, cfg, systemText, userText, true)
	default:
		res = APIResult{RawError: errors.New("unknown provider: " + cfg.Provider)}
	}

	// 统一日志（含错误）—— 日志控制台 🪵 可看
	providerName := cfg.Provider
	if providerName == "" {
		providerName = "claude"
	}
	detail := fmt.Sprintf("model=%s in=%d out=%d",
		cfg.Model, res.InputTokens, res.OutputTokens)
	if res.CacheReadTokens > 0 || res.CacheCreateTokens > 0 {
		detail += fmt.Sprintf(" cache_r=%d cache_c=%d", res.CacheReadTokens, res.CacheCreateTokens)
	}
	logOp("api", providerName+".generate", detail, start, res.RawError)
	return res
}

// ============ Claude (Anthropic) ============

func callClaudeProvider(cfg ProviderConfig, systemText, userText string) APIResult {
	reqBody := anthropicReq{
		Model:       cfg.Model,
		MaxTokens:   1200,
		Temperature: cfg.Temperature,
		System: []systemBlock{{
			Type: "text", Text: systemText,
			CacheControl: &cacheControl{Type: "ephemeral"},
		}},
		Messages: []anthMsg{{Role: "user", Content: userText}},
	}
	resp, err := callAnthropic(cfg.APIKey, reqBody)
	if err != nil {
		return APIResult{RawError: err}
	}
	if resp.Error != nil {
		return APIResult{RawError: fmt.Errorf("%s: %s", resp.Error.Type, resp.Error.Message)}
	}
	var text string
	for _, c := range resp.Content {
		if c.Type == "text" {
			text += c.Text
		}
	}
	return APIResult{
		Text:              text,
		InputTokens:       resp.Usage.InputTokens,
		OutputTokens:      resp.Usage.OutputTokens,
		CacheReadTokens:   resp.Usage.CacheReadInputTokens,
		CacheCreateTokens: resp.Usage.CacheCreationInputTokens,
	}
}

// ============ OpenAI 兼容 (Groq / OpenRouter / DeepSeek / Ollama) ============

type oaiReq struct {
	Model       string   `json:"model"`
	Messages    []oaiMsg `json:"messages"`
	Temperature *float64 `json:"temperature,omitempty"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Stream      bool     `json:"stream"`
}

type oaiMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type oaiResp struct {
	Choices []struct {
		Message oaiMsg `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
		Code    any    `json:"code,omitempty"`
	} `json:"error,omitempty"`
}

func callOAICompat(baseURL string, cfg ProviderConfig, systemText, userText string, needAuth bool) APIResult {
	maxTok := cfg.MaxTokens
	if maxTok <= 0 {
		maxTok = 1200
	}
	reqBody := oaiReq{
		Model: cfg.Model,
		Messages: []oaiMsg{
			{Role: "system", Content: systemText},
			{Role: "user", Content: userText},
		},
		Temperature: cfg.Temperature,
		MaxTokens:   maxTok,
		Stream:      false,
	}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", baseURL+"/chat/completions", bytes.NewReader(b))
	if err != nil {
		return APIResult{RawError: err}
	}
	req.Header.Set("content-type", "application/json")
	if needAuth {
		if cfg.APIKey == "" {
			return APIResult{RawError: errors.New("缺少 API Key")}
		}
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}
	// OpenRouter 推荐加这俩 header 便于统计
	if strings.Contains(baseURL, "openrouter.ai") {
		req.Header.Set("HTTP-Referer", "https://team-standards.local")
		req.Header.Set("X-Title", "Team Standards Skill Eval")
	}

	// 本地 Ollama 大模型可能跑得慢，超时放宽
	timeout := 120 * time.Second
	if strings.Contains(baseURL, "11434") {
		timeout = 600 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	httpResp, err := client.Do(req)
	if err != nil {
		return APIResult{RawError: err}
	}
	defer httpResp.Body.Close()
	raw, _ := io.ReadAll(httpResp.Body)

	var out oaiResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return APIResult{RawError: fmt.Errorf("bad response (HTTP %d): %s", httpResp.StatusCode, trimStr(string(raw), 400))}
	}
	if out.Error != nil {
		return APIResult{RawError: fmt.Errorf("%s: %s", out.Error.Type, out.Error.Message)}
	}
	if httpResp.StatusCode >= 400 {
		return APIResult{RawError: fmt.Errorf("HTTP %d: %s", httpResp.StatusCode, trimStr(string(raw), 300))}
	}
	text := ""
	if len(out.Choices) > 0 {
		text = out.Choices[0].Message.Content
	}
	return APIResult{
		Text:         text,
		InputTokens:  out.Usage.PromptTokens,
		OutputTokens: out.Usage.CompletionTokens,
	}
}

// ============ Ollama（本地）============

func callOllamaProvider(cfg ProviderConfig, systemText, userText string) APIResult {
	baseURL := strings.TrimSuffix(cfg.APIKey, "/")
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1"
	}
	// Ollama OpenAI 兼容模式不需要 auth
	return callOAICompat(baseURL, cfg, systemText, userText, false)
}

// ============ Gemini ============

type geminiReq struct {
	Contents          []geminiContent  `json:"contents"`
	SystemInstruction *geminiContent   `json:"systemInstruction,omitempty"`
	GenerationConfig  *geminiGenConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
}

type geminiResp struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func callGeminiProvider(cfg ProviderConfig, systemText, userText string) APIResult {
	if cfg.APIKey == "" {
		return APIResult{RawError: errors.New("缺少 Gemini API Key")}
	}
	reqBody := geminiReq{
		SystemInstruction: &geminiContent{Parts: []geminiPart{{Text: systemText}}},
		Contents: []geminiContent{
			{Role: "user", Parts: []geminiPart{{Text: userText}}},
		},
		GenerationConfig: &geminiGenConfig{
			Temperature:     cfg.Temperature,
			MaxOutputTokens: 1200,
		},
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		cfg.Model, cfg.APIKey)
	b, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Set("content-type", "application/json")
	client := &http.Client{Timeout: 120 * time.Second}
	httpResp, err := client.Do(req)
	if err != nil {
		return APIResult{RawError: err}
	}
	defer httpResp.Body.Close()
	raw, _ := io.ReadAll(httpResp.Body)

	var out geminiResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return APIResult{RawError: fmt.Errorf("bad response (HTTP %d): %s", httpResp.StatusCode, trimStr(string(raw), 400))}
	}
	if out.Error != nil {
		return APIResult{RawError: fmt.Errorf("gemini %d: %s", out.Error.Code, out.Error.Message)}
	}
	text := ""
	if len(out.Candidates) > 0 && len(out.Candidates[0].Content.Parts) > 0 {
		text = out.Candidates[0].Content.Parts[0].Text
	}
	return APIResult{
		Text:         text,
		InputTokens:  out.UsageMetadata.PromptTokenCount,
		OutputTokens: out.UsageMetadata.CandidatesTokenCount,
	}
}

// ============ Provider 元数据（给 UI 下拉）============

type ProviderMeta struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Models      []string `json:"models"`
	KeyLabel    string   `json:"key_label"`
	KeyPlaceholder string `json:"key_placeholder"`
	KeyHint     string   `json:"key_hint"`
	RequiresKey bool     `json:"requires_key"`
	FreeNote    string   `json:"free_note"`
	Recommended bool     `json:"recommended"`
}

// v1.7.10 起：测试 Skill 有效性只允许一个 AI 配置（clawnova，公司本地大模型）。
// 历史上的 gemini/groq/claude/openrouter/deepseek/ollama 实现保留在 callByProvider 里以备将来用，
// 但 UI 只暴露 clawnova。如果你切回多 provider，把下面的注释段恢复即可。
func listProviders() []ProviderMeta {
	return []ProviderMeta{
		{
			ID: "clawnova", Name: "Clawnova（公司本地大模型）",
			// 教程截图列的 chat 模型（v1.7.12 起）：
			//   - qwen3.5-27b-local   推荐用，27B 通用
			//   - minimax-m2.5-local  M 系
			//   - step-3.5-local      step 系
			//   - gemma-4-31b         多模态（支持图片/视频）
			// 不列入：
			//   - qwen3-em-8b   embeddings 专用，不能 chat
			//   - zimage-turbo-local  图片生成专用
			//   - qwen          v1.7.10 我误加，教程没有，删
			Models: []string{
				"qwen3.5-27b-local",
				"minimax-m2.5-local",
				"step-3.5-local",
				"gemma-4-31b",
			},
			KeyLabel: "Clawnova API Key", KeyPlaceholder: "kgb_...",
			KeyHint:     "公司内部签发，找内部管理员或在 IDP 平台领取",
			RequiresKey: true, FreeNote: "✅ 公司内部，免费 · OpenAI 兼容", Recommended: true,
		},
	}
}
