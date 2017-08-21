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
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"
)

// clobberFlag is the set of flags passed to FileOpen when Apply()ing a Mapping.
const clobberFlag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC

// preppiFS is an afero.Fs. It is a var for testing.
var preppiFS afero.Fs

func init() {
	preppiFS = afero.NewOsFs()
}

// Mapping represents a file mapped from the Source to Destination. Mode, UID and GID
// apply to the written Destination file.
type Mapping struct {
	Source      string      `json:"source"`
	Destination string      `json:"destination"`
	Mode        os.FileMode `json:"mode"`
	UID         int         `json:"uid"`
	GID         int         `json:"gid"`

	// Clobber is true when it's okay to overrwite Destination if it exists.
	Clobber bool `json:"clobber,omitempty"`
}

// Apply the mapping, copying Source to Destination and set the Mode, UID and GID.
func (m *Mapping) Apply() error {
	if _, err := preppiFS.Stat(m.Source); err != nil {
		return fmt.Errorf("couldn't stat source: %v", err)
	}
	// Make sure that if Destination exists, we are permitted to Clobber it.
	if _, err := preppiFS.Stat(m.Destination); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("something unexpected happened stat-ing destination: %v", err)
		}
		// NotExist errors are okay; that's what we're hoping for.
	} else if !m.Clobber {
		return fmt.Errorf("can't clobber existing destination file %q", m.Destination)
	}

	// Open Destination first, since it's more likely to fail.
	dst, err := preppiFS.OpenFile(m.Destination, clobberFlag, m.Mode)
	if err != nil {
		return fmt.Errorf("couldn't open destination the way we wanted: %v", err)
	}

	src, err := preppiFS.Open(m.Source)
	if err != nil {
		return fmt.Errorf("couldn't open source: %v", err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copy failed: %v", m)
	}

	dst.Close()
	src.Close()

	if err := preppiFS.Chown(m.Destination, m.UID, m.GID); err != nil {
		return fmt.Errorf("couldn't chown destination: %v", err)
	}

	return nil
}

// Mapper represents a set of file mappings.
type Mapper struct {
	Mappings []*Mapping
}

// Apply the set of mappings to the preppiFS
func (m *Mapper) Apply() error {
	for _, mapping := range m.Mappings {
		if err := mapping.Apply(); err != nil {
			return err
		}
	}
	return nil
}

// MapperFromConfig reads a config and returns a Mapper
func MapperFromConfig(config string) (*Mapper, error) {
	data, err := afero.ReadFile(preppiFS, config)
	if err != nil {
		return nil, fmt.Errorf("failed reading config %q: %v", config, err)
	}
	var m []*Mapping
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed reading config %q: %v", config, err)
	}
	return &Mapper{Mappings: m}, nil
}
