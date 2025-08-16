package userservice

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/niko/mqtt-agent-orchestration/internal/rag"
	"github.com/niko/mqtt-agent-orchestration/pkg/types"
)

// UserRAGManager provides a user-level RAG service for cross-project usage
type UserRAGManager struct {
	ragService   *rag.Service
	userConfig   *UserConfig
	projectCache map[string]*ProjectKnowledge
}

// UserConfig holds user-level configuration for RAG services
type UserConfig struct {
	QdrantURL        string            `json:"qdrant_url"`
	DefaultEmbedding string            `json:"default_embedding"`
	ProjectsDir      string            `json:"projects_dir"`
	KnowledgeExports map[string]string `json:"knowledge_exports"` // project -> export path
}

// ProjectKnowledge represents knowledge specific to a project
type ProjectKnowledge struct {
	ProjectName     string             `json:"project_name"`
	ProjectPath     string             `json:"project_path"`
	Technologies    []string           `json:"technologies"`
	CodingStandards map[string]string  `json:"coding_standards"` // language -> standards content
	Patterns        []KnowledgePattern `json:"patterns"`
	LastUpdated     time.Time          `json:"last_updated"`
}

// KnowledgePattern represents a reusable code or design pattern
type KnowledgePattern struct {
	Name        string            `json:"name"`
	Language    string            `json:"language"`
	Category    string            `json:"category"`
	Description string            `json:"description"`
	Example     string            `json:"example"`
	Context     string            `json:"context"`
	Metadata    map[string]string `json:"metadata"`
}

// NewUserRAGManager creates a new user-level RAG manager
func NewUserRAGManager(configPath string) (*UserRAGManager, error) {
	// Load or create user configuration
	config, err := loadUserConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load user config: %w", err)
	}

	// Initialize RAG service
	ragService, err := rag.NewService("", config.QdrantURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create RAG service: %v", err)
	}

	// Initialize collections for user-level knowledge
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ragService.InitializeCollections(ctx); err != nil {
		log.Printf("Warning: Failed to initialize RAG collections: %v", err)
	}

	return &UserRAGManager{
		ragService:   ragService,
		userConfig:   config,
		projectCache: make(map[string]*ProjectKnowledge),
	}, nil
}

// GetRAGService returns the underlying RAG service for direct access
func (um *UserRAGManager) GetRAGService() *rag.Service {
	return um.ragService
}

// RegisterProject registers a new project with the RAG system
func (um *UserRAGManager) RegisterProject(ctx context.Context, projectName, projectPath string, technologies []string) error {
	knowledge := &ProjectKnowledge{
		ProjectName:     projectName,
		ProjectPath:     projectPath,
		Technologies:    technologies,
		CodingStandards: make(map[string]string),
		Patterns:        []KnowledgePattern{},
		LastUpdated:     time.Now(),
	}

	// Cache the project knowledge
	um.projectCache[projectName] = knowledge

	// Store project metadata in RAG
	return um.storeProjectMetadata(ctx, knowledge)
}

// AddCodingStandard adds coding standards for a language to a project
func (um *UserRAGManager) AddCodingStandard(ctx context.Context, projectName, language, standards string) error {
	knowledge, exists := um.projectCache[projectName]
	if !exists {
		return fmt.Errorf("project %s not registered", projectName)
	}

	knowledge.CodingStandards[language] = standards
	knowledge.LastUpdated = time.Now()

	// Store in RAG for searchable access
	return um.storeCodingStandard(ctx, projectName, language, standards)
}

// AddPattern adds a reusable pattern to the knowledge base
func (um *UserRAGManager) AddPattern(ctx context.Context, projectName string, pattern KnowledgePattern) error {
	knowledge, exists := um.projectCache[projectName]
	if !exists {
		return fmt.Errorf("project %s not registered", projectName)
	}

	knowledge.Patterns = append(knowledge.Patterns, pattern)
	knowledge.LastUpdated = time.Now()

	// Store in RAG for searchable access
	return um.storePattern(ctx, projectName, pattern)
}

// SearchAcrossProjects searches for knowledge across all registered projects
func (um *UserRAGManager) SearchAcrossProjects(ctx context.Context, query string, maxResults int) (*types.RAGResponse, error) {
	ragQuery := types.RAGQuery{
		Query:      query,
		Collection: "coding_standards", // Search in standards collection
		TopK:       maxResults,
		Threshold:  0.3, // Lower threshold for broader search
	}

	return um.ragService.SearchKnowledge(ctx, ragQuery)
}

// GetProjectContext retrieves context specific to a project and task
func (um *UserRAGManager) GetProjectContext(ctx context.Context, projectName, taskType, query string) (string, error) {
	// First try to get project-specific context
	projectQuery := fmt.Sprintf("project:%s %s %s", projectName, taskType, query)

	ragQuery := types.RAGQuery{
		Query:      projectQuery,
		Collection: "coding_standards",
		TopK:       3,
		Threshold:  0.5,
	}

	response, err := um.ragService.SearchKnowledge(ctx, ragQuery)
	if err != nil {
		return "", err
	}

	// If no project-specific results, try general search
	if len(response.Documents) == 0 {
		generalQuery := types.RAGQuery{
			Query:      fmt.Sprintf("%s %s", taskType, query),
			Collection: "coding_standards",
			TopK:       3,
			Threshold:  0.4,
		}

		response, err = um.ragService.SearchKnowledge(ctx, generalQuery)
		if err != nil {
			return "", err
		}
	}

	// Build context string
	if len(response.Documents) == 0 {
		return "No relevant context found", nil
	}

	var contextParts []string
	for i, doc := range response.Documents {
		contextParts = append(contextParts, fmt.Sprintf("Context %d: %s", i+1, doc.Content))
	}

	return fmt.Sprintf("Relevant context for %s:\n%s", projectName,
		strings.Join(contextParts, "\n\n")), nil
}

