package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/niko/mqtt-agent-orchestration/internal/worker"
)

var (
	mqttHost   = flag.String("mqtt-host", "localhost", "MQTT broker host")
	mqttPort   = flag.Int("mqtt-port", 1883, "MQTT broker port")
	modelPath  = flag.String("model-path", "/data/models/Qwen3-Embedding-4B-Q8_0.gguf", "Path to embedding model")
	help       = flag.Bool("help", false, "Show help message")
)

func main() {
	flag.Parse()

	if *help {
		showUsage()
		os.Exit(0)
	}

	// Create embedding worker
	embeddingWorker := worker.NewEmbeddingWorker(*mqttHost, *mqttPort, *modelPath)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping embedding worker...")
		cancel()
	}()

	// Start the embedding worker
	log.Printf("Starting embedding worker on %s:%d with model: %s", *mqttHost, *mqttPort, *modelPath)
	
	if err := embeddingWorker.Start(ctx); err != nil {
		log.Fatalf("Failed to start embedding worker: %v", err)
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := embeddingWorker.Stop(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("Embedding worker stopped gracefully")
}

func showUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExample:\n")
	fmt.Fprintf(os.Stderr, "  %s --mqtt-host localhost --mqtt-port 1883 --model-path /data/models/Qwen3-Embedding-4B-Q8_0.gguf\n", os.Args[0])
}
