/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// CopyDirectory copies source to destination recursively.
func CopyDirectory(t *testing.T, source string, destination string) {
	t.Helper()

	files, err := ioutil.ReadDir(source)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		s := filepath.Join(source, f.Name())
		d := filepath.Join(destination, f.Name())

		if m := f.Mode(); m&os.ModeSymlink != 0 {
			CopySymlink(t, s, d)
		} else if f.IsDir() {
			CopyDirectory(t, s, d)
		} else {
			CopyFile(t, s, d)
		}
	}
}
