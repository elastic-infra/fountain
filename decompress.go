package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func decompressZip(f *os.File) error {
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	r, err := zip.NewReader(f, fi.Size())
	if err != nil {
		return err
	}

	for _, file := range r.File {
		path := filepath.Join(filepath.Dir(f.Name()), file.Name)

		if file.FileInfo().IsDir() {
			err = os.MkdirAll(path, file.Mode())
			if err != nil {
				return err
			}
		} else {
			err = createZippedFile(path, file.Mode(), file)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func createZippedFile(path string, mode os.FileMode, f *zip.File) error {
	src, err := f.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}
