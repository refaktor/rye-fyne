package repo_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/refaktor/rye-front/fynegen/generate/repo"
)

func testRepo(t *testing.T, dir, pkg, semver, wantFile string) {
	path, err := repo.Get(dir, pkg, semver)
	if err != nil {
		t.Fatal(err)
	}

	if semver != "" && semver != "latest" {
		gotPath := filepath.ToSlash(path)
		wantPath := filepath.ToSlash(filepath.Join(dir, strings.ToLower(pkg)+"@"+semver))
		if gotPath != wantPath {
			t.Fatalf("expected path %v, but got %v", wantPath, gotPath)
		}
	}

	if wantFile != "" {
		wantFile := filepath.Join(path, wantFile)
		if _, err := os.Stat(wantFile); err != nil {
			t.Fatalf("expected file %v to exist in archive, but not found", wantFile)
		}
	}
}

func TestRepo(t *testing.T) {
	// Regular library
	testRepo(t, "test-out", "golang.org/x/crypto", "v0.23.0", "ssh/terminal/terminal.go")
	// Capital letters
	testRepo(t, "test-out", "github.com/BurntSushi/toml", "v1.3.2", "")
	// No go.mod
	testRepo(t, "test-out", "github.com/fogleman/gg", "", "gradient.go")
	// Latest version
	testRepo(t, "test-out", "github.com/BurntSushi/toml", "latest", "")

	// Windows tends to complain of simultaneous access (although all files were closed).
	time.Sleep(500 * time.Millisecond)

	if err := os.RemoveAll("test-out"); err != nil {
		t.Fatal(err)
	}
}
