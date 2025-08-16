# Testing Framework (`test/`)

## Overview

The `test/` directory contains comprehensive testing infrastructure for the MQTT Agent Orchestration System. This directory implements the **"Test-Driven Development"** principle with **"Comprehensive Coverage"** goals, ensuring every component is thoroughly tested and validated.

## Architecture Philosophy

Following our **"Excellence through Rigor"** philosophy, testing is:
- **Comprehensive**: 100% code coverage with meaningful tests
- **Automated**: All tests run automatically in CI/CD pipeline
- **Reliable**: Tests are deterministic and repeatable
- **Fast**: Tests execute quickly for rapid feedback

## Testing Strategy

### Test Pyramid

```
┌─────────────────────────────────────────────────────────────┐
│                    E2E Tests (10%)                          │
│              Complete system workflows                      │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                  Integration Tests (20%)                    │
│              Component interaction testing                  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                    Unit Tests (70%)                         │
│              Individual component testing                   │
└─────────────────────────────────────────────────────────────┘
```

### Test Categories

#### 1. Unit Tests (`unit/`)

**Purpose**: Test individual functions and methods in isolation.

**Coverage Requirements**:
- **100% Code Coverage**: Every line of code must be tested
- **Branch Coverage**: All conditional branches must be tested
- **Error Paths**: All error conditions must be tested
- **Edge Cases**: Boundary conditions and edge cases must be tested

**Test Structure**:
```go
// Example unit test
func TestService_NewService(t *testing.T) {
    tests := []struct {
        name        string
        qdrantURL   string
        wantErr     bool
        expectedURL string
    }{
        {
            name:        "valid URL",
            qdrantURL:   "localhost:6333",
            wantErr:     false,
            expectedURL: "localhost:6333",
        },
        {
            name:      "empty URL",
            qdrantURL: "",
            wantErr:   false,
        },
        {
            name:      "invalid URL format",
            qdrantURL: "http://localhost:6333",
            wantErr:   false, // Should handle gracefully
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service, err := NewService("qdrant", tt.qdrantURL)
            
            if tt.wantErr && err == nil {
                t.Error("Expected error but got none")
            }
            if !tt.wantErr && err != nil {
                t.Errorf("Unexpected error: %v", err)
            }
            
            if service != nil && tt.expectedURL != "" {
                if service.qdrantURL != tt.expectedURL {
                    t.Errorf("Expected URL %s, got %s", tt.expectedURL, service.qdrantURL)
                }
            }
        })
    }
}
```

**Usage**:
```bash
# Run all unit tests
go test ./internal/... -v

# Run specific package tests
go test ./internal/rag -v

# Run with coverage
go test ./internal/... -cover

# Run with race detection
go test ./internal/... -race
```

#### 2. Integration Tests (`integration/`)

**Purpose**: Test component interactions and system integration.

**Test Scenarios**:
- **MQTT Communication**: Test message publishing and subscription
- **RAG Operations**: Test vector database operations
- **Model Management**: Test local model loading and inference
- **Worker Coordination**: Test worker task processing
- **API Integration**: Test external AI service integration

**Test Structure**:
```go
// Example integration test
func TestMQTTWorkflowIntegration(t *testing.T) {
    // Setup test environment
    ctx := context.Background()
    
    // Start MQTT broker
    broker := startTestMQTTBroker(t)
    defer broker.Stop()
    
    // Start orchestrator
    orchestrator := startTestOrchestrator(t, broker.Address())
    defer orchestrator.Stop()
    
    // Start worker
    worker := startTestWorker(t, broker.Address(), "developer")
    defer worker.Stop()
    
    // Submit task
    task := &types.WorkflowTask{
        ID:      "test-task-123",
        Content: "Create a simple HTTP server",
        Role:    types.RoleDeveloper,
    }
    
    result, err := submitTask(ctx, broker.Address(), task)
    if err != nil {
        t.Fatalf("Failed to submit task: %v", err)
    }
    
    // Verify result
    if result.Status != "completed" {
        t.Errorf("Expected status 'completed', got '%s'", result.Status)
    }
    
    if result.Content == "" {
        t.Error("Expected non-empty result content")
    }
}
```

