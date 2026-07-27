package main

import (
	"os"
	"path/filepath"
	"testing"
)

// handleReference 应优先读磁盘 references/（云同步 / 安装后的真实内容），磁盘没有才回退内嵌
func TestReadReferenceFile_DiskWins(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	refDir := filepath.Join(home, ".claude", "skills", "go-team-standards", "references")
	if err := os.MkdirAll(refDir, 0o755); err != nil {
		t.Fatal(err)
	}
	want := []byte("# 磁盘上的最新版本\n")
	if err := os.WriteFile(filepath.Join(refDir, "go-style.md"), want, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := readReferenceFile("go-style.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("应读磁盘内容，实际读到: %q", string(got))
	}

	// 磁盘没有的文件 → 回退内嵌（database.md 内嵌存在）
	got2, err := readReferenceFile("database.md")
	if err != nil || len(got2) == 0 {
		t.Fatalf("回退内嵌失败: err=%v len=%d", err, len(got2))
	}
}

// 云同步镜像清理：远端已移除的 .md/.png 应被删除，用户 custom-* 保留
func TestWriteStandardsMirror_DeletesOrphans(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	refDir := filepath.Join(home, ".claude", "skills", "go-team-standards", "references")
	if err := os.MkdirAll(refDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// 预置：一个遗留 orphan.md、一个用户 custom-mine.md
	_ = os.WriteFile(filepath.Join(refDir, "orphan.md"), []byte("旧"), 0o644)
	_ = os.WriteFile(filepath.Join(refDir, "custom-mine.md"), []byte("我的"), 0o644)

	// 远端只有 keep.md
	_, err := writeStandardsToReferences(map[string][]byte{"keep.md": []byte("新")})
	if err != nil {
		t.Fatal(err)
	}
	if _, e := os.Stat(filepath.Join(refDir, "keep.md")); e != nil {
		t.Fatal("keep.md 应写入")
	}
	if _, e := os.Stat(filepath.Join(refDir, "orphan.md")); !os.IsNotExist(e) {
		t.Fatal("orphan.md 应被镜像清理")
	}
	if _, e := os.Stat(filepath.Join(refDir, "custom-mine.md")); e != nil {
		t.Fatal("custom-mine.md 用户文件不应被删")
	}
}
