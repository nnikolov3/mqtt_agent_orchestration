package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/nnikolov3/mqtt-agent-orchestration/internal/config"
	"github.com/nnikolov3/mqtt-agent-orchestration/internal/mqtt"
	"github.com/nnikolov3/mqtt-agent-orchestration/internal/rag"
)

// High-verbosity, explicit, defensive CLI that replaces Bash scripts with parallel Go operations

type buildTarget struct {
	Name string
	Path string
}

func main() {
	if len(os.Args) < 2 {
		printOpsUsage()
		os.Exit(1)
	}

	sub := os.Args[1]
	args := os.Args[2:]

	switch sub {
	case "build":
		cmdBuild(args)
	case "run":
		cmdRun(args)
	case "lint":
		cmdLint(args)
	case "cleanup":
		cmdCleanup(args)
	case "mcp":
		cmdMCP(args)
	case "train":
		cmdTrain(args)
	case "help", "-h", "--help":
		printOpsUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", sub)
		printOpsUsage()
		os.Exit(1)
	}
}

func printOpsUsage() {
	fmt.Println("Usage: ops <subcommand> [options]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  build     Build binaries, run tests, validate configs (parallel)")
	fmt.Println("  run       Start orchestrator and workers with health checks")
	fmt.Println("  lint      Run formatting, vet, and optional linters")
	fmt.Println("  cleanup   Clean logs and caches with optional dry-run")
	fmt.Println("  mcp       Helpers for Qdrant MCP (compose/system checks)")
	fmt.Println("  train     Wrapper for LoRA training pipeline")
}

// -----------------
// build
// -----------------

func cmdBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	clean := fs.Bool("clean", false, "Clean build artifacts and Go caches before building")
	verbose := fs.Bool("verbose", false, "Verbose logging")
	skipTests := fs.Bool("skip-tests", false, "Skip 'go test' step")
	_ = fs.Parse(args)

	logInfo("Starting parallel build")
	projectRoot, err := os.Getwd()
	if err != nil {
		logError("failed to get working directory: %v", err)
		os.Exit(1)
	}

	binDir := filepath.Join(projectRoot, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		logError("failed to create bin dir: %v", err)
		os.Exit(1)
	}

	if *clean {
		logInfo("Cleaning build outputs and Go caches")
		_ = os.RemoveAll(binDir)
		_ = os.MkdirAll(binDir, 0o755)
		if err := runLogged(exec.Command("go", "clean", "-cache", "-modcache", "-testcache"), *verbose); err != nil {
			logError("go clean failed: %v", err)
			os.Exit(1)
		}
	}

	if !*skipTests {
		logInfo("Running tests (race, all packages)")
		if err := runLogged(exec.Command("go", "test", "./...", "-race"), *verbose); err != nil {
			logError("tests failed: %v", err)
			os.Exit(1)
		}
	}

	// Validate model config using internal loader (no yq dependency)
	validateModels(filepath.Join(projectRoot, "configs", "models.yaml"))

	// Build targets concurrently
	targets := []buildTarget{
		{Name: "worker", Path: "./cmd/worker"},
		{Name: "server", Path: "./cmd/server"},
		{Name: "orchestrator", Path: "./cmd/orchestrator"},
		{Name: "role-worker", Path: "./cmd/role-worker"},
		{Name: "client", Path: "./cmd/client"},
		{Name: "rag-service", Path: "./cmd/rag-service"},
		{Name: "embedding-worker", Path: "./cmd/embedding-worker"},
		{Name: "ops", Path: "./cmd/ops"},
	}

	var wg sync.WaitGroup
	maxWorkers := runtime.GOMAXPROCS(0)
	sem := make(chan struct{}, maxWorkers)
	bErr := make(chan error, len(targets))

	for _, t := range targets {
		wg.Add(1)
		sem <- struct{}{}
		go func(target buildTarget) {
			defer wg.Done()
			defer func() { <-sem }()

			out := filepath.Join(binDir, target.Name)
			cmd := exec.Command("go", "build", "-o", out, target.Path)
			if *verbose {
				logInfo("Building %s from %s", target.Name, target.Path)
			}
			if err := runLogged(cmd, *verbose); err != nil {
				bErr <- fmt.Errorf("build %s failed: %w", target.Name, err)
				return
			}
			logInfo("Built %s", target.Name)
		}(t)
	}

	wg.Wait()
	close(bErr)
	for err := range bErr {
		if err != nil {
			logError(err.Error())
			os.Exit(1)
		}
	}

	logSuccess("Build completed. Binaries in ./bin")
}

