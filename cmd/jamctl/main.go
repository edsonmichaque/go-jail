package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/edsonmichaque/jam/internal/jam"
)

func main() {
	if len(os.Args) < 2 {
		panic("insufficient args")
	}

	root := "tmp/etc/jail.conf.d"

	var (
		jamRoot   = "tmp/var/jam"
		jailsPath = filepath.Join(jamRoot, "jails")
	)

	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(root, 0o755); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	cfg := filepath.Join(jamRoot, "jail.json")

	if _, err := os.Stat(cfg); err != nil {
		if os.IsNotExist(err) {
			f, err := os.Create(cfg)
			if err != nil {
				panic(err)
			}

			f.Close()

		} else {
			panic(err)
		}
	}

	b, err := os.ReadFile(cfg)
	if err != nil {
		panic(err)
	}

	var config Config
	if len(b) != 0 {
		if err := json.Unmarshal(b, &config); err != nil {
			panic(err)
		}
	}

	if len(config.Jails) == 0 {
		config.Jails = make([]string, 0)
	}

	err = jam.Create(context.Background(), root, &jam.Opts{
		Name: os.Args[1],
		Host: &jam.HostOpts{
			Hostname: "localhost",
		},
		IPv4: &jam.IPv4Opts{
			IPOpts: jam.IPOpts{
				Addr: []string{"127.0.0.1", "127.0.0.2"},
			},
		},
		Path: filepath.Join(jailsPath, os.Args[1]),
		Exec: &jam.ExecOpts{
			PreStart: `echo "pre-start"`,
			Start:    `echo "start"`,
			Clean:    true,
		},
		Interface: "em0",
		Persist:   true,
	})
	if err != nil {
		panic(err)
	}

	config.Jails = append(config.Jails, os.Args[1])
	b, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(cfg, b, 0o644); err != nil {
		panic(err)
	}
}

type Config struct {
	Jails []string
}
