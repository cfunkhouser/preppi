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

// Testing a real-world OS can be done like this:
//
//   dir, err := ioutil.TempDir("", "TestNameOrWhatever")
//   if err != nil {
//     t.Fatal(err)
//   }
//   ofs := NewOsFs()
//   preppiFS = NewBasePathFs(ofs, dir)
//   defer os.RemoveAll(dir)
//   setUpFilesystemForTest(t, preppiFS, files)
//   ...
//
// I've not added any tests like this, because a test on the actual filesystem
// depends on system configuration, and is unlikely to work well between any two
// arbitrary systems. We'll need to eventually establish some integration tests.
// TODO(cfunkhouser): integration tests.

package preppi

import (
	"bytes"
	"io"
	"os"
	"path"
	"testing"
)

type testFile struct {
	Content []byte
	Mode    os.FileMode
	DirMode os.FileMode
	UID     int
	GID     int
}

// setUpFilesystemForTest adds files - as described by a mapping of file name to
// testFile structs - to a file system. If you pass an OsFs to this function, be
// prepared for the consequences.
func setUpFilesystemForTest(t *testing.T, fs Fs, files map[string]*testFile) {
	for name, tf := range files {
		dir := path.Dir(name)
		if err := fs.MkdirAll(dir, tf.DirMode); err != nil {
			t.Fatalf("Couldn't set up test directory %q: %v", dir, err)
		}
		f, err := fs.OpenFile(name, os.O_CREATE|os.O_WRONLY, tf.Mode)
		if err != nil {
			t.Fatalf("Couldn't set up test with file %q: %v", name, err)
		}
		if _, err := io.Copy(f, bytes.NewBuffer(tf.Content)); err != nil {
			t.Fatalf("Couldn't set content for test file %q: %v", name, err)
		}
	}
}

func TestDestinationExists(t *testing.T) {
	origPreppiFS := preppiFS
	defer func() { preppiFS = origPreppiFS }()
	preppiFS = NewMemMapFs()

	files := map[string]*testFile{
		"/some/file": &testFile{
			[]byte("This is something"),
			0644,
			0755,
			500,
			500,
		},
	}
	setUpFilesystemForTest(t, preppiFS, files)

	for _, tt := range []struct {
		name    string
		mapping *Mapping
		want    bool
	}{
		{
			name:    "doesn't exist",
			mapping: &Mapping{"/whatever", "/doesnt/exists", 0640, 0750, 500, 500, false},
		},
		{
			name:    "exists",
			mapping: &Mapping{"/whatever", "/some/file", 0640, 0750, 500, 500, false},
			want:    true,
		},
	} {
		got, err := tt.mapping.destinationExists()
		if err != nil {
			t.Fatalf("wanted no error, got: %v", err)
		}
		if got != tt.want {
			t.Errorf("%v: wanted %v, got %v", tt.name, tt.want, got)
		}
	}
}

func TestShouldCopy(t *testing.T) {
	origPreppiFS := preppiFS
	defer func() { preppiFS = origPreppiFS }()
	preppiFS = NewMemMapFs()

	files := map[string]*testFile{
		"/exists/should/copy": &testFile{
			[]byte("This is not the source file."),
			0644,
			0755,
			500,
			500,
		},
		"/exists/skipped": &testFile{
			[]byte("Here we are extending into shooting stars"),
			0644,
			0755,
			500,
			500,
		},
		"/exists/cant/clobber": &testFile{
			[]byte("This is not the source file."),
			0644,
			0755,
			500,
			500,
		},
	}
	setUpFilesystemForTest(t, preppiFS, files)

	// Hash from a file identical to /exists/skipped in the test filesystem
	srcCksm := []byte{
		175, 27, 219, 187, 224, 133, 38, 183, 254, 243, 232, 251, 57, 249, 11, 97, 140, 82, 253, 208, 209, 21, 224, 225,
		112, 105, 18, 224, 67, 79, 5, 154}

	for _, tt := range []struct {
		name    string
		mapping *Mapping
		want    bool
		wantErr error
	}{
		{
			name:    "exists, should copy",
			mapping: &Mapping{"/whatever", "/exists/should/copy", 0640, 0750, 500, 500, true},
			want:    true,
		},
		{
			name:    "exists, skipped",
			mapping: &Mapping{"/whatever", "/exists/skipped", 0640, 0750, 500, 500, true},
		},
		{
			name:    "exists, can't clobber",
			mapping: &Mapping{"/whatever", "/exists/cant/clobber", 0640, 0750, 500, 500, false},
			wantErr: errCantClobber,
		},
		{
			name:    "doesn't exist, should copy",
			mapping: &Mapping{"/whatever", "/doesnt/exist", 0640, 0750, 500, 500, false},
			want:    true,
		},
		// No such thing as "doesn't exist, shouldn't copy"
	} {
		got, gotErr := tt.mapping.shouldCopy(srcCksm)
		if gotErr != tt.wantErr {
			t.Fatalf("%v: wanted error: %v\ngot error: %v", tt.name, tt.wantErr, gotErr)
		}
		if got != tt.want {
			t.Errorf("%v: wanted %v, got %v", tt.name, tt.want, got)
		}
	}
}