func validateModels(configPath string) {
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		logWarn("models.yaml not found at %s (local models will be limited)", configPath)
		return
	}
	cfg, err := config.LoadModelConfig(configPath)
	if err != nil {
		logWarn("model configuration validation warning: %v", err)
		return
	}
	_ = cfg // parsed to ensure validity; further checks can be added
	logInfo("Model configuration validated: %s", filepath.Base(configPath))
}

// -----------------
// run
// -----------------

func cmdRun(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	mqttHost := fs.String("mqtt-host", "localhost", "MQTT broker host")
	mqttPort := fs.Int("mqtt-port", 1883, "MQTT broker port")
	qdrantURL := fs.String("qdrant-url", envOr("QDRANT_URL", "localhost:6333"), "Qdrant HTTP health (6333) or host:port; gRPC is 6334")
	rolesCSV := fs.String("roles", "developer,reviewer,approver,tester", "Comma-separated roles to start")
	verbose := fs.Bool("verbose", false, "Verbose logging")
	_ = fs.Parse(args)

	// Check binaries
	required := []string{"orchestrator", "role-worker", "client"}
	for _, name := range required {
		if _, err := os.Stat(filepath.Join("bin", name)); err != nil {
			logError("missing binary ./bin/%s. Run: ops build", name)
			os.Exit(1)
		}
	}

	// MQTT health (native client, not mosquitto)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mc := mqtt.NewClient(*mqttHost, *mqttPort)
	if err := mc.Connect(ctx); err != nil {
		logError("MQTT connect failed to %s:%d: %v", *mqttHost, *mqttPort, err)
		os.Exit(1)
	}
	mc.Disconnect()
	logInfo("MQTT reachable at %s:%d", *mqttHost, *mqttPort)

	// Qdrant health; if not reachable, try to start local binary
	if err := qdrantHealth(*qdrantURL); err != nil {
		logWarn("Qdrant not reachable (%s): %v (RAG will fallback)", *qdrantURL, err)
		// Attempt to run local qdrant binary from PATH or project bin
		if tryStartQdrantLocal() == nil {
			// give it a moment then re-check
			time.Sleep(600 * time.Millisecond)
			_ = qdrantHealth(*qdrantURL)
		}
	} else {
		logInfo("Qdrant reachable: %s", *qdrantURL)
	}

	// Initialize RAG service (non-fatal if fails)
	if _, err := rag.NewService("qdrant", envOr("QDRANT_GRPC", "localhost:6334")); err != nil {
		logWarn("RAG init warning: %v", err)
	}

	logsDir := filepath.Join("logs")
	_ = os.MkdirAll(logsDir, 0o755)

	// Start orchestrator
	orchCmd := exec.Command(filepath.Join("bin", "orchestrator"), "--mqtt-host", *mqttHost, "--mqtt-port", fmt.Sprint(*mqttPort))
	if *verbose {
		orchCmd.Args = append(orchCmd.Args, "--verbose")
	}
	orchCancel := startProcToLog(orchCmd, filepath.Join(logsDir, "orchestrator.log"))
	defer orchCancel()
	logInfo("orchestrator started")

	// Start role workers concurrently
	roles := splitCSV(*rolesCSV)
	var wg sync.WaitGroup
	stoppers := make([]func(), 0, len(roles))
	for _, role := range roles {
		role := strings.TrimSpace(role)
		if role == "" {
			continue
		}
		wg.Add(1)
		go func(r string) {
			defer wg.Done()
			args := []string{"--id", r + "-worker-1", "--role", r, "--mqtt-host", *mqttHost, "--mqtt-port", fmt.Sprint(*mqttPort)}
			if *verbose {
				args = append(args, "--verbose")
			}
			cmd := exec.Command(filepath.Join("bin", "role-worker"), args...)
			stop := startProcToLog(cmd, filepath.Join(logsDir, r+"-worker.log"))
			stoppers = append(stoppers, stop)
			logInfo("%s worker started", r)
		}(role)
	}
	wg.Wait()

	// Monitor until interrupted
	logInfo("System is running. Press Ctrl+C to stop. Logs: %s", logsDir)
	waitForInterrupt()
	for _, stop := range stoppers {
		stop()
	}
	logSuccess("System stopped")
}

