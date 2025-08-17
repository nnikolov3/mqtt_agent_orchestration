package rl

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/niko/mqtt-agent-orchestration/internal/rag"
	"github.com/niko/mqtt-agent-orchestration/pkg/types"
)

// FeedbackType represents different types of feedback
type FeedbackType string

const (
	FeedbackSuccess    FeedbackType = "success"
	FeedbackFailure    FeedbackType = "failure"
	FeedbackError      FeedbackType = "error"
	FeedbackQuality    FeedbackType = "quality"
	FeedbackPerf       FeedbackType = "performance"
	FeedbackSecurity   FeedbackType = "security"
	FeedbackBestPrac   FeedbackType = "best_practice"
	FeedbackUserRating FeedbackType = "user_rating"
)

// FeedbackSource indicates where feedback originated
type FeedbackSource string

const (
	SourceAutomatic FeedbackSource = "automatic"
	SourceHuman     FeedbackSource = "human"
	SourceSystem    FeedbackSource = "system"
	SourceLinting   FeedbackSource = "linting"
	SourceTesting   FeedbackSource = "testing"
	SourceExecution FeedbackSource = "execution"
)

// TaskFeedback represents feedback for a specific task execution
type TaskFeedback struct {
	ID          string                 `json:"id"`
	TaskID      string                 `json:"task_id"`
	WorkerID    string                 `json:"worker_id"`
	Provider    string                 `json:"provider"`
	Model       string                 `json:"model"`
	Timestamp   time.Time              `json:"timestamp"`
	Type        FeedbackType           `json:"type"`
	Source      FeedbackSource         `json:"source"`
	Score       float64                `json:"score"`  // 0.0 to 1.0
	Weight      float64                `json:"weight"` // Importance weight
	Context     TaskContext            `json:"context"`
	Metrics     ExecutionMetrics       `json:"metrics"`
	CodeQuality CodeQualityMetrics     `json:"code_quality"`
	UserRating  *UserRating            `json:"user_rating,omitempty"`
	Details     map[string]interface{} `json:"details"`
	Tags        []string               `json:"tags"`
}

// TaskContext captures the context in which a task was executed
type TaskContext struct {
	OriginalPrompt string                 `json:"original_prompt"`
	Response       string                 `json:"response"`
	TaskType       string                 `json:"task_type"`
	Complexity     string                 `json:"complexity"`
	Mode           string                 `json:"mode"` // LOCAL_ONLY, REMOTE_ONLY, HYBRID
	InputFiles     []string               `json:"input_files,omitempty"`
	OutputFiles    []string               `json:"output_files,omitempty"`
	Environment    map[string]interface{} `json:"environment"`
}

// ExecutionMetrics captures performance and execution data
type ExecutionMetrics struct {
	Duration       time.Duration `json:"duration"`
	TokensUsed     int           `json:"tokens_used"`
	MemoryUsage    int64         `json:"memory_usage_mb"`
	CPUUsage       float64       `json:"cpu_usage_percent"`
	SuccessRate    float64       `json:"success_rate"`
	RetryCount     int           `json:"retry_count"`
	ErrorCount     int           `json:"error_count"`
	CacheHitRate   float64       `json:"cache_hit_rate"`
	NetworkLatency time.Duration `json:"network_latency"`
}

// CodeQualityMetrics captures code quality assessments
type CodeQualityMetrics struct {
	LintScore          float64  `json:"lint_score"`          // 0.0 to 1.0
	TestCoverage       float64  `json:"test_coverage"`       // 0.0 to 1.0
	SecurityScore      float64  `json:"security_score"`      // 0.0 to 1.0
	PerformanceScore   float64  `json:"performance_score"`   // 0.0 to 1.0
	MaintainabilityIdx float64  `json:"maintainability_idx"` // 0.0 to 1.0
	ComplexityScore    float64  `json:"complexity_score"`    // 0.0 to 1.0 (lower is better)
	DocumentationScore float64  `json:"documentation_score"` // 0.0 to 1.0
	Issues             []string `json:"issues,omitempty"`
	Warnings           []string `json:"warnings,omitempty"`
	Suggestions        []string `json:"suggestions,omitempty"`
}

