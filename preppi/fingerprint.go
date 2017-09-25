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
	"crypto/sha256"
	"encoding/binary"
	"io"
	"os"
)

// fingerprint checksums the file and its relavent metadata for PrepPi.
// TODO(cfunkhouser): UID/GID
func Fingerprint(mode os.FileMode, f io.ReadSeeker) ([]byte, error) {
	h := sha256.New()
	b := make([]byte, 4)

	// Hash the file mode
	binary.LittleEndian.PutUint32(b, uint32(mode))
	h.Write(b)

	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
