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
	"path"
	"strings"
	"time"

	"github.com/cfunkhouser/preppi/preppi"
	"github.com/google/subcommands"
)

var (
	prepConfigDefault     = "/boot/preppi/preppi.conf"
	bakeRecipeRootDefault = "/etc/preppi/recipes"
	bakeRecipeNameDefault = "recipe.json"
)

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
	f.StringVar(&c.config, "config", prepConfigDefault, "override the default config file path.")

	f.StringVar(&preppi.RebootCommand, "reboot_command", preppi.RebootCommand,
		"Command to run to reboot the system. No arguments may be passed.")
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
	if !c.dryRun {
		n, err := mapper.Apply()
		if err != nil {
			log.Printf("Error: %v", err)
		}
		log.Printf("preppi processed %v files, modified %v in %v", len(mapper.Mappings), n, time.Since(start))
		if n > 0 && err == nil && c.reboot {
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
	f.StringVar(&c.recipeRoot, "root", bakeRecipeRootDefault, "override default recipe root location.")
	f.StringVar(&c.destination, "out", "", "path under which generated files are written")
}

func unpackKV(kvs []string) (map[string]string, error) {
	v := make(map[string]string)
	for _, kv := range kvs {
		s := strings.Split(kv, "=")
		if len(s) != 2 {
			return nil, fmt.Errorf("Invalid Value Specification: %q", kv)
		}
		v[s[0]] = s[1]
	}
	return v, nil
}

func (c *bakeCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if c.recipe == "" {
		log.Print("No -recipe provided, nothing to do!")
		return subcommands.ExitFailure
	}
	if c.destination == "" {
		log.Print("No -out provided, refusing to write to current directory without explicit instruction")
		return subcommands.ExitFailure
	}
	vars, err := unpackKV(f.Args())
	if err != nil {
		log.Printf("error processing variables: %v", err)
		return subcommands.ExitFailure
	}
	rd := &preppi.RecipeData{}
	rd.Vars = vars

	recipePath := path.Join(c.recipeRoot, c.recipe, bakeRecipeNameDefault)
	recipe, err := preppi.RecipeFromFile(recipePath)
	if err != nil {
		log.Printf("error reading recipe %q: %v", recipePath, err)
		return subcommands.ExitFailure
	}

	start := time.Now()
	log.Printf("baking recipe %q", c.recipe)
	if err := recipe.Bake(c.destination, rd); err != nil {
		log.Printf("error baking recipe: %v", err)
	}
	log.Printf("preppi baked recipe %q in %v", c.recipe, time.Since(start))
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
