package main

import (
	"os"
	"path/filepath"
)

func ListDirs(rootDir string) ([]TextureFile, error) {
	var files []TextureFile

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if path == rootDir {
			return nil
		}

		if info != nil && info.IsDir() && (info.Name() == "." || info.Name() == ".thumbnail") {
			return filepath.SkipDir
		}

		parentdir := filepath.Dir(path)

		files = append(files, TextureFile{
			Parent:   filepath.Base(parentdir),
			Filename: filepath.Base(path),
			Favorite: false,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
