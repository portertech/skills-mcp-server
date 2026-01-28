// Package main implements the skills MCP server CLI.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/portertech/skills-mcp-server/internal/registry"
	"github.com/portertech/skills-mcp-server/internal/server"
)

var (
	version = "dev"
)

func main() {
	var (
		listSkills  bool
		verbose     bool
		showVersion bool
	)

	flag.BoolVar(&listSkills, "list", false, "List discovered skills and exit")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&showVersion, "version", false, "Print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [skills_root]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "An MCP server that exposes Claude-compatible skills as tools.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nDefault skills root: ~/.skills\n")
	}
	flag.Parse()

	if showVersion {
		fmt.Printf("skills %s\n", version)
		os.Exit(0)
	}

	logLevel := slog.LevelError
	if verbose {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	skillsRoot := flag.Arg(0)
	if skillsRoot == "" {
		skillsRoot = defaultSkillsRoot()
	}

	skillsRoot, err := expandPath(skillsRoot)
	if err != nil {
		logger.Error("failed to expand skills root path", "error", err)
		os.Exit(1)
	}

	if _, err := os.Stat(skillsRoot); os.IsNotExist(err) {
		logger.Error("skills root directory does not exist", "path", skillsRoot)
		os.Exit(1)
	}

	reg := registry.NewRegistry(skillsRoot, logger)
	if err := reg.Scan(); err != nil {
		logger.Error("failed to scan skills", "error", err)
		os.Exit(1)
	}

	if listSkills {
		skills := reg.List()
		if len(skills) == 0 {
			fmt.Println("No skills found.")
			os.Exit(0)
		}
		fmt.Printf("Found %d skill(s) in %s:\n\n", len(skills), skillsRoot)
		for _, s := range skills {
			fmt.Printf("  %s\n", s.Name)
			fmt.Printf("    %s\n", s.Description)
			fmt.Printf("    Path: %s\n\n", s.Path)
		}
		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Info("received shutdown signal")
		cancel()
	}()

	srv := server.New(reg, logger)
	if err := srv.Run(ctx); err != nil && ctx.Err() == nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func defaultSkillsRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".skills"
	}
	return filepath.Join(home, ".skills")
}

func expandPath(path string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}
	return filepath.Abs(path)
}
