package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/niko/mqtt-agent-orchestration/pkg/userservice"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		showUsage()
		os.Exit(1)
	}
	
	command := os.Args[1]
	
	// Initialize user RAG manager
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "mqtt-agent-orchestration", "user-config.json")
	ragManager, err := userservice.NewUserRAGManager(configPath)
	if err != nil {
		log.Fatalf("Failed to initialize RAG manager: %v", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	switch command {
	case "register":
		handleRegisterProject(ctx, ragManager, os.Args[2:])
	case "add-standard":
		handleAddStandard(ctx, ragManager, os.Args[2:])
	case "add-pattern":
		handleAddPattern(ctx, ragManager, os.Args[2:])
	case "search":
		handleSearch(ctx, ragManager, os.Args[2:])
	case "context":
		handleGetContext(ctx, ragManager, os.Args[2:])
	case "export":
		handleExport(ctx, ragManager, os.Args[2:])
	case "import":
		handleImport(ctx, ragManager, os.Args[2:])
	case "version":
		fmt.Printf("RAG Service v%s\n", version)
	case "help", "--help", "-h":
		showUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		showUsage()
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Printf(`RAG Service v%s - User-level knowledge management for development projects

Usage: rag-service <command> [options]

Commands:
  register <project-name> <project-path> <tech1,tech2,...>
                          Register a new project with technologies

  add-standard <project> <language> <standards-file>
                          Add coding standards for a language to a project

  add-pattern <project> <name> <language> <category> <description> <example>
                          Add a reusable pattern to the knowledge base

  search <query>          Search for knowledge across all projects
  
  context <project> <task-type> <query>
                          Get context for a specific project and task

  export <project> <output-file>
                          Export project knowledge to a file

  import <export-file>    Import project knowledge from an exported file

  version                 Show version information
  help                    Show this help message

Examples:
  rag-service register myapp /home/user/myapp go,docker,postgres
  rag-service add-standard myapp go /home/user/.claude/GO_CODING_STANDARD_CLAUDE.md
  rag-service search "error handling patterns"
  rag-service context myapp development "create HTTP handler"
  rag-service export myapp /home/user/myapp-knowledge.json

Environment:
  QDRANT_URL              Qdrant server URL (default: http://localhost:6333)
  RAG_CONFIG_PATH         Path to RAG configuration file

`, version)
}

func handleRegisterProject(ctx context.Context, ragManager *userservice.UserRAGManager, args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: rag-service register <project-name> <project-path> <tech1,tech2,...>")
		os.Exit(1)
	}
	
	projectName := args[0]
	projectPath := args[1]
	techString := args[2]
	
	// Parse technologies
	var technologies []string
	if techString != "" {
		technologies = parseCommaSeparated(techString)
	}
	
	err := ragManager.RegisterProject(ctx, projectName, projectPath, technologies)
	if err != nil {
		log.Fatalf("Failed to register project: %v", err)
	}
	
	fmt.Printf("Successfully registered project '%s' at %s with technologies: %v\n", 
		projectName, projectPath, technologies)
}

func handleAddStandard(ctx context.Context, ragManager *userservice.UserRAGManager, args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: rag-service add-standard <project> <language> <standards-file>")
		os.Exit(1)
	}
	
	projectName := args[0]
	language := args[1]
	standardsFile := args[2]
	
	// Read standards file
	content, err := os.ReadFile(standardsFile)
	if err != nil {
		log.Fatalf("Failed to read standards file: %v", err)
	}
	
	err = ragManager.AddCodingStandard(ctx, projectName, language, string(content))
	if err != nil {
		log.Fatalf("Failed to add coding standard: %v", err)
	}
	
	fmt.Printf("Successfully added %s coding standards for project '%s'\n", language, projectName)
}

func handleAddPattern(ctx context.Context, ragManager *userservice.UserRAGManager, args []string) {
	if len(args) < 6 {
		fmt.Println("Usage: rag-service add-pattern <project> <name> <language> <category> <description> <example>")
		os.Exit(1)
	}
	
	pattern := userservice.KnowledgePattern{
		Name:        args[1],
		Language:    args[2],
		Category:    args[3],
		Description: args[4],
		Example:     args[5],
		Context:     fmt.Sprintf("Added to project %s", args[0]),
		Metadata:    make(map[string]string),
	}
	pattern.Metadata["added_at"] = time.Now().Format(time.RFC3339)
	
	err := ragManager.AddPattern(ctx, args[0], pattern)
	if err != nil {
		log.Fatalf("Failed to add pattern: %v", err)
	}
	
	fmt.Printf("Successfully added pattern '%s' to project '%s'\n", pattern.Name, args[0])
}

func handleSearch(ctx context.Context, ragManager *userservice.UserRAGManager, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: rag-service search <query>")
		os.Exit(1)
	}
	
	query := args[0]
	if len(args) > 1 {
		// Join multiple args as a single query
		query = fmt.Sprintf("%s", args)
	}
	
	response, err := ragManager.SearchAcrossProjects(ctx, query, 5)
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}
	
	fmt.Printf("Search results for '%s':\n\n", query)
	if len(response.Documents) == 0 {
		fmt.Println("No results found.")
		return
	}
	
	for i, doc := range response.Documents {
		fmt.Printf("%d. Score: %.3f\n", i+1, doc.Score)
		fmt.Printf("   Source: %s\n", doc.Source)
		fmt.Printf("   Content: %s\n", truncateString(doc.Content, 200))
		if len(doc.Metadata) > 0 {
			fmt.Printf("   Metadata: %v\n", doc.Metadata)
		}
		fmt.Println()
	}
}

func handleGetContext(ctx context.Context, ragManager *userservice.UserRAGManager, args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: rag-service context <project> <task-type> <query>")
		os.Exit(1)
	}
	
	projectName := args[0]
	taskType := args[1]
	query := args[2]
	
	context, err := ragManager.GetProjectContext(ctx, projectName, taskType, query)
	if err != nil {
		log.Fatalf("Failed to get context: %v", err)
	}
	
	fmt.Printf("Context for project '%s', task '%s', query '%s':\n\n", projectName, taskType, query)
	fmt.Println(context)
}

func handleExport(ctx context.Context, ragManager *userservice.UserRAGManager, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: rag-service export <project> <output-file>")
		os.Exit(1)
	}
	
	projectName := args[0]
	outputFile := args[1]
	
	err := ragManager.ExportKnowledge(ctx, projectName, outputFile)
	if err != nil {
		log.Fatalf("Failed to export knowledge: %v", err)
	}
	
	fmt.Printf("Successfully exported knowledge for project '%s' to %s\n", projectName, outputFile)
}

func handleImport(ctx context.Context, ragManager *userservice.UserRAGManager, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: rag-service import <export-file>")
		os.Exit(1)
	}
	
	exportFile := args[0]
	
	err := ragManager.ImportKnowledge(ctx, exportFile)
	if err != nil {
		log.Fatalf("Failed to import knowledge: %v", err)
	}
	
	fmt.Printf("Successfully imported knowledge from %s\n", exportFile)
}

// Helper functions

func parseCommaSeparated(input string) []string {
	if input == "" {
		return nil
	}
	
	var result []string
	for _, item := range strings.Split(input, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}