package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

// POST /api/agents-init
// body: {"path": "/absolute/path/to/project"}
// 在项目根目录生成 AGENTS.md（已存在则跳过，不覆盖）
func handleAgentsInit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" {
		writeError(w, http.StatusBadRequest, "path 必填", nil)
		return
	}

	abs, err := filepath.Abs(req.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "路径无效", err)
		return
	}
	if _, err := os.Stat(abs); err != nil {
		writeError(w, http.StatusBadRequest, "目录不存在", err)
		return
	}

	dst := filepath.Join(abs, "AGENTS.md")

	// 已存在 → 不覆盖，告知用户
	if _, err := os.Stat(dst); err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":      true,
			"created": false,
			"file":    dst,
		})
		return
	}

	tmpl, err := readEmbedFile("assets/AGENTS.md.template")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读模板失败", err)
		return
	}
	if err := os.WriteFile(dst, tmpl, 0644); err != nil {
		writeError(w, http.StatusInternalServerError, "写文件失败", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"created": true,
		"file":    dst,
	})
}
