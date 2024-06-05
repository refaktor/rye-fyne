package repo

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const proxyURL = "https://proxy.golang.org/"

func proxyRequestURL(proxyURL, pkg string, path ...string) (string, error) {
	pkg = strings.ToLower(pkg)
	pkgElems := strings.Split(pkg, "/")

	u, err := url.Parse(proxyURL)
	if err != nil {
		return "", err
	}
	u = u.JoinPath(append(pkgElems, path...)...)

	return u.String(), nil
}

func pkgPath(pkg, semver string) string {
	pkg = strings.ToLower(pkg)
	pkgElems := strings.Split(pkg, "/")

	var pathElems []string
	pathElems = append(pathElems, pkgElems[:len(pkgElems)-1]...)
	pathElems = append(pathElems, pkgElems[len(pkgElems)-1]+"@"+semver)
	return filepath.Join(pathElems...)
}

func GetLatestVersion(pkg string) (string, error) {
	url, err := proxyRequestURL(proxyURL, pkg, "@latest")
	if err != nil {
		return "", err
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data struct {
		Version string
		Time    string
		Origin  *struct {
			VCS  string
			URL  string
			Ref  string
			Hash string
		}
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return "", err
	}

	return data.Version, nil
}

func Have(dstPath, pkg, semver string) (bool, string, error) {
	if semver == "" {
		v, err := GetLatestVersion(pkg)
		if err != nil {
			return false, "", err
		}
		semver = v
	}

	outPath := filepath.Join(dstPath, pkgPath(pkg, semver))

	if _, err := os.Stat(outPath); err == nil {
		return true, outPath, nil
	}
	return false, outPath, nil
}

func Get(dstPath, pkg, semver string) (string, error) {
	if semver == "" {
		v, err := GetLatestVersion(pkg)
		if err != nil {
			return "", err
		}
		semver = v
	}

	outPath := filepath.Join(dstPath, pkgPath(pkg, semver))

	if _, err := os.Stat(outPath); err == nil {
		return outPath, nil
	}

	zipURL, err := proxyRequestURL(proxyURL, pkg, "@v", semver+".zip")
	if err != nil {
		return "", err
	}

	resp, err := http.Get(zipURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return "", errors.New(string(data))
		}
		return "", fmt.Errorf("get %v: %v (%v)", zipURL, resp.Status, resp.StatusCode)
	}

	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	extractFile := func(f *zip.File) error {
		path := filepath.Join(dstPath, f.Name)
		if !strings.HasPrefix(path, filepath.Clean(dstPath)+string(os.PathSeparator)) {
			return fmt.Errorf("zip: illegal file path: %v", path)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, f.Mode()); err != nil {
				return err
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(path), f.Mode()); err != nil {
				return err
			}
			out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer out.Close()

			in, err := f.Open()
			if err != nil {
				return err
			}
			defer in.Close()

			if _, err := io.Copy(out, in); err != nil {
				return err
			}
		}
		return nil
	}
	for _, f := range archive.File {
		if err := extractFile(f); err != nil {
			return "", err
		}
	}

	return outPath, nil
}
