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
	"os"

	"github.com/spf13/afero"
)

// Fs represents a file system with which PrepPi will interact. It extends
// afero.Fs with a few functions we need but which are not supplied there.
type Fs interface {
	afero.Fs
	// Chown changes the numeric uid and gid of the named file.
	Chown(name string, uid, gid int) error
}

// BasePathFs extends afero.BasePathFs with some things needed by PrepPi.
type BasePathFs struct {
	*afero.BasePathFs

	source Fs
	path   string
}

// Chown changes the numeric uid and gid of the named file.
func (b *BasePathFs) Chown(name string, uid, gid int) error {
	if name, err := b.RealPath(name); err != nil {
		return &os.PathError{Op: "chown", Path: name, Err: err}
	}
	return b.source.Chown(name, uid, gid)
}

// NewBasePathFs creates and return a BasePathFs instance.
func NewBasePathFs(source Fs, path string) Fs {
	return &BasePathFs{afero.NewBasePathFs(source, path).(*afero.BasePathFs), source, path}
}

// OsFs extends afero.OsFs with some things needed by PrepPi.
type OsFs struct {
	*afero.OsFs
}

// Chown changes the numeric uid and gid of the named file.
func (o *OsFs) Chown(name string, uid, gid int) error {
	return os.Chown(name, uid, gid)
}

// NewOsFs creates and return a OsFs instance.
func NewOsFs() Fs {
	return &OsFs{afero.NewOsFs().(*afero.OsFs)}
}

// MemMapFs extends afero.MemMapFs with some things needed by PrepPi.
type MemMapFs struct {
	*afero.MemMapFs
}

// Chown for MemMapFs does nothing. The underlying afero.MemMapFs doesn't
// support UID/GID at this time, and so there's no point. It would be good for
// testing in the future.
// TODO(cfunkhouser): Implement this for better testing.
func (m *MemMapFs) Chown(name string, uid, gid int) error {
	return nil
}

// NewMemMapFs creates and return a MemMapFs instance.
func NewMemMapFs() Fs {
	return &MemMapFs{afero.NewMemMapFs().(*afero.MemMapFs)}
}
