package repo_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/refaktor/rye-front/fynegen/generate/repo"
)

func TestRepo(t *testing.T) {
	path, err := repo.Get("test-out", "golang.org/x/crypto", "v0.23.0")
	if err != nil {
		t.Fatal(err)
	}

	{
		gotPath := filepath.ToSlash(path)
		wantPath := "test-out/golang.org/x/crypto@v0.23.0"
		if gotPath != wantPath {
			t.Fatalf("expected path %v, but got %v", wantPath, gotPath)
		}
	}

	{
		wantFile := filepath.Join(path, "ssh", "terminal", "terminal.go")
		if _, err := os.Stat(wantFile); err != nil {
			t.Fatalf("expected file %v to exist in archive, but not found", wantFile)
		}
	}

	// Windows tends to complain of simultaneous access (although all files were closed).
	time.Sleep(500 * time.Millisecond)

	if err := os.RemoveAll("test-out"); err != nil {
		t.Fatal(err)
	}
}