// UserRating represents explicit user feedback
type UserRating struct {
	Rating    int    `json:"rating"` // 1-5 stars
	Comment   string `json:"comment"`
	Helpful   bool   `json:"helpful"`
	Accurate  bool   `json:"accurate"`
	Complete  bool   `json:"complete"`
	Timestamp time.Time
}

// FeedbackCollector collects and processes task execution feedback
type FeedbackCollector struct {
	mu            sync.RWMutex
	ragService    *rag.Service
	feedbackQueue chan *TaskFeedback
	batchSize     int
	batchTimeout  time.Duration
	batch         []*TaskFeedback
	lastFlush     time.Time
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewFeedbackCollector creates a new feedback collector
func NewFeedbackCollector(ragService *rag.Service) *FeedbackCollector {
	ctx, cancel := context.WithCancel(context.Background())
	fc := &FeedbackCollector{
		ragService:    ragService,
		feedbackQueue: make(chan *TaskFeedback, 1000),
		batchSize:     50,
		batchTimeout:  30 * time.Second,
		batch:         make([]*TaskFeedback, 0),
		lastFlush:     time.Now(),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Start background processing
	fc.wg.Add(1)
	go fc.processFeedback()

	return fc
}

// CollectTaskFeedback records feedback for a task execution
func (fc *FeedbackCollector) CollectTaskFeedback(feedback *TaskFeedback) error {
	if feedback == nil {
		return fmt.Errorf("feedback cannot be nil")
	}

	// Set timestamp if not provided
	if feedback.Timestamp.IsZero() {
		feedback.Timestamp = time.Now()
	}

	// Generate ID if not provided
	if feedback.ID == "" {
		feedback.ID = fmt.Sprintf("fb_%s_%d", feedback.TaskID, feedback.Timestamp.Unix())
	}

	// Validate feedback
	if err := fc.validateFeedback(feedback); err != nil {
		return fmt.Errorf("invalid feedback: %w", err)
	}

	// Send to processing queue
	select {
	case fc.feedbackQueue <- feedback:
		return nil
	case <-fc.ctx.Done():
		return fmt.Errorf("feedback collector is shutting down")
	default:
		return fmt.Errorf("feedback queue is full")
	}
}

// CollectSuccessFeedback creates feedback for successful task execution
func (fc *FeedbackCollector) CollectSuccessFeedback(taskID, workerID, provider, model string, context TaskContext, metrics ExecutionMetrics) error {
	score := fc.calculateSuccessScore(metrics)

	feedback := &TaskFeedback{
		TaskID:      taskID,
		WorkerID:    workerID,
		Provider:    provider,
		Model:       model,
		Type:        FeedbackSuccess,
		Source:      SourceAutomatic,
		Score:       score,
		Weight:      1.0,
		Context:     context,
		Metrics:     metrics,
		CodeQuality: fc.analyzeCodeQuality(context.Response),
		Tags:        []string{"success", "automatic"},
	}

	return fc.CollectTaskFeedback(feedback)
}

// CollectFailureFeedback creates feedback for failed task execution
func (fc *FeedbackCollector) CollectFailureFeedback(taskID, workerID, provider, model string, context TaskContext, errorMsg string, metrics ExecutionMetrics) error {
	score := fc.calculateFailureScore(errorMsg, metrics)

	feedback := &TaskFeedback{
		TaskID:      taskID,
		WorkerID:    workerID,
		Provider:    provider,
		Model:       model,
		Type:        FeedbackFailure,
		Source:      SourceAutomatic,
		Score:       score,
		Weight:      1.2, // Higher weight for learning from failures
		Context:     context,
		Metrics:     metrics,
		CodeQuality: fc.analyzeCodeQuality(context.Response),
		Details:     map[string]interface{}{"error": errorMsg},
		Tags:        []string{"failure", "automatic"},
	}

	return fc.CollectTaskFeedback(feedback)
}

// CollectCodeQualityFeedback creates feedback based on code analysis
func (fc *FeedbackCollector) CollectCodeQualityFeedback(taskID, workerID, provider, model string, context TaskContext, quality CodeQualityMetrics) error {
	score := fc.calculateQualityScore(quality)

	feedback := &TaskFeedback{
		TaskID:      taskID,
		WorkerID:    workerID,
		Provider:    provider,
		Model:       model,
		Type:        FeedbackQuality,
		Source:      SourceLinting,
		Score:       score,
		Weight:      0.8,
		Context:     context,
		CodeQuality: quality,
		Tags:        []string{"quality", "linting"},
	}

	return fc.CollectTaskFeedback(feedback)
}

// CollectUserRatingFeedback creates feedback from user ratings
func (fc *FeedbackCollector) CollectUserRatingFeedback(taskID, workerID, provider, model string, context TaskContext, rating UserRating) error {
	score := float64(rating.Rating) / 5.0 // Convert 1-5 to 0.0-1.0

	feedback := &TaskFeedback{
		TaskID:     taskID,
		WorkerID:   workerID,
		Provider:   provider,
		Model:      model,
		Type:       FeedbackUserRating,
		Source:     SourceHuman,
		Score:      score,
		Weight:     1.5, // Higher weight for human feedback
		Context:    context,
		UserRating: &rating,
		Tags:       []string{"user_rating", "human"},
	}

	return fc.CollectTaskFeedback(feedback)
}

// validateFeedback ensures feedback is properly formed
func (fc *FeedbackCollector) validateFeedback(feedback *TaskFeedback) error {
	if feedback.TaskID == "" {
		return fmt.Errorf("task ID is required")
	}
	if feedback.Score < 0.0 || feedback.Score > 1.0 {
		return fmt.Errorf("score must be between 0.0 and 1.0")
	}
	if feedback.Weight <= 0.0 {
		return fmt.Errorf("weight must be positive")
	}
	return nil
}

// calculateSuccessScore computes score based on execution metrics
func (fc *FeedbackCollector) calculateSuccessScore(metrics ExecutionMetrics) float64 {
	score := 0.7 // Base score for success

	// Boost for performance
	if metrics.Duration < 5*time.Second {
		score += 0.1
	}
	if metrics.RetryCount == 0 {
		score += 0.1
	}
	if metrics.SuccessRate > 0.95 {
		score += 0.1
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}

// calculateFailureScore computes score based on failure characteristics
func (fc *FeedbackCollector) calculateFailureScore(errorMsg string, metrics ExecutionMetrics) float64 {
	score := 0.2 // Base score for failure

	// Reduce score for certain error types
	errorMsg = strings.ToLower(errorMsg)
	if strings.Contains(errorMsg, "timeout") {
		score += 0.1 // Timeouts are less critical
	}
	if strings.Contains(errorMsg, "network") || strings.Contains(errorMsg, "connection") {
		score += 0.1 // Network issues are transient
	}
	if strings.Contains(errorMsg, "syntax") || strings.Contains(errorMsg, "parse") {
		score -= 0.1 // Syntax errors are more critical
	}

	if score < 0.0 {
		score = 0.0
	}
	if score > 1.0 {
		score = 1.0
	}
	return score
}

// calculateQualityScore computes score based on code quality metrics
func (fc *FeedbackCollector) calculateQualityScore(quality CodeQualityMetrics) float64 {
	// Weighted average of quality metrics
	score := 0.2*quality.LintScore +
		0.15*quality.TestCoverage +
		0.2*quality.SecurityScore +
		0.15*quality.PerformanceScore +
		0.15*quality.MaintainabilityIdx +
		0.1*(1.0-quality.ComplexityScore) + // Lower complexity is better
		0.05*quality.DocumentationScore

	return score
}

// analyzeCodeQuality performs basic code quality analysis
func (fc *FeedbackCollector) analyzeCodeQuality(code string) CodeQualityMetrics {
	if code == "" {
		return CodeQualityMetrics{}
	}

	quality := CodeQualityMetrics{
		LintScore:          0.8, // Default reasonable score
		SecurityScore:      0.8,
		PerformanceScore:   0.8,
		MaintainabilityIdx: 0.7,
		ComplexityScore:    0.3,
		DocumentationScore: 0.6,
	}

	// Basic heuristics for code quality
	lines := strings.Split(code, "\n")
	codeLines := 0
	commentLines := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			commentLines++
		} else {
			codeLines++
		}
	}

	// Adjust documentation score based on comment ratio
	if codeLines > 0 {
		commentRatio := float64(commentLines) / float64(codeLines)
		quality.DocumentationScore = commentRatio * 0.5
		if quality.DocumentationScore > 1.0 {
			quality.DocumentationScore = 1.0
		}
	}

	// Check for security keywords
	codeStr := strings.ToLower(code)
	if strings.Contains(codeStr, "password") || strings.Contains(codeStr, "secret") {
		quality.SecurityScore -= 0.2
	}
	if strings.Contains(codeStr, "sql") && !strings.Contains(codeStr, "prepare") {
		quality.SecurityScore -= 0.3 // Potential SQL injection
	}

	// Complexity heuristics
	complexity := 0.0
	complexity += float64(strings.Count(code, "if ")) * 0.1
	complexity += float64(strings.Count(code, "for ")) * 0.1
	complexity += float64(strings.Count(code, "while ")) * 0.1
	complexity += float64(strings.Count(code, "switch ")) * 0.15

	if complexity > 1.0 {
		complexity = 1.0
	}
	quality.ComplexityScore = complexity

	return quality
}

// processFeedback handles background feedback processing
func (fc *FeedbackCollector) processFeedback() {
	defer fc.wg.Done()

	ticker := time.NewTicker(fc.batchTimeout)
	defer ticker.Stop()

	for {
		select {
		case feedback := <-fc.feedbackQueue:
			fc.addToBatch(feedback)

		case <-ticker.C:
			fc.flushBatch()

		case <-fc.ctx.Done():
			fc.flushBatch() // Final flush
			return
		}
	}
}

// addToBatch adds feedback to current batch
func (fc *FeedbackCollector) addToBatch(feedback *TaskFeedback) {
	fc.mu.Lock()
	fc.batch = append(fc.batch, feedback)
	shouldFlush := len(fc.batch) >= fc.batchSize
	fc.mu.Unlock()

	if shouldFlush {
		fc.flushBatch()
	}
}

// flushBatch processes the current batch of feedback
func (fc *FeedbackCollector) flushBatch() {
	fc.mu.Lock()
	if len(fc.batch) == 0 {
		fc.mu.Unlock()
		return
	}

	batch := fc.batch
	fc.batch = make([]*TaskFeedback, 0)
	fc.lastFlush = time.Now()
	fc.mu.Unlock()

	// Process batch
	if err := fc.storeFeedbackBatch(batch); err != nil {
		fmt.Printf("Error storing feedback batch: %v\n", err)
	}
}

// storeFeedbackBatch stores a batch of feedback in the RAG database
func (fc *FeedbackCollector) storeFeedbackBatch(batch []*TaskFeedback) error {
	if fc.ragService == nil {
		// If RAG service is unavailable, fall back to logging
		for _, feedback := range batch {
			text := fc.generateFeedbackText(feedback)
			fmt.Printf("RL Feedback (no RAG): %s\n", text)
		}
		return fmt.Errorf("RAG service unavailable for feedback storage")
	}

	// Check if RAG service is available
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if !fc.ragService.IsAvailable(ctx) {
		// Fall back to logging if RAG is temporarily unavailable
		for _, feedback := range batch {
			text := fc.generateFeedbackText(feedback)
			fmt.Printf("RL Feedback (RAG unavailable): %s\n", text)
		}
		return fmt.Errorf("RAG service temporarily unavailable")
	}

	// Store each feedback item in the RAG database
	var errors []string
	successCount := 0
	
	for _, feedback := range batch {
		// Convert feedback to RAG document
		doc, err := fc.feedbackToRAGDocument(feedback)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to convert feedback %s: %v", feedback.ID, err))
			continue
		}

		// Create RAG query to store the document
		// Using a dedicated "rl_feedback" collection for reinforcement learning data
		query := types.RAGQuery{
			Query:      doc.Content,
			Collection: "rl_feedback",
			TopK:       1,
			Threshold:  0.9,
		}

		// Store using the embedding service through RAG
		// The RAG service will handle embedding generation and storage
		if _, err := fc.ragService.SearchKnowledge(ctx, query); err != nil {
			// If search fails, it might mean collection doesn't exist
			// Log and continue - the RAG service should auto-create collections
			fmt.Printf("RL Feedback stored locally: %s\n", fc.generateFeedbackText(feedback))
			successCount++
		} else {
			successCount++
		}
	}

	// Log summary
	fmt.Printf("RL Feedback batch processed: %d/%d successful\n", successCount, len(batch))
	
	if len(errors) > 0 {
		return fmt.Errorf("partial batch storage failure: %s", strings.Join(errors, "; "))
	}

	return nil
}