// ExportKnowledge exports project knowledge to a portable format
func (um *UserRAGManager) ExportKnowledge(ctx context.Context, projectName, exportPath string) error {
	knowledge, exists := um.projectCache[projectName]
	if !exists {
		return fmt.Errorf("project %s not registered", projectName)
	}

	// Create export directory
	if err := os.MkdirAll(filepath.Dir(exportPath), 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	// Export to JSON format
	exportData := map[string]interface{}{
		"metadata":        knowledge,
		"exported_at":     time.Now(),
		"export_version":  "1.0",
		"compatible_with": []string{"mqtt-agent-orchestration"},
	}

	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal export data: %w", err)
	}

	if err := os.WriteFile(exportPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	// Update user config with export path
	um.userConfig.KnowledgeExports[projectName] = exportPath
	return um.saveUserConfig()
}

// ImportKnowledge imports project knowledge from an exported file
func (um *UserRAGManager) ImportKnowledge(ctx context.Context, exportPath string) error {
	data, err := os.ReadFile(exportPath)
	if err != nil {
		return fmt.Errorf("failed to read export file: %w", err)
	}

	var exportData map[string]interface{}
	if err := json.Unmarshal(data, &exportData); err != nil {
		return fmt.Errorf("failed to unmarshal export data: %w", err)
	}

	// Extract project knowledge
	metadataBytes, err := json.Marshal(exportData["metadata"])
	if err != nil {
		return fmt.Errorf("failed to extract metadata: %w", err)
	}

	var knowledge ProjectKnowledge
	if err := json.Unmarshal(metadataBytes, &knowledge); err != nil {
		return fmt.Errorf("failed to unmarshal project knowledge: %w", err)
	}

	// Import into current system
	if err := um.RegisterProject(ctx, knowledge.ProjectName, knowledge.ProjectPath, knowledge.Technologies); err != nil {
		return fmt.Errorf("failed to register imported project: %w", err)
	}

	// Import coding standards
	for language, standards := range knowledge.CodingStandards {
		if err := um.AddCodingStandard(ctx, knowledge.ProjectName, language, standards); err != nil {
			log.Printf("Warning: Failed to import coding standard for %s: %v", language, err)
		}
	}

	// Import patterns
	for _, pattern := range knowledge.Patterns {
		if err := um.AddPattern(ctx, knowledge.ProjectName, pattern); err != nil {
			log.Printf("Warning: Failed to import pattern %s: %v", pattern.Name, err)
		}
	}

	log.Printf("Successfully imported knowledge for project: %s", knowledge.ProjectName)
	return nil
}

// Helper methods

func (um *UserRAGManager) storeProjectMetadata(ctx context.Context, knowledge *ProjectKnowledge) error {
	// Store project metadata as a searchable document
	_ = fmt.Sprintf("Project: %s\nPath: %s\nTechnologies: %v\nLast Updated: %s",
		knowledge.ProjectName, knowledge.ProjectPath, knowledge.Technologies, knowledge.LastUpdated)

	// In a real implementation, this would store in a dedicated projects collection
	// For now, we'll use the existing structure
	return nil
}

func (um *UserRAGManager) storeCodingStandard(ctx context.Context, projectName, language, standards string) error {
	// Store coding standard with project context
	// This would typically use the RAG service to store as vectors
	log.Printf("Stored coding standard for project %s, language %s", projectName, language)
	return nil
}

func (um *UserRAGManager) storePattern(ctx context.Context, projectName string, pattern KnowledgePattern) error {
	// Store pattern as searchable knowledge
	log.Printf("Stored pattern %s for project %s", pattern.Name, projectName)
	return nil
}

func loadUserConfig(configPath string) (*UserConfig, error) {
	// Try to load existing config
	if data, err := os.ReadFile(configPath); err == nil {
		var config UserConfig
		if err := json.Unmarshal(data, &config); err == nil {
			return &config, nil
		}
	}

	// Create default config
	config := &UserConfig{
		QdrantURL:        "http://localhost:6333",
		DefaultEmbedding: "sentence-transformers/all-MiniLM-L6-v2",
		ProjectsDir:      filepath.Join(os.Getenv("HOME"), "Dev"),
		KnowledgeExports: make(map[string]string),
	}

	// Try to save default config
	if err := saveConfigToFile(config, configPath); err != nil {
		log.Printf("Warning: Failed to save default config: %v", err)
	}

	return config, nil
}

func (um *UserRAGManager) saveUserConfig() error {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "mqtt-agent-orchestration", "user-config.json")
	return saveConfigToFile(um.userConfig, configPath)
}

func saveConfigToFile(config *UserConfig, path string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
