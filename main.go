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

	confFile = flag.String("config", "/boot/preppi/preppi.conf", "Mappings file path")
	verFlag  = flag.Bool("version", false, "If true, print version and exit.")
	dryRun   = flag.Bool("dry_run", false, "If true, parses the config but changes nothing.")
	reboot   = flag.Bool("reboot", false, "Reboot when file changes have been written")
)

func init() {
	flag.StringVar(&preppi.RebootCommand, "reboot_command", preppi.RebootCommand,
		"Command to run to reboot the system. No arguments may be passed.")
}

func checkConfigExists(p string) (bool, error) {
	_, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func main() {
	flag.Parse()

	if *verFlag {
		fmt.Printf("preppi v%v (%v)\n", version, buildID)
		return
	}

	log.Printf("preppi v%v (%v) starting", version, buildID)
	if *confFile == "" {
		log.Print("No --config specified, nothing to do!")
		return
	}
	if ok, err := checkConfigExists(*confFile); !ok {
		if err != nil {
			log.Fatalf("Couldn't stat --config %q: %v", *confFile, err)
		}
		log.Printf("Specified --config %q doesn't exist, nothing to do!", *confFile)
		return
	}
	mapper, err := preppi.MapperFromConfig(*confFile)
	if err != nil {
		log.Fatalf("Error processing --config %q: %v", *confFile, err)
	}
	start := time.Now()
	if !*dryRun {
		n, err := mapper.Apply()
		if err != nil {
			log.Printf("Error: %v", err)
		}
		log.Printf("preppi processed %v files, modified %v in %v", len(mapper.Mappings), n, time.Since(start))
		if n > 0 && err == nil && *reboot {
			log.Printf("Files changed, rebooting with: %q", preppi.RebootCommand)
			if err := preppi.RebootSystem(); err != nil {
				log.Printf("preppi tried to reboot the system but failed: %v", err)
			}
		}
	}
}
