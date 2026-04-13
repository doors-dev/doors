// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package imports

import (
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

func cookModule(fs fs.FS) (string, error) {
	path, err := copyTemp(fs)
	if err != nil {
		return path, err
	}
	err = install(path)
	return path, err
}

func install(tempDir string) error {
	cmd := exec.Command("bun", "install")
	cmd.Dir = tempDir
	return cmd.Run()
}

var temps = make([]string, 0)

func clean() {
	for _, temp := range temps {
		os.RemoveAll(temp)
	}
}

func copyTemp(source fs.FS) (string, error) {
	tempDir, err := os.MkdirTemp("", "doors-test-")
	if err != nil {
		return "", err
	}
	err = fs.WalkDir(source, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		targetPath := filepath.Join(tempDir, path)
		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}
		srcFile, err := source.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		dstFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()
		_, err = io.Copy(dstFile, srcFile)
		return err
	})

	if err != nil {
		os.RemoveAll(tempDir) // Clean up on failure
		return "", err
	}
	temps = append(temps, tempDir)
	return tempDir, nil
}
