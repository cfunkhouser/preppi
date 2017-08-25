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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cfunkhouser/preppi/preppi"
)

var (
	version = "0.1.0"
	buildID = "dev" // Overriden at build time by build scripts.

	mapFile = flag.String("config", "", "Mappings file path")
	verFlag = flag.Bool("version", false, "If true, print version and exit.")
)

func main() {
	flag.Parse()

	if *verFlag {
		fmt.Printf("preppi v%v (%v)\n", version, buildID)
		os.Exit(0)
	}

	log.Printf("preppi v%v (%v) starting", version, buildID)
	start := time.Now()
	if *mapFile == "" {
		log.Fatal("No --config specified, nothing to do!")
	}
	mapper, err := preppi.MapperFromConfig(*mapFile)
	if err != nil {
		log.Fatal(err)
	}
	if err := mapper.Apply(); err != nil {
		log.Fatal(err)
	}
	log.Printf("preppi applied %v files in %v", len(mapper.Mappings), time.Since(start))
}
