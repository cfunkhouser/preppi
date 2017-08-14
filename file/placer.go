// Copyright 2017 Google Inc.
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

package file

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Mapper represents a set of file mappings.
type Mapper struct {
	Mappings []*Mapping
}

// Apply the set of mappings to the filesystem
func (m *Mapper) Apply() error {
	for _, mapping := range m.Mappings {
		if err := mapping.Apply(); err != nil {
			return err
		}
	}
	return nil
}

// MapperFromConfig loads a config JSON file, returns a Mapper
func MapperFromConfig(config string) (*Mapper, error) {
	data, err := ioutil.ReadFile(config)
	if err != nil {
		return nil, fmt.Errorf("failed reading config %q: %v", config, err)
	}
	p := &Mapper{Mappings: make([]*Mapping, 0)}
	if err := json.Unmarshal(data, &p.Mappings); err != nil {
		return nil, fmt.Errorf("failed reading config %q: %v", config, err)
	}
	return p, nil
}
