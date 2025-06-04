package main

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestGetIgnoreList(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "gptignore")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "# comment\nlogs/\n\nassets/*.min.js\n"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	list, err := getIgnoreList(tmpFile.Name())
	if err != nil {
		t.Fatalf("getIgnoreList returned error: %v", err)
	}

	expected := []string{"logs/**", "assets/*.min.js"}
	if runtime.GOOS == "windows" {
		expected = []string{"logs\\**", "assets\\*.min.js"}
	}
	if !reflect.DeepEqual(list, expected) {
		t.Fatalf("expected %v, got %v", expected, list)
	}
}

func TestShouldIgnore_DefaultAndNested(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"node_modules", true},
		{filepath.Join("node_modules", "pkg", "index.js"), true},
		{filepath.Join("src", "node_modules", "index.js"), false},
	}

	for _, c := range cases {
		if got := shouldIgnore(c.path, nil); got != c.want {
			t.Errorf("shouldIgnore(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestShouldIgnore_WithPatterns(t *testing.T) {
	ignoreList := []string{"logs/**", "**/*.min.js", "docs/**/ignored/**"}
	cases := []struct {
		path string
		want bool
	}{
		{filepath.Join("logs", "app.log"), true},
		{filepath.Join("static", "js", "file.min.js"), true},
		{filepath.Join("docs", "api", "ignored", "readme.md"), true},
		{filepath.Join("docs", "api", "readme.md"), false},
	}

	for _, c := range cases {
		if got := shouldIgnore(c.path, ignoreList); got != c.want {
			t.Errorf("shouldIgnore(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}
