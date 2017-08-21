// Copyright (c) 2017 Christian Funkhouser <christian.funkhouser@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package preppi

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/afero"
)

const testConfigFileName = "/boot/stuff/preppi.conf"

type testFile struct {
	Content []byte
	Mode    os.FileMode
	UID     int
	GID     int
}

var mapperTestCases = []*struct {
	CaseName  string
	Files     map[string]*testFile
	Mapper    *Mapper
	WantError bool
}{
	{
		CaseName: "Success Casse",
		Files: map[string]*testFile{
			"/boot/stuff/something": &testFile{
				[]byte("This is something"),
				0644,
				500,
				500,
			},
			"/boot/stuff/something-else": &testFile{
				[]byte("This is something else"),
				0644,
				500,
				500,
			},
		},
		Mapper: &Mapper{
			Mappings: []*Mapping{
				&Mapping{
					Source:      "/boot/stuff/something",
					Destination: "/etc/something",
					Mode:        0640,
					UID:         0,
					GID:         5,
				},
				&Mapping{
					Source:      "/boot/stuff/something-else",
					Destination: "/etc/something.d/else",
					Mode:        0640,
					UID:         0,
					GID:         5,
				},
			},
		},
	},
	{
		CaseName: "Source Doesn't Exist",
		Files:    map[string]*testFile{},
		Mapper: &Mapper{
			Mappings: []*Mapping{
				&Mapping{
					Source:      "/boot/stuff/doesnt-exist",
					Destination: "/etc/something",
					Mode:        0640,
					UID:         0,
					GID:         5,
				},
			},
		},
		WantError: true,
	},
}

func setUpFilesystemForTest(t *testing.T, fs afero.Fs, files map[string]*testFile) {
	for name, tf := range files {
		f, err := fs.OpenFile(name, os.O_CREATE, tf.Mode)
		if err != nil {
			t.Fatalf("Couldn't set up test with file %q: %v", name, err)
		}
		if _, err := io.Copy(f, bytes.NewBuffer(tf.Content)); err != nil {
			t.Fatalf("Couldn't set content for test file %q: %v", name, err)
		}
	}
}

func TestMapper(t *testing.T) {
	origPreppiFS := preppiFS
	defer func() { preppiFS = origPreppiFS }()

	for _, tt := range mapperTestCases {
		preppiFS = afero.NewMemMapFs()
		setUpFilesystemForTest(t, preppiFS, tt.Files)

		err := tt.Mapper.Apply()
		if err != nil && !tt.WantError {
			t.Errorf("%v: wanted no error, got: %v", tt.CaseName, err)
		} else if err == nil && tt.WantError {
			t.Errorf("%v: wanted error, got none", tt.CaseName)
		} else {
			// Make sure we end up in the intended end state.
			for _, m := range tt.Mapper.Mappings {
				f := tt.Files[m.Source]
				s, err := preppiFS.Stat(m.Destination)
				if err != nil {
					t.Errorf("%v: error stat-ing %q: %v", tt.CaseName, m.Source, err)
					break
				}
				if s.Mode() != m.Mode {
					t.Errorf("%v: wanted %v, got %v for %q", tt.CaseName, f.Mode, s.Mode(), err)
				}
			}
		}
	}
}