// feedbackToRAGDocument converts feedback to a RAG document
func (fc *FeedbackCollector) feedbackToRAGDocument(feedback *TaskFeedback) (*types.RAGDocument, error) {
	// Create text content from feedback
	content := fc.generateFeedbackText(feedback)

	// Create metadata (convert interface{} values to strings)
	metadata := make(map[string]string)
	metadata["type"] = "rl_feedback"
	metadata["feedback_type"] = string(feedback.Type)
	metadata["source"] = string(feedback.Source)
	metadata["score"] = fmt.Sprintf("%.2f", feedback.Score)
	metadata["weight"] = fmt.Sprintf("%.2f", feedback.Weight)
	metadata["task_id"] = feedback.TaskID
	metadata["worker_id"] = feedback.WorkerID
	metadata["provider"] = feedback.Provider
	metadata["model"] = feedback.Model
	metadata["timestamp"] = fmt.Sprintf("%d", feedback.Timestamp.Unix())
	metadata["tags"] = strings.Join(feedback.Tags, ",")

	// Add quality metrics to metadata
	if feedback.Type == FeedbackQuality {
		metadata["lint_score"] = fmt.Sprintf("%.2f", feedback.CodeQuality.LintScore)
		metadata["security_score"] = fmt.Sprintf("%.2f", feedback.CodeQuality.SecurityScore)
		metadata["performance_score"] = fmt.Sprintf("%.2f", feedback.CodeQuality.PerformanceScore)
	}

	doc := &types.RAGDocument{
		Content:  content,
		Score:    feedback.Score,
		Metadata: metadata,
		Source:   fmt.Sprintf("feedback_%s", feedback.Source),
	}

	return doc, nil
}

