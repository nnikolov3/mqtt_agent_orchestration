package health

import (
	"context"
	"fmt"
	"time"

	"github.com/niko/mqtt-agent-orchestration/internal/mqtt"
	"github.com/niko/mqtt-agent-orchestration/internal/rag"
)

// HealthChecker provides comprehensive health checking
type HealthChecker struct {
	ragService  *rag.Service
	mqttClient *mqtt.Client
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(ragService *rag.Service, mqttClient *mqtt.Client) *HealthChecker {
	return &HealthChecker{
		ragService:  ragService,
		mqttClient: mqttClient,
	}
}

// Check performs comprehensive health check
func (h *HealthChecker) Check(ctx context.Context) map[string]string {
	results := make(map[string]string)
	
	// Check RAG service
	if h.ragService != nil {
		if h.ragService.IsAvailable(ctx) {
			results["qdrant"] = "healthy"
		} else {
			results["qdrant"] = "unhealthy"
		}
	} else {
		results["qdrant"] = "not_initialized"
	}
	
	// Check MQTT connection
	if h.mqttClient != nil {
		if h.mqttClient.IsConnected() {
			results["mqtt"] = "healthy"
		} else {
			results["mqtt"] = "disconnected"
		}
	} else {
		results["mqtt"] = "not_initialized"
	}
	
	return results
}

// IsHealthy returns true if all components are healthy
func (h *HealthChecker) IsHealthy(ctx context.Context) bool {
	results := h.Check(ctx)
	
	for _, status := range results {
		if status != "healthy" {
			return false
		}
	}
	
	return true
}

// GetHealthReport returns detailed health report
func (h *HealthChecker) GetHealthReport(ctx context.Context) string {
	results := h.Check(ctx)
	
	report := "Health Check Report:\n"
	for component, status := range results {
		report += fmt.Sprintf("  %s: %s\n", component, status)
	}
	
	return report
}
