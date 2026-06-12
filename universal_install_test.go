package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUniversalInstall_FreshProject_CreatesAllAdapters(t *testing.T) {
	dir := t.TempDir()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/universal/install",
		strings.NewReader(`{"path":"`+dir+`"}`))
	handleUniversalInstall(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP %d: %s", rec.Code, rec.Body.String())
	}
	var res struct {
		OK      bool     `json:"ok"`
		Created []string `json:"created"`
		Failed  []string `json:"failed"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if !res.OK || len(res.Failed) != 0 {
		t.Fatalf("failed: %+v", res.Failed)
	}
	// AGENTS.md + 7 家适配
	if len(res.Created) != 1+len(universalAdapters) {
		t.Fatalf("created %d, want %d: %v", len(res.Created), 1+len(universalAdapters), res.Created)
	}
	for _, a := range universalAdapters {
		fp := filepath.Join(dir, filepath.FromSlash(a.Rel))
		b, err := os.ReadFile(fp)
		if err != nil {
			t.Fatalf("%s 未创建: %v", a.Rel, err)
		}
		if !strings.Contains(string(b), "AGENTS.md") {
			t.Errorf("%s 内容未指向 AGENTS.md", a.Rel)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err != nil {
		t.Fatal("AGENTS.md 未创建")
	}
}

func TestUniversalInstall_ExistingFiles_NotOverwritten(t *testing.T) {
	dir := t.TempDir()
	own := []byte("# 我自己的 CLAUDE.md\n")
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), own, 0o644); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/universal/install",
		strings.NewReader(`{"path":"`+dir+`"}`))
	handleUniversalInstall(rec, req)

	var res struct {
		Skipped []string `json:"skipped"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	found := false
	for _, s := range res.Skipped {
		if strings.HasPrefix(s, "CLAUDE.md") {
			found = true
		}
	}
	if !found {
		t.Fatalf("CLAUDE.md 应被跳过: %v", res.Skipped)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if string(b) != string(own) {
		t.Fatal("CLAUDE.md 被覆盖了")
	}
}

func TestUniversalInstall_BadPath(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/universal/install",
		strings.NewReader(`{"path":"/nonexistent/dir/xyz"}`))
	handleUniversalInstall(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}