**Usage**:
```bash
# Run integration tests
go test ./test/integration -v

# Run with specific services
go test ./test/integration -tags=qdrant,mqtt -v

# Run with external dependencies
go test ./test/integration -tags=external -v
```

#### 3. End-to-End Tests (`e2e/`)

**Purpose**: Test complete system workflows from start to finish.

**Test Scenarios**:
- **Complete Workflow**: Developer → Reviewer → Approver → Tester
- **Error Recovery**: System recovery from failures
- **Performance**: System performance under load
- **Scalability**: System behavior with multiple workers
- **Reliability**: Long-running system stability

**Test Structure**:
```go
// Example E2E test
func TestCompleteWorkflowE2E(t *testing.T) {
    // Setup complete system
    system := setupCompleteSystem(t)
    defer system.Cleanup()
    
    // Start all components
    err := system.Start()
    if err != nil {
        t.Fatalf("Failed to start system: %v", err)
    }
    
    // Wait for system to be ready
    err = system.WaitForReady(30 * time.Second)
    if err != nil {
        t.Fatalf("System not ready: %v", err)
    }
    
    // Submit complex task
    task := &types.WorkflowTask{
        ID:         "e2e-task-123",
        Content:    "Create a complete REST API with authentication",
        Role:       types.RoleDeveloper,
        Complexity: types.ComplexityHigh,
    }
    
    // Execute complete workflow
    result, err := system.ExecuteWorkflow(ctx, task)
    if err != nil {
        t.Fatalf("Workflow execution failed: %v", err)
    }
    
    // Verify complete result
    verifyCompleteResult(t, result)
}
```

**Usage**:
```bash
# Run E2E tests
go test ./test/e2e -v

# Run with specific configuration
go test ./test/e2e -config=e2e-config.yaml -v

# Run performance tests
go test ./test/e2e -tags=performance -v
```

#### 4. Performance Tests (`performance/`)

**Purpose**: Test system performance and scalability.

**Test Metrics**:
- **Throughput**: Tasks processed per second
- **Latency**: Response time for tasks
- **Memory Usage**: Memory consumption under load
- **CPU Usage**: CPU utilization patterns
- **Resource Efficiency**: Resource usage optimization

**Test Structure**:
```go
// Example performance test
func BenchmarkMQTTThroughput(b *testing.B) {
    // Setup performance test environment
    system := setupPerformanceTest(b)
    defer system.Cleanup()
    
    // Reset timer
    b.ResetTimer()
    
    // Run benchmark
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            task := generateTestTask()
            _, err := system.SubmitTask(ctx, task)
            if err != nil {
                b.Fatalf("Task submission failed: %v", err)
            }
        }
    })
}

func TestSystemPerformance(t *testing.T) {
    // Setup load test
    loadTest := setupLoadTest(t, 100) // 100 concurrent workers
    defer loadTest.Cleanup()
    
    // Run load test
    metrics, err := loadTest.Run(5 * time.Minute)
    if err != nil {
        t.Fatalf("Load test failed: %v", err)
    }
    
    // Verify performance requirements
    if metrics.Throughput < 100 {
        t.Errorf("Throughput below requirement: %f tasks/sec", metrics.Throughput)
    }
    
    if metrics.AverageLatency > 2*time.Second {
        t.Errorf("Latency above requirement: %v", metrics.AverageLatency)
    }
    
    if metrics.MemoryUsage > 2*1024*1024*1024 { // 2GB
        t.Errorf("Memory usage above requirement: %d bytes", metrics.MemoryUsage)
    }
}
```

**Usage**:
```bash
# Run performance benchmarks
go test ./test/performance -bench=. -v

# Run load tests
go test ./test/performance -tags=load -v

# Run stress tests
go test ./test/performance -tags=stress -v
```

#### 5. Security Tests (`security/`)

**Purpose**: Test system security and vulnerability assessment.

**Test Areas**:
- **Input Validation**: Test for injection attacks
- **Authentication**: Test authentication mechanisms
- **Authorization**: Test access control
- **Data Protection**: Test data encryption and privacy
- **Network Security**: Test network communication security

