package repo

import (
	"archive/zip"
	"bytes"
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

func Get(dstPath, pkg, semver string) (string, error) {
	pkgElems := strings.Split(pkg, "/")

	var outPath string
	{
		var pathElems []string
		pathElems = append(pathElems, dstPath)
		pathElems = append(pathElems, pkgElems[:len(pkgElems)-1]...)
		pathElems = append(pathElems, pkgElems[len(pkgElems)-1]+"@"+semver)
		outPath = filepath.Join(pathElems...)
	}

	if _, err := os.Stat(outPath); err == nil {
		return outPath, nil
	}

	var zipURL string
	{
		u, err := url.Parse(proxyURL)
		if err != nil {
			return "", err
		}
		u = u.JoinPath(append(pkgElems, "@v", semver+".zip")...)
		zipURL = u.String()
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