func startProcToLog(cmd *exec.Cmd, logPath string) func() {
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		logError("failed to open log %s: %v", logPath, err)
		// fallback to stdout
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		mw := io.MultiWriter(f)
		cmd.Stdout = mw
		cmd.Stderr = mw
	}
	if err := cmd.Start(); err != nil {
		logError("failed to start %s: %v", strings.Join(cmd.Args, " "), err)
	}
	return func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		if f != nil {
			_ = f.Close()
		}
	}
}

func waitForInterrupt() {
	// Use simple stdin wait to keep cross-platform without extra deps
	// User can press Enter to stop
	fmt.Println("Press Enter to stop...")
	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func qdrantHealth(url string) error {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		client := http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get(strings.TrimRight(url, "/") + "/health")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status: %s", resp.Status)
		}
		return nil
	}
	// Try http if host:port provided
	return qdrantHealth("http://" + url)
}

// -----------------
// lint
// -----------------

func cmdLint(args []string) {
	fs := flag.NewFlagSet("lint", flag.ExitOnError)
	fix := fs.Bool("fix", false, "Attempt to auto-fix formatting and imports")
	verbose := fs.Bool("verbose", false, "Verbose logging")
	all := fs.Bool("all", false, "Run extended checks (benchmarks, profiles if available)")
	_ = fs.Parse(args)

	failures := 0
	increment := func(cond bool) {
		if cond {
			failures++
		}
	}

	// Formatting
	logInfo("Checking formatting (gofmt)")
	if out, err := exec.Command("gofmt", "-l", ".").CombinedOutput(); err != nil {
		logError("gofmt error: %v\n%s", err, string(out))
		increment(true)
	} else if strings.TrimSpace(string(out)) != "" {
		logWarn("files need formatting:\n%s", strings.TrimSpace(string(out)))
		if *fix {
			_ = runLogged(exec.Command("gofmt", "-w", "."), *verbose)
			logInfo("formatted with gofmt")
		}
	}

	// Imports (if goimports available)
	if hasCmd("goimports") {
		logInfo("Checking imports (goimports)")
		if out, err := exec.Command("goimports", "-l", ".").CombinedOutput(); err != nil {
			logError("goimports error: %v\n%s", err, string(out))
			increment(true)
		} else if strings.TrimSpace(string(out)) != "" {
			logWarn("files need import fixes:\n%s", strings.TrimSpace(string(out)))
			if *fix {
				_ = runLogged(exec.Command("goimports", "-w", "."), *verbose)
				logInfo("fixed imports with goimports")
			}
		}
	} else {
		logWarn("goimports not found; skipping import checks")
	}

	// go vet
	logInfo("Running go vet")
	if err := runLogged(exec.Command("go", "vet", "./..."), *verbose); err != nil {
		logError("go vet failed: %v", err)
		increment(true)
	}

	// golangci-lint (optional)
	if hasCmd("golangci-lint") {
		logInfo("Running golangci-lint")
		args := []string{"run", "--timeout=10m"}
		if *verbose {
			args = append(args, "--verbose")
		}
		if *fix {
			args = append(args, "--fix")
		}
		if err := runLogged(exec.Command("golangci-lint", args...), *verbose); err != nil {
			logWarn("golangci-lint reported issues: %v", err)
			increment(true)
		}
	} else {
		logWarn("golangci-lint not found; skipping")
	}

	// Optional extended checks
	if *all {
		_ = runLogged(exec.Command("go", "test", "-bench=.", "-benchmem", "./..."), *verbose)
	}

	report := filepath.Join("lint_report.txt")
	_ = os.WriteFile(report, []byte(fmt.Sprintf("Failed Checks: %d\nTimestamp: %s\n", failures, time.Now().UTC().Format(time.RFC3339))), 0o644)
	if failures > 0 {
		logWarn("lint completed with %d failed checks (see %s)", failures, report)
		os.Exit(1)
	}
	logSuccess("lint passed (report: %s)", report)
}