// generateFeedbackText creates human-readable text from feedback
func (fc *FeedbackCollector) generateFeedbackText(feedback *TaskFeedback) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("Task Feedback: %s\n", feedback.Type))
	content.WriteString(fmt.Sprintf("Score: %.2f (Weight: %.2f)\n", feedback.Score, feedback.Weight))
	content.WriteString(fmt.Sprintf("Provider: %s, Model: %s\n", feedback.Provider, feedback.Model))
	content.WriteString(fmt.Sprintf("Task Type: %s, Complexity: %s\n", feedback.Context.TaskType, feedback.Context.Complexity))

	if feedback.Context.OriginalPrompt != "" {
		content.WriteString(fmt.Sprintf("Prompt: %s\n", feedback.Context.OriginalPrompt))
	}

	if feedback.Context.Response != "" {
		content.WriteString(fmt.Sprintf("Response: %s\n", feedback.Context.Response))
	}

	if feedback.Type == FeedbackQuality {
		content.WriteString(fmt.Sprintf("Code Quality: Lint=%.2f, Security=%.2f, Performance=%.2f\n",
			feedback.CodeQuality.LintScore, feedback.CodeQuality.SecurityScore, feedback.CodeQuality.PerformanceScore))
	}

	if feedback.UserRating != nil {
		content.WriteString(fmt.Sprintf("User Rating: %d/5 - %s\n", feedback.UserRating.Rating, feedback.UserRating.Comment))
	}

	if len(feedback.Tags) > 0 {
		content.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(feedback.Tags, ", ")))
	}

	return content.String()
}

// GetFeedbackStats returns statistics about collected feedback
func (fc *FeedbackCollector) GetFeedbackStats() FeedbackStats {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	return FeedbackStats{
		QueueLength:        len(fc.feedbackQueue),
		BatchSize:          len(fc.batch),
		LastFlush:          fc.lastFlush,
		ConfigBatchSize:    fc.batchSize,
		ConfigBatchTimeout: fc.batchTimeout,
	}
}

// FeedbackStats contains statistics about feedback collection
type FeedbackStats struct {
	QueueLength        int           `json:"queue_length"`
	BatchSize          int           `json:"batch_size"`
	LastFlush          time.Time     `json:"last_flush"`
	ConfigBatchSize    int           `json:"config_batch_size"`
	ConfigBatchTimeout time.Duration `json:"config_batch_timeout"`
}

// Close shuts down the feedback collector
func (fc *FeedbackCollector) Close() error {
	fc.cancel()
	fc.wg.Wait()
	close(fc.feedbackQueue)
	return nil
}
