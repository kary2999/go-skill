package main

import "testing"

// 真实调 GitHub API 验证同步链路（需联网）
func TestGitHubSync_FetchAndDownload(t *testing.T) {
	sha, err := fetchLatestCommitSHA("https://github.com/kary2999/standards.git", "main")
	if err != nil {
		t.Fatalf("fetchLatestCommitSHA: %v", err)
	}
	if len(sha) != 40 {
		t.Fatalf("SHA 异常: %q", sha)
	}
	t.Logf("latest sha: %s", sha)

	zipData, err := downloadArchiveZip("https://github.com/kary2999/standards.git", "main")
	if err != nil {
		t.Fatalf("downloadArchiveZip: %v", err)
	}
	files, err := extractStandardsFromZip(zipData)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if _, ok := files["go-style.md"]; !ok {
		t.Fatalf("zip 里没有 go-style.md，拿到 %d 个文件", len(files))
	}
	t.Logf("提取到 %d 个规范文件", len(files))
}
