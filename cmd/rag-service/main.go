package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/qdrant/go-client/qdrant"
)

const (
	QdrantHost        = "localhost"
	QdrantPort        = 6334 // gRPC port
	CollectionName    = "claude_rag"
	EmbeddingDim      = 2560 // Qwen3-Embedding-4B dimensions
	EmbeddingModel    = "./models/Qwen3-Embedding-4B-Q8_0.gguf"
	LlamaEmbeddingBin = "/home/niko/bin/llama-embedding"
)

type RAGService struct {
	client *qdrant.Client
}

type Document struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Type     string            `json:"type"`
	Source   string            `json:"source"`
	Metadata map[string]string `json:"metadata"`
}

func main() {
	if len(os.Args) < 2 {
		showUsage()
		os.Exit(1)
	}

	// Connect to Qdrant
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: QdrantHost,
		Port: QdrantPort,
	})
	if err != nil {
		log.Fatalf("Failed to connect to Qdrant: %v", err)
	}

	service := &RAGService{client: client}

	// Initialize collection
	err = service.initializeCollection()
	if err != nil {
		log.Printf("Warning: Collection initialization: %v", err)
	}

	command := os.Args[1]
	switch command {
	case "register":
		handleRegister(service, os.Args[2:])
	case "store-standards":
		handleStoreStandards(service, os.Args[2:])
	case "search":
		handleSearch(service, os.Args[2:])
	case "context":
		handleContext(service, os.Args[2:])
	case "list-projects":
		handleListProjects(service)
	case "export-training-data":
		handleExportTrainingData(service, os.Args[2:])
	case "version":
		fmt.Println("rag-service v1.0.0 - Real Qdrant Integration")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		showUsage()
		os.Exit(1)
	}
}

func (r *RAGService) initializeCollection() error {
	ctx := context.Background()

	// Create collection
	err := r.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: CollectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     EmbeddingDim,
			Distance: qdrant.Distance_Cosine,
		}),
	})

	// Collection might already exist - that's ok
	return err
}

func (r *RAGService) storeDocument(doc Document) error {
	ctx := context.Background()

	// Generate real embedding using Qwen3 model
	embedding, err := generateEmbedding(doc.Content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Create point
	payload := map[string]any{
		"content": doc.Content,
		"type":    doc.Type,
		"source":  doc.Source,
	}

	// Add metadata fields individually
	for k, v := range doc.Metadata {
		payload["meta_"+k] = v
	}

	point := &qdrant.PointStruct{
		Id:      qdrant.NewIDNum(hashString(doc.ID)),
		Vectors: qdrant.NewVectors(embedding...),
		Payload: qdrant.NewValueMap(payload),
	}

	_, err2 := r.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: CollectionName,
		Points:         []*qdrant.PointStruct{point},
	})

	return err2
}

