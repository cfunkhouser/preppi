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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/spf13/afero"
)

// clobberFlag is the set of flags passed to FileOpen when Apply()ing a Mapping.
const clobberFlag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC

var (
	// preppiFS is a Fs. It is a var for testing.
	preppiFS Fs

	errCantClobber = errors.New("Can't clobber file")
)

func init() {
	preppiFS = NewOsFs()
}

// Mapping represents a file mapped from the Source to Destination. Mode, UID
// and GID apply to the written Destination file. DirMode is applied to any
// directories created.
type Mapping struct {
	Source      string      `json:"source"`
	Destination string      `json:"destination"`
	Mode        os.FileMode `json:"mode"`
	DirMode     os.FileMode `json:"dirmode"`
	UID         int         `json:"uid"`
	GID         int         `json:"gid"`

	// Clobber is true when it's okay to overrwite Destination if it exists.
	Clobber bool `json:"clobber,omitempty"`
}

// destinationExists checks if the file exists. If anything unexpected happens,
// return the error we encountered.
func (m *Mapping) destinationExists() (bool, error) {
	_, err := preppiFS.Stat(m.Destination)
	if err != nil {
		if !os.IsNotExist(err) {
			// There was some unexpected error stat-ing the destination
			return false, err
		}
		return false, nil
	}
	return true, nil
}

// destinationFingerprint opens the destination file for read and checksums
// the file. If the destination doesn't exist, it is an error.
func (m *Mapping) destinationFingerprint() ([]byte, error) {
	dst, err := preppiFS.OpenFile(m.Destination, os.O_RDONLY, 0000)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	s, err := dst.Stat()
	if err != nil {
		return nil, err
	}

	dstCksm, err := Fingerprint(s.Mode(), dst)
	if err != nil {
		return nil, err
	}
	return dstCksm, nil
}

// prepareNewDestination opens the destination for write with the appropriate
// mode, creating any non-extant parent directories.
func (m *Mapping) prepareNewDestination() (afero.File, error) {
	// Make sure all destination parent directories exist
	if err := preppiFS.MkdirAll(path.Dir(m.Destination), m.DirMode); err != nil {
		return nil, err
	}
	dst, err := preppiFS.OpenFile(m.Destination, clobberFlag, m.Mode)
	if err != nil {
		return nil, err
	}
	return dst, nil
}

// shouldCopy determines if the source should be applied to the destination.
func (m *Mapping) shouldCopy(srcCksm []byte) (bool, error) {
	exists, err := m.destinationExists()
	if err != nil {
		return false, err
	}
	// If the file doesn't exist, we're safe to copy.
	if !exists {
		return true, nil
	}
	// Get a fingerprint for the existing destination.
	dstCksm, err := m.destinationFingerprint()
	if err != nil {
		return false, err
	}
	log.Printf("Destination %x", srcCksm)
	log.Printf("     Source %x", dstCksm)
	// If the fingerprints match, nothing to do.
	if bytes.Compare(srcCksm, dstCksm) == 0 {
		return false, nil
	}
	// If we're allowed to clobber the file, say so.
	if m.Clobber {
		return true, nil
	}
	return false, errCantClobber
}

// source opens and checksums the file. The caller is responsible for closing.
// Return values are undefined if the returned error is not nil.
func (m *Mapping) source() (afero.File, []byte, error) {
	src, err := preppiFS.Open(m.Source)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't open source: %v", err)
	}
	cksm, err := Fingerprint(m.Mode, src)
	if err != nil {
		return nil, nil, err
	}
	return src, cksm, nil
}

// Apply the mapping, copying Source to Destination and set the metadata.
func (m *Mapping) Apply() (bool, error) {
	src, srcCksm, err := m.source()
	if err != nil {
		return false, err
	}
	defer src.Close()

	ok, err := m.shouldCopy(srcCksm)
	if err != nil {
		return false, err
	}
	if !ok {
		log.Printf("skipping %q", m.Destination)
		return false, nil
	}

	log.Printf("beginning copy %q -> %q", m.Source, m.Destination)
	dst, err := m.prepareNewDestination()
	if err != nil {
		return false, err
	}
	if _, err := io.Copy(dst, src); err != nil {
		return false, err
	}
	if err := preppiFS.Chown(m.Destination, m.UID, m.GID); err != nil {
		return false, err
	}
	// No errors, and the destination has changed.
	return true, nil
}

// Mapper represents a set of file mappings.
type Mapper struct {
	Mappings []*Mapping `json:"map"`
}

// Apply the set of mappings to the preppiFS. Returns a count of files modified,
// and the first error encountered. If an error is encoutered, modified count
// reflects number of files modified beforehand.
func (m *Mapper) Apply() (int, error) {
	modified := 0
	for _, mapping := range m.Mappings {
		ok, err := mapping.Apply()
		if err != nil {
			return modified, err
		}
		if ok {
			modified++
		}
	}
	return modified, nil
}

// MapperFromConfig reads a config and returns a Mapper
func MapperFromConfig(config string) (*Mapper, error) {
	data, err := afero.ReadFile(preppiFS, config)
	if err != nil {
		return nil, fmt.Errorf("failed reading config %q: %v", config, err)
	}
	m := &Mapper{}
	if err := json.Unmarshal(data, m); err != nil {
		return nil, fmt.Errorf("failed reading config %q: %v", config, err)
	}
	return m, nil
}

// MapperToFile marshals m to path as JSON.
func MapperToFile(path string, m *Mapper) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	f, err := preppiFS.Create(path)
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		return err
	}
	return nil
}