**Test Structure**:
```go
// Example security test
func TestInputValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {
            name:    "SQL injection attempt",
            input:   "'; DROP TABLE users; --",
            wantErr: true,
        },
        {
            name:    "XSS attempt",
            input:   "<script>alert('xss')</script>",
            wantErr: true,
        },
        {
            name:    "Path traversal attempt",
            input:   "../../../etc/passwd",
            wantErr: true,
        },
        {
            name:    "Valid input",
            input:   "normal task content",
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateInput(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateInput() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Usage**:
```bash
# Run security tests
go test ./test/security -v

# Run with security scanning
go test ./test/security -tags=scan -v

# Run vulnerability assessment
go test ./test/security -tags=vuln -v
```

## Test Infrastructure

### Test Utilities (`utils/`)

**Mock Services**:
```go
// Mock MQTT client for testing
type MockMQTTClient struct {
    publishedMessages []Message
    subscribedTopics  []string
    connected         bool
}

func (m *MockMQTTClient) Publish(ctx context.Context, topic string, payload []byte) error {
    m.publishedMessages = append(m.publishedMessages, Message{
        Topic:   topic,
        Payload: payload,
    })
    return nil
}

func (m *MockMQTTClient) Subscribe(ctx context.Context, topic string) error {
    m.subscribedTopics = append(m.subscribedTopics, topic)
    return nil
}
```

**Test Helpers**:
```go
// Test helper functions
func setupTestEnvironment(t *testing.T) *TestEnvironment {
    env := &TestEnvironment{
        TempDir: t.TempDir(),
        Config:  loadTestConfig(),
    }
    
    // Setup test database
    env.Database = setupTestDatabase(t, env.TempDir)
    
    // Setup test MQTT broker
    env.MQTTBroker = setupTestMQTTBroker(t)
    
    return env
}

func cleanupTestEnvironment(t *testing.T, env *TestEnvironment) {
    if env.Database != nil {
        env.Database.Close()
    }
    if env.MQTTBroker != nil {
        env.MQTTBroker.Stop()
    }
}
```

### Test Data (`data/`)

**Test Fixtures**:
```yaml
# test/data/tasks.yaml
tasks:
  simple_task:
    id: "task-001"
    content: "Create a simple HTTP server"
    role: "developer"
    complexity: "low"
    
  complex_task:
    id: "task-002"
    content: "Implement a distributed caching system"
    role: "developer"
    complexity: "high"
    
  review_task:
    id: "task-003"
    content: "Review the authentication implementation"
    role: "reviewer"
    complexity: "medium"
```

**Test Configurations**:
```yaml
# test/config/test-config.yaml
mqtt:
  host: "localhost"
  port: 1883
  timeout: "5s"

qdrant:
  url: "http://localhost:6333"
  timeout: "10s"

models:
  qwen-omni:
    path: "/tmp/test-models/qwen-omni.gguf"
    memory_limit: 1024
```

## Test Execution

### Local Testing

**Run All Tests**:
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detection
make test-race

# Run tests with verbose output
make test-verbose
```

**Run Specific Test Categories**:
```bash
# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Run E2E tests only
make test-e2e

# Run performance tests only
make test-performance
```

### CI/CD Integration

**GitHub Actions Workflow**:
```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.24'
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run unit tests
      run: go test ./internal/... -v -race -cover
    
    - name: Run integration tests
      run: go test ./test/integration -v -tags=integration
    
    - name: Run security tests
      run: go test ./test/security -v -tags=security
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
```

### Test Reporting

**Coverage Reports**:
```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View coverage in browser
go tool cover -html=coverage.out -o coverage.html

# Generate coverage badge
gocov convert coverage.out | gocov-html > coverage.html
```

**Test Results**:
```bash
# Generate test report
go test ./... -json > test-results.json

# Generate JUnit XML report
go test ./... -v 2>&1 | go-junit-report > test-results.xml
```

## Test Best Practices

### Test Organization

**File Naming**:
- Unit tests: `service_test.go` (same package as `service.go`)
- Integration tests: `integration_test.go` (in `test/integration/`)
- E2E tests: `e2e_test.go` (in `test/e2e/`)