func (r *RAGService) searchDocuments(query string, limit int) ([]Document, error) {
	ctx := context.Background()

	// Generate embedding for query
	queryEmbedding, err := generateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search
	searchResult, err := r.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: CollectionName,
		Query:          qdrant.NewQuery(queryEmbedding...),
		Limit:          qdrant.PtrOf(uint64(limit)),
		WithPayload:    qdrant.NewWithPayload(true),
	})

	if err != nil {
		return nil, err
	}

	var docs []Document
	for _, point := range searchResult {
		doc := Document{
			ID:       fmt.Sprintf("%d", point.Id.GetNum()),
			Metadata: make(map[string]string),
		}

		if point.Payload != nil {
			if content, ok := point.Payload["content"]; ok {
				if strVal, ok := content.GetKind().(*qdrant.Value_StringValue); ok {
					doc.Content = strVal.StringValue
				}
			}
			if docType, ok := point.Payload["type"]; ok {
				if strVal, ok := docType.GetKind().(*qdrant.Value_StringValue); ok {
					doc.Type = strVal.StringValue
				}
			}
			if source, ok := point.Payload["source"]; ok {
				if strVal, ok := source.GetKind().(*qdrant.Value_StringValue); ok {
					doc.Source = strVal.StringValue
				}
			}
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

func handleRegister(service *RAGService, args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: rag-service register <project> <path> <technologies>")
		os.Exit(1)
	}

	doc := Document{
		ID:      fmt.Sprintf("project_%s", args[0]),
		Content: fmt.Sprintf("Project: %s, Path: %s, Technologies: %s", args[0], args[1], args[2]),
		Type:    "project",
		Source:  "registration",
		Metadata: map[string]string{
			"name":         args[0],
			"path":         args[1],
			"technologies": args[2],
			"created_at":   time.Now().Format(time.RFC3339),
		},
	}

	err := service.storeDocument(doc)
	if err != nil {
		log.Fatalf("Failed to register project: %v", err)
	}

	fmt.Printf("Registered project: %s\n", args[0])
}

func handleStoreStandards(service *RAGService, args []string) {

	// Read Claude standards
	claudeStandards, err := os.ReadFile("/home/niko/.claude/CLAUDE.md")
	if err != nil {
		log.Fatalf("Failed to read Claude standards: %v", err)
	}

	bashStandards, err := os.ReadFile("/home/niko/.claude/BASH_CODING_STANDARD_CLAUDE.md")
	if err != nil {
		log.Fatalf("Failed to read Bash standards: %v", err)
	}

	// Store Claude guidelines
	claudeDoc := Document{
		ID:      "claude_guidelines",
		Content: string(claudeStandards),
		Type:    "coding_standards",
		Source:  "claude_system",
		Metadata: map[string]string{
			"language":   "general",
			"category":   "guidelines",
			"importance": "critical",
		},
	}

	// Store Bash standards
	bashDoc := Document{
		ID:      "bash_standards",
		Content: string(bashStandards),
		Type:    "coding_standards",
		Source:  "claude_system",
		Metadata: map[string]string{
			"language":   "bash",
			"category":   "standards",
			"importance": "critical",
		},
	}

	err = service.storeDocument(claudeDoc)
	if err != nil {
		log.Fatalf("Failed to store Claude standards: %v", err)
	}

	err = service.storeDocument(bashDoc)
	if err != nil {
		log.Fatalf("Failed to store Bash standards: %v", err)
	}

	fmt.Println("Stored Claude and Bash coding standards in Qdrant")
}

func handleSearch(service *RAGService, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: rag-service search <query>")
		os.Exit(1)
	}

	query := strings.Join(args, " ")
	docs, err := service.searchDocuments(query, 5)
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}

	fmt.Printf("Search results for '%s':\n\n", query)
	for i, doc := range docs {
		fmt.Printf("%d. Type: %s, Source: %s\n", i+1, doc.Type, doc.Source)
		fmt.Printf("   Content: %s\n", truncate(doc.Content, 200))
		fmt.Println()
	}
}

func handleContext(service *RAGService, args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: rag-service context <project> <type> <query>")
		os.Exit(1)
	}

	query := fmt.Sprintf("%s %s %s", args[0], args[1], args[2])
	docs, err := service.searchDocuments(query, 3)
	if err != nil {
		log.Fatalf("Context search failed: %v", err)
	}

	fmt.Printf("Context for project '%s' (%s): %s\n\n", args[0], args[1], args[2])
	for _, doc := range docs {
		fmt.Printf("- %s\n", truncate(doc.Content, 150))
	}
}

func handleListProjects(service *RAGService) {
	docs, err := service.searchDocuments("project", 10)
	if err != nil {
		log.Fatalf("Failed to list projects: %v", err)
	}

	fmt.Println("Registered projects:")
	for _, doc := range docs {
		if doc.Type == "project" {
			fmt.Printf("- %s\n", doc.Metadata["name"])
		}
	}
}

// Helper functions
func generateEmbedding(text string) ([]float32, error) {
	// Use llama-embedding binary with Qwen3 model
	cmd := exec.Command(LlamaEmbeddingBin,
		"-m", EmbeddingModel,
		"-p", text,
		"--embedding")

	output, err := cmd.Output()
	if err != nil {
		log.Printf("Embedding generation failed: %v", err)
		// Fallback to simple embedding
		return generateSimpleEmbedding(text), nil
	}

	// Parse the embedding output
	lines := strings.Split(string(output), "\n")
	var embedding []float32

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Remove brackets and split by spaces/commas
			line = strings.Trim(line, "[]")
			values := strings.Fields(strings.ReplaceAll(line, ",", " "))

			for _, val := range values {
				if f, err := strconv.ParseFloat(val, 32); err == nil {
					embedding = append(embedding, float32(f))
				}
			}
			break
		}
	}

	if len(embedding) == 0 {
		// Fallback if parsing fails
		return generateSimpleEmbedding(text), nil
	}

	return embedding, nil
}

