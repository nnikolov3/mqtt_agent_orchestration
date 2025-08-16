package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/niko/mqtt-agent-orchestration/internal/localmodels"
	"github.com/niko/mqtt-agent-orchestration/pkg/types"
	"github.com/qdrant/go-client/qdrant"
)

// TrainingDataExporter handles export of RAG data for training
type TrainingDataExporter struct {
	service *Service
}

// NewTrainingDataExporter creates a new training data exporter
func NewTrainingDataExporter(service *Service) *TrainingDataExporter {
	return &TrainingDataExporter{
		service: service,
	}
}

// ExportTrainingData exports successful interactions from RAG for training
func (e *TrainingDataExporter) ExportTrainingData(ctx context.Context, collection string, minScore float64) ([]localmodels.TrainingExample, error) {
	// Query all documents from the collection with high scores
	searchResult, err := e.service.client.Scroll(ctx, &qdrant.ScrollPoints{
		CollectionName: collection,
		Limit:          qdrant.PtrOf(uint32(1000)), // Batch size
		WithPayload:    qdrant.NewWithPayload(true),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scroll collection %s: %w", collection, err)
	}

	var examples []localmodels.TrainingExample
	
	for _, point := range searchResult {
		example, err := e.extractTrainingExample(point, minScore)
		if err != nil {
			log.Printf("Skipping point %v: %v", point.Id, err)
			continue
		}
		
		if example != nil {
			examples = append(examples, *example)
		}
	}

	log.Printf("Exported %d training examples from collection %s", len(examples), collection)
	return examples, nil
}

// extractTrainingExample converts a Qdrant point to a training example
func (e *TrainingDataExporter) extractTrainingExample(point *qdrant.RetrievedPoint, minScore float64) (*localmodels.TrainingExample, error) {
	if point.Payload == nil {
		return nil, fmt.Errorf("no payload found")
	}

	// Extract required fields
	var input, output string
	var score float64 = 1.0 // Default score

	// Get input (query or prompt)
	if inputField, exists := point.Payload["input"]; exists {
		if stringValue, ok := inputField.GetKind().(*qdrant.Value_StringValue); ok {
			input = stringValue.StringValue
		}
	} else if promptField, exists := point.Payload["prompt"]; exists {
		if stringValue, ok := promptField.GetKind().(*qdrant.Value_StringValue); ok {
			input = stringValue.StringValue
		}
	}

	// Get output (response or content)
	if outputField, exists := point.Payload["output"]; exists {
		if stringValue, ok := outputField.GetKind().(*qdrant.Value_StringValue); ok {
			output = stringValue.StringValue
		}
	} else if contentField, exists := point.Payload["content"]; exists {
		if stringValue, ok := contentField.GetKind().(*qdrant.Value_StringValue); ok {
			output = stringValue.StringValue
		}
	}

	// Get score (for reinforcement learning)
	if scoreField, exists := point.Payload["score"]; exists {
		if doubleValue, ok := scoreField.GetKind().(*qdrant.Value_DoubleValue); ok {
			score = doubleValue.DoubleValue
		} else if intValue, ok := scoreField.GetKind().(*qdrant.Value_IntegerValue); ok {
			score = float64(intValue.IntegerValue)
		}
	}

	// Skip examples below minimum score threshold
	if score < minScore {
		return nil, fmt.Errorf("score %.2f below threshold %.2f", score, minScore)
	}

	// Skip if we don't have both input and output
	if input == "" || output == "" {
		return nil, fmt.Errorf("missing input or output")
	}

	// Clean and validate the training example
	input = strings.TrimSpace(input)
	output = strings.TrimSpace(output)

	if len(input) < 10 || len(output) < 10 {
		return nil, fmt.Errorf("input or output too short")
	}

	return &localmodels.TrainingExample{
		Input:  input,
		Output: output,
		Score:  score,
	}, nil
}

// ExportSuccessfulCodeInteractions exports coding interactions that compiled and passed tests
func (e *TrainingDataExporter) ExportSuccessfulCodeInteractions(ctx context.Context) ([]localmodels.TrainingExample, error) {
	// Query for successful coding interactions
	query := types.RAGQuery{
		Query:      "successful code compile test pass",
		Collection: "coding_standards",
		TopK:       1000,
		Threshold:  0.0, // Get all results, filter by score later
	}

	response, err := e.service.SearchKnowledge(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search successful interactions: %w", err)
	}

	var examples []localmodels.TrainingExample
	
	for _, doc := range response.Documents {
		// Only include high-score examples (successful interactions)
		if doc.Score < 0.8 {
			continue
		}

		// Try to extract input/output from the content
		example := e.parseCodeInteraction(doc.Content, doc.Score)
		if example != nil {
			examples = append(examples, *example)
		}
	}

	log.Printf("Exported %d successful code interaction examples", len(examples))
	return examples, nil
}

// parseCodeInteraction attempts to parse a code interaction into input/output
func (e *TrainingDataExporter) parseCodeInteraction(content string, score float64) *localmodels.TrainingExample {
	// Look for common patterns in successful code interactions
	lines := strings.Split(content, "\n")
	
	var input, output strings.Builder
	var inOutput bool
	
	for _, line := range lines {
		// Detect transition from input to output
		if strings.Contains(strings.ToLower(line), "response:") ||
		   strings.Contains(strings.ToLower(line), "output:") ||
		   strings.Contains(strings.ToLower(line), "solution:") {
			inOutput = true
			continue
		}
		
		if inOutput {
			output.WriteString(line + "\n")
		} else {
			input.WriteString(line + "\n")
		}
	}

	inputStr := strings.TrimSpace(input.String())
	outputStr := strings.TrimSpace(output.String())

	// Validate that we have meaningful input and output
	if len(inputStr) < 20 || len(outputStr) < 20 {
		return nil
	}

	return &localmodels.TrainingExample{
		Input:  inputStr,
		Output: outputStr,
		Score:  score,
	}
}

// StoreTrainingMetrics stores metrics about training data quality in RAG
func (e *TrainingDataExporter) StoreTrainingMetrics(ctx context.Context, examples []localmodels.TrainingExample) error {
	// Calculate metrics
	totalExamples := len(examples)
	var totalScore float64
	var highQualityCount int

	for _, example := range examples {
		totalScore += example.Score
		if example.Score > 0.9 {
			highQualityCount++
		}
	}

	avgScore := totalScore / float64(totalExamples)
	
	// Create metrics document
	metrics := map[string]interface{}{
		"timestamp":           time.Now().Unix(),
		"total_examples":      totalExamples,
		"average_score":       avgScore,
		"high_quality_count":  highQualityCount,
		"high_quality_ratio":  float64(highQualityCount) / float64(totalExamples),
		"export_date":         time.Now().Format("2006-01-02"),
	}

	// Store in RAG database
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// Generate embedding for metrics (for future retrieval)
	embedding := e.service.generateLocalEmbedding("training data quality metrics")
	if embedding == nil {
		log.Printf("Warning: Could not generate embedding for training metrics")
		return nil // Don't fail if embedding generation fails
	}

	// Store metrics point
	point := &qdrant.PointStruct{
		Id:      qdrant.NewIDNum(uint64(time.Now().Unix())),
		Vectors: qdrant.NewVectors(embedding...),
		Payload: qdrant.NewValueMap(map[string]any{
			"content":        string(metricsJSON),
			"content_type":   "training_metrics",
			"source":         "training_exporter",
			"total_examples": totalExamples,
			"average_score":  avgScore,
			"updated_at":     time.Now().Format(time.RFC3339),
		}),
	}

	_, err = e.service.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: "coding_standards",
		Points:         []*qdrant.PointStruct{point},
	})

	if err != nil {
		return fmt.Errorf("failed to store training metrics: %w", err)
	}

	log.Printf("Stored training metrics: %d examples, avg score %.3f", totalExamples, avgScore)
	return nil
}