// -----------------
// cleanup
// -----------------

func cmdCleanup(args []string) {
	fs := flag.NewFlagSet("cleanup", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show actions without executing")
	verbose := fs.Bool("verbose", false, "Verbose logging")
	_ = fs.Parse(args)

	paths := []string{
		"logs/*.log",
		"log/*.log",
		"dist/*",
		"build/*",
		"coverage.out",
	}
	for _, glob := range paths {
		matches, _ := filepath.Glob(glob)
		for _, m := range matches {
			if *dryRun {
				logInfo("[dry-run] remove %s", m)
				continue
			}
			if *verbose {
				logInfo("removing %s", m)
			}
			_ = os.RemoveAll(m)
		}
	}
	if !*dryRun {
		_ = runLogged(exec.Command("go", "clean", "-cache", "-testcache"), *verbose)
	}
	logSuccess("cleanup completed")
}

// -----------------
// mcp
// -----------------

func cmdMCP(args []string) {
	fs := flag.NewFlagSet("mcp", flag.ExitOnError)
	action := fs.String("action", "compose", "Action: compose|check")
	qdrant := fs.String("qdrant-url", envOr("QDRANT_URL", "http://localhost:6333"), "Qdrant URL")
	port := fs.Int("port", 8000, "MCP port")
	_ = fs.Parse(args)

	switch *action {
	case "check":
		if err := qdrantHealth(*qdrant); err != nil {
			logWarn("Qdrant not reachable: %v", err)
			os.Exit(1)
		}
		logSuccess("Qdrant OK: %s", *qdrant)
	case "compose":
		compose := fmt.Sprintf(`version: '3.8'

services:
  qdrant-mcp-server:
    image: qdrant/mcp-server-qdrant:latest
    container_name: qdrant-mcp-server
    ports:
      - "%d:8000"
    environment:
      - QDRANT_URL=%s
      - COLLECTION_NAME=project_knowledge
      - EMBEDDING_MODEL=sentence-transformers/all-MiniLM-L6-v2
      - FASTMCP_HOST=0.0.0.0
      - FASTMCP_PORT=8000
    restart: unless-stopped
`, *port, *qdrant)
		p := filepath.Join("docker-compose.mcp.yml")
		if err := os.WriteFile(p, []byte(compose), 0o644); err != nil {
			logError("failed to write compose: %v", err)
			os.Exit(1)
		}
		logSuccess("wrote %s", p)
	default:
		fmt.Println("Supported actions: compose, check")
		os.Exit(1)
	}
}

// -----------------
// train (wrapper)
// -----------------

func cmdTrain(args []string) {
	fs := flag.NewFlagSet("train", flag.ExitOnError)
	llamaBin := fs.String("llama-bin", envOr("LLAMA_BIN_PATH", "/home/niko/bin"), "Path to llama.cpp binaries")
	models := fs.String("models", envOr("LOCAL_MODELS_PATH", "/data/models"), "Path to models directory")
	output := fs.String("output", envOr("LORA_ADAPTERS_PATH", "/data/models/lora-adapters"), "Output adapters directory")
	_ = fs.Parse(args)

	logInfo("Exporting training data via rag-service")
	if _, err := os.Stat(filepath.Join("bin", "rag-service")); err != nil {
		logError("rag-service not built. Run ops build")
		os.Exit(1)
	}
	trainDir := filepath.Join(os.TempDir(), "training_data")
	_ = os.MkdirAll(trainDir, 0o755)
	trainFile := filepath.Join(trainDir, "training.jsonl")

	cmd := exec.Command(filepath.Join("bin", "rag-service"), "export-training-data", "--format", "llama-finetune")
	out, err := cmd.StdoutPipe()
	if err != nil {
		logError("failed to pipe training data: %v", err)
		os.Exit(1)
	}
	if err := cmd.Start(); err != nil {
		logError("rag-service export failed: %v", err)
		os.Exit(1)
	}
	f, err := os.Create(trainFile)
	if err != nil {
		logError("failed to create %s: %v", trainFile, err)
		os.Exit(1)
	}
	if _, err := io.Copy(f, out); err != nil {
		logError("failed to save training data: %v", err)
		_ = f.Close()
		os.Exit(1)
	}
	_ = f.Close()
	_ = cmd.Wait()
	logInfo("training data saved: %s", trainFile)

	// Check binaries existence lightly
	req := []string{"llama-finetune", "llama-export-lora"}
	for _, r := range req {
		if _, err := os.Stat(filepath.Join(*llamaBin, r)); err != nil {
			logWarn("%s not found in %s; skipping heavy training step", r, *llamaBin)
			logInfo("Training preparation complete. Use your training pipeline with %s", trainFile)
			return
		}
	}

	logInfo("invoke your training pipeline with:\n  %s/llama-finetune --model %s/<BASE_MODEL>.gguf --train-data %s --lora-out %s/<ADAPTER>.bin ...",
		*llamaBin, *models, trainFile, *output)
	logSuccess("train helper completed")
}

// -----------------
// helpers
// -----------------

func runLogged(cmd *exec.Cmd, verbose bool) error {
	cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	if verbose {
		logInfo("exec: %s", strings.Join(cmd.Args, " "))
	}
	out, err := cmd.CombinedOutput()
	if verbose {
		os.Stdout.Write(out)
	}
	if err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func hasCmd(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func logInfo(format string, a ...any) {
	fmt.Printf("[%s] INFO: "+format+"\n", append([]any{time.Now().Format("2006-01-02 15:04:05")}, a...)...)
}

func logWarn(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "[%s] WARN: "+format+"\n", append([]any{time.Now().Format("2006-01-02 15:04:05")}, a...)...)
}

func logError(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "[%s] ERROR: "+format+"\n", append([]any{time.Now().Format("2006-01-02 15:04:05")}, a...)...)
}

func logSuccess(format string, a ...any) {
	fmt.Printf("[%s] SUCCESS: "+format+"\n", append([]any{time.Now().Format("2006-01-02 15:04:05")}, a...)...)
}

func tryStartQdrantLocal() error {
	// Prefer PATH
	if hasCmd("qdrant") {
		cmd := exec.Command("qdrant")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err == nil {
			logInfo("started qdrant from PATH (pid %d)", cmd.Process.Pid)
			return nil
		}
	}
	// Fallback to project bin
	p := filepath.Join("bin", "qdrant")
	if _, err := os.Stat(p); err == nil {
		cmd := exec.Command(p)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err == nil {
			logInfo("started qdrant from project bin (pid %d)", cmd.Process.Pid)
			return nil
		}
	}
	return fmt.Errorf("qdrant binary not found")
}