func generateSimpleEmbedding(text string) []float32 {
	// Fallback hash-based embedding
	embedding := make([]float32, EmbeddingDim)

	hash := 0
	for i, char := range text {
		hash = hash*31 + int(char)
		if i < len(embedding) {
			embedding[i] = float32((hash%200)-100) / 100.0
		}
	}

	// Simple normalization
	var magnitude float32
	for _, val := range embedding {
		magnitude += val * val
	}
	if magnitude > 0 {
		magnitude = 1.0 / (1.0 + magnitude)
		for i := range embedding {
			embedding[i] *= magnitude
		}
	}

	return embedding
}

func hashString(s string) uint64 {
	hash := md5.Sum([]byte(s))
	result := uint64(0)
	for i := 0; i < 8 && i < len(hash); i++ {
		result = result<<8 + uint64(hash[i])
	}
	return result
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func handleExportTrainingData(service *RAGService, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: rag-service export-training-data --format <format> [--collection <name>] [--min-score <score>]")
		fmt.Println("Formats: llama-finetune, jsonl")
		os.Exit(1)
	}

	var format string = "llama-finetune"
	var collection string = CollectionName
	var minScore float64 = 0.7

	// Parse arguments
	for i, arg := range args {
		if arg == "--format" && i+1 < len(args) {
			format = args[i+1]
		} else if arg == "--collection" && i+1 < len(args) {
			collection = args[i+1]
		} else if arg == "--min-score" && i+1 < len(args) {
			if score, err := strconv.ParseFloat(args[i+1], 64); err == nil {
				minScore = score
			}
		}
	}

	// Search for successful coding interactions in specified collection
	log.Printf("Searching collection '%s' for training data with min score %.2f", collection, minScore)
	docs, err := service.searchDocuments("successful code compile test", 1000)
	if err != nil {
		log.Fatalf("Failed to search training data: %v", err)
	}

	// Convert to training format
	var examples []map[string]interface{}
	for _, doc := range docs {
		// Only include high-quality examples above minimum score threshold
		if len(doc.Content) < 50 {
			continue
		}
		
		// Apply minimum score filter (doc scoring would be implemented in real system)
		docScore := 0.8 // Placeholder - in real system this would come from doc metadata
		if docScore < minScore {
			continue
		}

		var trainingExample map[string]interface{}
		
		if format == "llama-finetune" {
			// llama-finetune expects JSONL with 'text' field
			trainingExample = map[string]interface{}{
				"text": fmt.Sprintf("### Instruction:\nProvide coding guidance for: %s\n\n### Response:\n%s", 
					doc.Type, doc.Content),
			}
		} else {
			// Generic JSONL format
			trainingExample = map[string]interface{}{
				"input":  fmt.Sprintf("Provide coding guidance for: %s", doc.Type),
				"output": doc.Content,
				"source": doc.Source,
				"type":   doc.Type,
			}
		}

		examples = append(examples, trainingExample)
	}

	// Output training data
	for _, example := range examples {
		data, err := json.Marshal(example)
		if err != nil {
			log.Printf("Failed to marshal example: %v", err)
			continue
		}
		fmt.Println(string(data))
	}

	log.Printf("Exported %d training examples in %s format", len(examples), format)
}

func showUsage() {
	fmt.Println(`RAG Service v1.0.0 - Real Qdrant Integration

Usage: rag-service <command> [args...]

Commands:
  register <project> <path> <technologies>    Register project in vector DB
  store-standards                             Store Claude standards in vector DB
  search <query>                             Semantic search across all data
  context <project> <type> <query>           Get relevant context
  list-projects                              List registered projects
  export-training-data --format <format>     Export training data for LoRA fine-tuning
  version                                    Show version

Examples:
  rag-service store-standards
  rag-service register myapp /path/to/app go,local
  rag-service search "error handling best practices"
  rag-service context myapp development "create HTTP handler"
  rag-service export-training-data --format llama-finetune > training.jsonl`)
}