**Test Structure**:
```go
// Use table-driven tests for multiple scenarios
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        // Test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Test Data Management

**Test Data Isolation**:
```go
// Use unique test data for each test
func TestWithUniqueData(t *testing.T) {
    uniqueID := fmt.Sprintf("test-%d", time.Now().UnixNano())
    
    // Use uniqueID in test
    result := processData(uniqueID)
    
    // Verify result
    if result.ID != uniqueID {
        t.Errorf("Expected ID %s, got %s", uniqueID, result.ID)
    }
}
```

**Test Cleanup**:
```go
// Always clean up test resources
func TestWithCleanup(t *testing.T) {
    // Setup
    tempDir := t.TempDir()
    db := setupTestDB(t, tempDir)
    
    // Test implementation
    result := testFunction(db)
    
    // Cleanup is automatic with t.TempDir()
    // and deferred cleanup functions
}
```

## Performance Testing

### Load Testing

**Load Test Configuration**:
```yaml
# test/performance/load-config.yaml
load_test:
  duration: "5m"
  users: 100
  ramp_up: "30s"
  scenarios:
    - name: "normal_load"
      weight: 70
      requests_per_second: 10
    - name: "peak_load"
      weight: 30
      requests_per_second: 50
```

**Load Test Execution**:
```bash
# Run load test
go test ./test/performance -tags=load -config=load-config.yaml -v

# Run with custom parameters
go test ./test/performance -tags=load \
  -duration=10m \
  -users=200 \
  -ramp-up=1m \
  -v
```

### Stress Testing

**Stress Test Scenarios**:
```go
func TestSystemStress(t *testing.T) {
    // Test system under extreme load
    stressTest := &StressTest{
        Duration:      30 * time.Minute,
        MaxUsers:      1000,
        RampUpTime:    5 * time.Minute,
        FailureRate:   0.01, // 1% failure rate allowed
    }
    
    results := stressTest.Run(t)
    
    // Verify system stability
    if results.FailureRate > 0.01 {
        t.Errorf("Failure rate too high: %.2f%%", results.FailureRate*100)
    }
    
    if results.ResponseTime > 5*time.Second {
        t.Errorf("Response time too high: %v", results.ResponseTime)
    }
}
```

## Security Testing

### Vulnerability Scanning

**Security Test Execution**:
```bash
# Run security tests
go test ./test/security -v

# Run with OWASP ZAP
zap-baseline.py -t http://localhost:8080

# Run with gosec
gosec ./...

# Run with staticcheck
staticcheck ./...
```

### Penetration Testing

**Penetration Test Scenarios**:
```go
func TestSQLInjectionProtection(t *testing.T) {
    payloads := []string{
        "'; DROP TABLE users; --",
        "' OR '1'='1",
        "'; INSERT INTO users VALUES ('hacker', 'password'); --",
    }
    
    for _, payload := range payloads {
        t.Run(fmt.Sprintf("payload_%s", payload), func(t *testing.T) {
            err := processUserInput(payload)
            if err == nil {
                t.Error("Expected error for SQL injection attempt")
            }
        })
    }
}
```

## Troubleshooting

### Common Test Issues

1. **Flaky Tests**: Use proper synchronization and timeouts
2. **Resource Leaks**: Ensure proper cleanup in tests
3. **Test Data Conflicts**: Use unique identifiers and isolated data
4. **Performance Issues**: Optimize test setup and teardown

### Debug Commands

```bash
# Run tests with debug output
go test -v -debug ./...

# Run specific test with verbose output
go test -v -run TestSpecificFunction ./...

# Run tests with timeout
go test -timeout 30s ./...

# Run tests with specific tags
go test -tags=integration,slow ./...
```

## Future Enhancements

### Planned Features

- **Test Parallelization**: Parallel test execution for faster feedback
- **Test Caching**: Cache test results for faster subsequent runs
- **Visual Test Reports**: Enhanced test reporting with visualizations
- **Test Metrics**: Advanced test metrics and analytics

### Extension Points

- **Custom Test Frameworks**: Support for custom testing frameworks
- **Test Data Generation**: Automated test data generation
- **Test Environment Management**: Automated test environment setup
- **Test Result Analysis**: Advanced test result analysis and insights

---

**Production Ready**: The testing framework is designed for production use with comprehensive coverage, automated execution, and reliable results. It provides a robust foundation for ensuring system quality and reliability.
