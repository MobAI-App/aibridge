package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/MobAI-App/aibridge/internal/bridge"
	"github.com/MobAI-App/aibridge/internal/config"
	"github.com/MobAI-App/aibridge/internal/patterns"
	"github.com/MobAI-App/aibridge/internal/server"
	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"

	flagPort        int
	flagHost        string
	flagBusyPattern string
	flagTimeout     int
	flagVerbose     bool
	flagVersion     bool
	flagParanoid    bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "aibridge [flags] -- <command> [args...]",
		Short: "AiBridge wraps AI CLI tools with HTTP API for text injection",
		Long: `AiBridge is a CLI tool that wraps AI coding assistants (Claude Code, Codex, Gemini CLI)
with a PTY and exposes an HTTP API for external text injection.`,
		DisableFlagParsing: false,
		Args:               cobra.MinimumNArgs(1),
		Run:                run,
	}

	rootCmd.Flags().IntVarP(&flagPort, "port", "p", config.DefaultPort, "HTTP server port")
	rootCmd.Flags().StringVar(&flagHost, "host", config.DefaultHost, "HTTP server host")
	rootCmd.Flags().StringVar(&flagBusyPattern, "busy-pattern", "", "Custom busy detection regex pattern")
	rootCmd.Flags().IntVarP(&flagTimeout, "timeout", "t", config.DefaultTimeout, "Sync injection timeout in seconds")
	rootCmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.Flags().BoolVar(&flagParanoid, "paranoid", false, "Inject text without hitting Enter")
	rootCmd.Flags().BoolVar(&flagVersion, "version", false, "Print version and exit")

	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if flagVersion {
			fmt.Printf("aibridge version %s\n", version)
			os.Exit(0)
		}
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Fatal("No command specified")
	}

	command := args[0]
	commandArgs := args[1:]

	toolName := filepath.Base(command)

	var pattern *patterns.Pattern
	if flagBusyPattern != "" {
		pattern = &patterns.Pattern{Regex: flagBusyPattern}
	} else {
		pattern = patterns.GetPattern(toolName)
		if pattern == nil {
			pattern = patterns.DefaultPattern()
		}
	}

	if flagVerbose {
		log.Printf("Starting aibridge with command: %s %v", command, commandArgs)
		log.Printf("Using pattern: %s", pattern.Regex)
	}

	b, err := bridge.New(command, commandArgs, pattern.Regex, flagVerbose, flagParanoid)
	if err != nil {
		log.Fatalf("Failed to create bridge: %v", err)
	}

	srv := server.New(b, flagHost, flagPort, flagVerbose)
	go func() {
		if err := srv.Start(); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		srv.GracefulShutdown()
		_ = b.Close()
		os.Exit(0)
	}()

	if err := b.Start(); err != nil {
		log.Fatalf("Failed to start bridge: %v", err)
	}

	if err := b.Wait(); err != nil {
		if flagVerbose {
			log.Printf("Child process exited: %v", err)
		}
	}

	_ = b.Close()
	srv.GracefulShutdown()
}
