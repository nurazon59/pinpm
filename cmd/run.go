package cmd

import (
	"github.com/adrg/xdg"
	"github.com/alecthomas/kong"
	template "github.com/nurazon59/pinpm"
)

const appVersion = "v0.1.0"

var CLI struct {
	Config  string           `help:"Path to config file." env:"PINPM"`
	Version kong.VersionFlag `name:"version" help:"Print version information and quit."`
}

func Run() error {
	kong.Parse(&CLI, kong.Name("go-pinpm"), kong.Vars{
		"version": appVersion,
	})

	configPath := CLI.Config
	if configPath == "" {
		var err error
		configPath, err = xdg.ConfigFile("pinpm/config.yaml")
		if err != nil {
			return err
		}
	}

	_, err := pinpm.Load(configPath)
	if err != nil {
		return err
	}

	return nil
}
