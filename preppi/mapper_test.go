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

func TestSourceOK(t *testing.T) {
	origPreppiFS := preppiFS
	defer func() { preppiFS = origPreppiFS }()

	preppiFS = NewMemMapFs()

	files := map[string]*testFile{
		"/boot/stuff/something": &testFile{
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
		wantErr bool
	}{
		{
			name:    "source ok",
			mapping: &Mapping{"/boot/stuff/something", "/doesnt/matter", 0640, 0750, 500, 500, false},
			wantErr: false,
		},
		{
			name:    "source not ok",
			mapping: &Mapping{"/boot/stuff/nothing", "/doesnt/matter", 0640, 0750, 500, 500, false},
			wantErr: true,
		},
	} {
		err := tt.mapping.Apply()
		if err == nil && tt.wantErr {
			t.Errorf("%s: wanted error, got none", tt.name)
		} else if err != nil && !tt.wantErr {
			t.Errorf("%s: wanted no error, got: %q", tt.name, err)
		}
	}
}

func TestDestinationOK(t *testing.T) {
	origPreppiFS := preppiFS
	defer func() { preppiFS = origPreppiFS }()
	preppiFS = NewMemMapFs()

	files := map[string]*testFile{
		"/this/exists": &testFile{
			[]byte("This is something"),
			0644,
			0755,
			500,
			500,
		},
		"/boot/stuff/something": &testFile{
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
		wantErr bool
	}{
		{
			name:    "destination doesn't exist",
			mapping: &Mapping{"/boot/stuff/something", "/this/doesnt/exist", 0640, 0750, 500, 500, false},
			wantErr: false,
		},
		{
			name:    "destination exists, can't clobber",
			mapping: &Mapping{"/boot/stuff/something", "/this/exists", 0640, 0750, 500, 500, false},
			wantErr: true,
		},
		{
			name:    "destination exists, can clobber",
			mapping: &Mapping{"/boot/stuff/something", "/this/exists", 0640, 0750, 500, 500, true},
			wantErr: false,
		},
	} {
		err := tt.mapping.Apply()
		if err == nil && tt.wantErr {
			t.Errorf("%s: wanted error, got none", tt.name)
		} else if err != nil && !tt.wantErr {
			t.Errorf("%s: wanted no error, got: %q", tt.name, err)
		}
	}
}

func TestCopyToDestination(t *testing.T) {
	origPreppiFS := preppiFS
	defer func() { preppiFS = origPreppiFS }()

	files := map[string]*testFile{
		"/boot/stuff/something": &testFile{
			[]byte("This is something"),
			0644,
			0755,
			500,
			500,
		},
	}

	for _, tt := range []struct {
		name    string
		mapping *Mapping
	}{
		{
			name:    "same permissions, same mode",
			mapping: &Mapping{"/boot/stuff/something", "/some/place", 0644, 0755, 500, 500, false},
		},
		{
			name:    "same permissions, different mode",
			mapping: &Mapping{"/boot/stuff/something", "/some/place", 0512, 0623, 500, 500, false},
		},
		{
			name:    "different permissions, same mode",
			mapping: &Mapping{"/boot/stuff/something", "/some/place", 0644, 0755, 0, 0, false},
		},
		{
			name:    "different permissions, different mode",
			mapping: &Mapping{"/boot/stuff/something", "/some/place", 0444, 0555, 0, 0, false},
		},
	} {
		// Start with a cleanly-configured FS for each case.
		preppiFS = NewMemMapFs()
		setUpFilesystemForTest(t, preppiFS, files)

		if err := tt.mapping.copyToDestination(); err != nil {
			t.Errorf("%s: wanted no error, got: %v", tt.name, err)
		}
		s, err := preppiFS.Stat(tt.mapping.Destination)
		if err != nil {
			t.Fatalf("%s: stat failed: %v", tt.name, err)
		}
		if s.Mode() != tt.mapping.Mode {
			t.Errorf("%s: wanted mode %v, got mode %v", tt.name, tt.mapping.Mode, s.Mode())
		}
		// TODO(cfunkhouser): Figure out how to check UID/GID here.
	}
}
