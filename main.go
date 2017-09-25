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
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cfunkhouser/preppi/preppi"
	"github.com/google/subcommands"
)

var (
	confFile = flag.String("config", "/boot/preppi/preppi.conf", "Mappings file path")
	verFlag  = flag.Bool("version", false, "If true, print version and exit.")
	dryRun   = flag.Bool("dry_run", false, "If true, parses the config but changes nothing.")
	reboot   = flag.Bool("reboot", false, "Reboot when file changes have been written")

	prepConfigDefault = "/boot/preppi/preppi.conf"
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

type versionCmd struct{}

func (*versionCmd) Name() string     { return "version" }
func (*versionCmd) Synopsis() string { return "print the version" }
func (*versionCmd) Usage() string {
	return "Usage:\tpreppi version\n"
}

func (*versionCmd) SetFlags(_ *flag.FlagSet) {}

func (*versionCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	fmt.Println(preppi.VersionString())
	return subcommands.ExitSuccess
}

type prepCmd struct {
	reboot bool
	dryRun bool
	config string
}

func (*prepCmd) Name() string     { return "prepare" }
func (*prepCmd) Synopsis() string { return "prepare the system" }
func (*prepCmd) Usage() string {
	return "Usage:\tpreppi prepare [-config <path>] [-dry_run] [-reboot]\n"
}

func (c *prepCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&c.reboot, "reboot", false, "reboot the system after preparation.")
	f.BoolVar(&c.dryRun, "dry_run", false, "parse the config and simulate, but make no changes.")
	f.StringVar(&c.config, "config", prepConfigDefault, fmt.Sprintf("override the default config file path. default: %q", prepConfigDefault))
}

func (c *prepCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	log.Printf("%v starting", preppi.VersionString())

	if c.config == "" {
		log.Print("No -config specified, nothing to do!")
		return subcommands.ExitFailure
	}

	if ok, err := checkConfigExists(c.config); !ok {
		if err != nil {
			log.Printf("couldn't stat -config %q: %v", c.config, err)
			return subcommands.ExitFailure
		}
		log.Printf("specified -config %q doesn't exist, nothing to do!", c.config)
		return subcommands.ExitSuccess
	}

	mapper, err := preppi.MapperFromConfig(c.config)
	if err != nil {
		log.Printf("error processing -config %q: %v", c.config, err)
		return subcommands.ExitFailure
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
	return subcommands.ExitSuccess
}

type bakeCmd struct {
	recipe      string
	recipeRoot  string
	destination string
}

func (*bakeCmd) Name() string     { return "bake" }
func (*bakeCmd) Synopsis() string { return "bake a recipe" }
func (*bakeCmd) Usage() string {
	return "Usage:\tpreppi bake [-root <path>] -recipe <name> -out <path> [var1=val1 [var2=val2] ...]\n"
}

func (c *bakeCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.recipe, "recipe", "", "name of the recipe to bake. required.")
	f.StringVar(&c.recipeRoot, "root", "", "override default recipe root location.")
	f.StringVar(&c.destination, "out", "", "path under which generated files are written")
}

func (c *bakeCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	log.Println("Not yet implemented")
	return subcommands.ExitSuccess
}

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&versionCmd{}, "")
	subcommands.Register(&prepCmd{}, "")
	subcommands.Register(&bakeCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