// ReinforcementLearningExport exports data specifically for reinforcement learning
func (e *TrainingDataExporter) ReinforcementLearningExport(ctx context.Context) ([]localmodels.TrainingExample, error) {
	// Get successful examples (positive rewards)
	successfulExamples, err := e.ExportSuccessfulCodeInteractions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to export successful examples: %w", err)
	}

	// Apply reinforcement learning scoring
	for i := range successfulExamples {
		// Boost scores for examples that demonstrate good practices
		content := strings.ToLower(successfulExamples[i].Input + " " + successfulExamples[i].Output)
		
		// Reward good practices
		if strings.Contains(content, "error handling") {
			successfulExamples[i].Score += 0.1
		}
		if strings.Contains(content, "test") || strings.Contains(content, "testing") {
			successfulExamples[i].Score += 0.1
		}
		if strings.Contains(content, "documentation") || strings.Contains(content, "comment") {
			successfulExamples[i].Score += 0.05
		}
		if strings.Contains(content, "security") || strings.Contains(content, "validation") {
			successfulExamples[i].Score += 0.1
		}

		// Cap the score at 1.0
		if successfulExamples[i].Score > 1.0 {
			successfulExamples[i].Score = 1.0
		}
	}

	log.Printf("Prepared %d examples for reinforcement learning", len(successfulExamples))
	return successfulExamples, nil
}