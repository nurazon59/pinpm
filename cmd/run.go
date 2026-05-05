package cmd

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	pinpm "github.com/nurazon59/pinpm"
)

const appVersion = "v0.1.0"

var CLI struct {
	Config  string           `help:"Path to config file." env:"PINPM"`
	Version kong.VersionFlag `help:"Print version information and quit."`
	Init    InitCmd          `cmd:"" help:"Generate .pinpm.yaml config file."`
	Pin     PinCmd           `cmd:"" help:"Pin dependencies in package.json."`
	Check   CheckCmd         `cmd:"" help:"Check for unpinned dependencies."`
}

type InitCmd struct {
	Output string `short:"o" default:".pinpm.yaml" help:"Output file path."`
}

func (i *InitCmd) Run() error {
	cfg, _ := pinpm.LoadConfig("")
	return pinpm.SaveConfig(i.Output, cfg)
}

type PinCmd struct {
	File   string `short:"f" default:"package.json" help:"Target package.json."`
	DryRun bool   `short:"n" help:"Show changes without writing."`
}

func (p *PinCmd) Run() error {
	pkg, err := pinpm.LoadPackageJSON(p.File)
	if err != nil {
		return err
	}

	client := pinpm.NewClient("")
	checker := pinpm.NewChecker(client)

	result, err := checker.Check(pkg)
	if err != nil {
		return err
	}

	if len(result.Pinned) == 0 {
		fmt.Println("All dependencies are already pinned.")
		return nil
	}

	if p.DryRun {
		for _, change := range result.Pinned {
			fmt.Printf("%s (%s): %s -> %s\n", change.Name, change.Section, change.OldVersion, change.NewVersion)
		}
		return nil
	}

	if err := checker.Apply(p.File, pkg, result); err != nil {
		return err
	}

	fmt.Printf("Pinned %d dependencies.\n", len(result.Pinned))
	return nil
}

type CheckCmd struct {
	File   string `short:"f" default:"package.json" help:"Target package.json."`
	Format string `short:"o" default:"text" enum:"text,json" help:"Output format."`
}

func (c *CheckCmd) Run() error {
	pkg, err := pinpm.LoadPackageJSON(c.File)
	if err != nil {
		return err
	}

	client := pinpm.NewClient("")
	checker := pinpm.NewChecker(client)

	result, err := checker.Check(pkg)
	if err != nil {
		return err
	}

	if len(result.Pinned) == 0 {
		fmt.Println("All dependencies are already pinned.")
		return nil
	}

	if c.Format == "json" {
		data, err := pinpm.MarshalResultJSON(result)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	for _, change := range result.Pinned {
		fmt.Printf("%s (%s): %s -> %s\n", change.Name, change.Section, change.OldVersion, change.NewVersion)
	}

	os.Exit(1)
	return nil
}

func Run() error {
	ctx := kong.Parse(&CLI, kong.Name("pinpm"), kong.Vars{
		"version": appVersion,
	})

	return ctx.Run()
}
