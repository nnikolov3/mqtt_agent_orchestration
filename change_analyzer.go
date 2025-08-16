package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type ChangeStats struct {
	File           string            `json:"file"`
	ChangeCount    int               `json:"change_count"`
	LastModified   string            `json:"last_modified"`
	Contributors   map[string]int    `json:"contributors"`
	ChangeTypes    map[string]int    `json:"change_types"`
	LinesAdded     int               `json:"lines_added"`
	LinesRemoved   int               `json:"lines_removed"`
	Directory      string            `json:"directory"`
	FileExtension  string            `json:"file_extension"`
}

type DirectoryStats struct {
	Path           string       `json:"path"`
	ChangeCount    int          `json:"change_count"`
	Files          []string     `json:"files"`
	TotalLines     int          `json:"total_lines"`
	Contributors   map[string]int `json:"contributors"`
}

type ChangeReport struct {
	GeneratedAt    string                    `json:"generated_at"`
	Repository     string                    `json:"repository"`
	TimeRange      string                    `json:"time_range"`
	FileStats      map[string]*ChangeStats   `json:"file_stats"`
	DirectoryStats map[string]*DirectoryStats `json:"directory_stats"`
	Summary        struct {
		TotalFiles     int `json:"total_files"`
		TotalChanges   int `json:"total_changes"`
		MostChanged    []string `json:"most_changed"`
		MostActiveDirs []string `json:"most_active_dirs"`
	} `json:"summary"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run change_analyzer.go <repository_path> [time_range]")
		fmt.Println("Example: go run change_analyzer.go /path/to/repo '1 month ago'")
		os.Exit(1)
	}

	repoPath := os.Args[1]
	timeRange := "1 year ago"
	if len(os.Args) > 2 {
		timeRange = os.Args[2]
	}

	// Change to repository directory
	if err := os.Chdir(repoPath); err != nil {
		log.Fatalf("Failed to change to repository directory: %v", err)
	}

	report := analyzeChanges(timeRange)
	
	// Generate JSON report
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal report: %v", err)
	}

	// Write report to file
	outputFile := fmt.Sprintf("change_report_%s.json", time.Now().Format("2006-01-02"))
	if err := os.WriteFile(outputFile, reportJSON, 0644); err != nil {
		log.Fatalf("Failed to write report: %v", err)
	}

	// Print summary
	printSummary(report)
	fmt.Printf("\nDetailed report saved to: %s\n", outputFile)
}

func analyzeChanges(timeRange string) *ChangeReport {
	report := &ChangeReport{
		GeneratedAt:    time.Now().Format(time.RFC3339),
		Repository:     getRepoName(),
		TimeRange:      timeRange,
		FileStats:      make(map[string]*ChangeStats),
		DirectoryStats: make(map[string]*DirectoryStats),
	}

	// Get git log with detailed information
	cmd := exec.Command("git", "log", "--since="+timeRange, "--pretty=format:%H|%an|%ad|%s", "--date=short", "--numstat")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to get git log: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		if strings.Contains(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 4 {
				commitHash := parts[0]
				author := parts[1]
				date := parts[2]
				message := parts[3]
				
				// Get detailed changes for this commit
				analyzeCommit(commitHash, author, date, message, report)
			}
		}
	}

	// Calculate summary statistics
	calculateSummary(report)
	
	return report
}

func analyzeCommit(commitHash, author, date, message string, report *ChangeReport) {
	// Get detailed changes for this commit
	cmd := exec.Command("git", "show", "--numstat", "--format=", commitHash)
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			linesAdded := parseInt(fields[0])
			linesRemoved := parseInt(fields[1])
			filePath := fields[2]
			
			if filePath != "" && !strings.Contains(filePath, "=>") {
				updateFileStats(filePath, author, date, linesAdded, linesRemoved, message, report)
			}
		}
	}
}

func updateFileStats(filePath, author, date string, linesAdded, linesRemoved int, message string, report *ChangeReport) {
	if report.FileStats[filePath] == nil {
		report.FileStats[filePath] = &ChangeStats{
			File:          filePath,
			ChangeCount:   0,
			LastModified:  date,
			Contributors:  make(map[string]int),
			ChangeTypes:   make(map[string]int),
			Directory:     filepath.Dir(filePath),
			FileExtension: filepath.Ext(filePath),
		}
	}

	stats := report.FileStats[filePath]
	stats.ChangeCount++
	stats.LinesAdded += linesAdded
	stats.LinesRemoved += linesRemoved
	stats.Contributors[author]++
	
	// Categorize change type based on message
	changeType := categorizeChange(message)
	stats.ChangeTypes[changeType]++
	
	if date > stats.LastModified {
		stats.LastModified = date
	}

	// Update directory stats
	updateDirectoryStats(filePath, author, linesAdded, linesRemoved, report)
}

func updateDirectoryStats(filePath, author string, linesAdded, linesRemoved int, report *ChangeReport) {
	dir := filepath.Dir(filePath)
	if dir == "." {
		dir = "root"
	}
	
	if report.DirectoryStats[dir] == nil {
		report.DirectoryStats[dir] = &DirectoryStats{
			Path:         dir,
			ChangeCount:  0,
			Files:        []string{},
			Contributors: make(map[string]int),
		}
	}
	
	dirStats := report.DirectoryStats[dir]
	dirStats.ChangeCount++
	dirStats.Contributors[author]++
	
	// Add file to directory if not already present
	found := false
	for _, f := range dirStats.Files {
		if f == filePath {
			found = true
			break
		}
	}
	if !found {
		dirStats.Files = append(dirStats.Files, filePath)
	}
}

func categorizeChange(message string) string {
	message = strings.ToLower(message)
	
	switch {
	case strings.Contains(message, "fix") || strings.Contains(message, "bug"):
		return "bugfix"
	case strings.Contains(message, "feat") || strings.Contains(message, "add"):
		return "feature"
	case strings.Contains(message, "refactor"):
		return "refactor"
	case strings.Contains(message, "test"):
		return "test"
	case strings.Contains(message, "doc") || strings.Contains(message, "readme"):
		return "documentation"
	case strings.Contains(message, "style") || strings.Contains(message, "format"):
		return "style"
	default:
		return "other"
	}
}

func calculateSummary(report *ChangeReport) {
	report.Summary.TotalFiles = len(report.FileStats)
	
	// Calculate total changes
	for _, stats := range report.FileStats {
		report.Summary.TotalChanges += stats.ChangeCount
	}
	
	// Find most changed files
	var files []*ChangeStats
	for _, stats := range report.FileStats {
		files = append(files, stats)
	}
	
	sort.Slice(files, func(i, j int) bool {
		return files[i].ChangeCount > files[j].ChangeCount
	})
	
	for i := 0; i < 10 && i < len(files); i++ {
		report.Summary.MostChanged = append(report.Summary.MostChanged, files[i].File)
	}
	
	// Find most active directories
	var dirs []*DirectoryStats
	for _, stats := range report.DirectoryStats {
		dirs = append(dirs, stats)
	}
	
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].ChangeCount > dirs[j].ChangeCount
	})
	
	for i := 0; i < 5 && i < len(dirs); i++ {
		report.Summary.MostActiveDirs = append(report.Summary.MostActiveDirs, dirs[i].Path)
	}
}

func printSummary(report *ChangeReport) {
	fmt.Printf("\n=== Codebase Change Analysis Report ===\n")
	fmt.Printf("Repository: %s\n", report.Repository)
	fmt.Printf("Time Range: %s\n", report.TimeRange)
	fmt.Printf("Generated: %s\n", report.GeneratedAt)
	fmt.Printf("\nSummary:\n")
	fmt.Printf("- Total Files: %d\n", report.Summary.TotalFiles)
	fmt.Printf("- Total Changes: %d\n", report.Summary.TotalChanges)
	
	fmt.Printf("\nMost Changed Files:\n")
	for i, file := range report.Summary.MostChanged {
		stats := report.FileStats[file]
		fmt.Printf("  %d. %s (%d changes)\n", i+1, file, stats.ChangeCount)
	}
	
	fmt.Printf("\nMost Active Directories:\n")
	for i, dir := range report.Summary.MostActiveDirs {
		stats := report.DirectoryStats[dir]
		fmt.Printf("  %d. %s (%d changes)\n", i+1, dir, stats.ChangeCount)
	}
}

func getRepoName() string {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	
	url := strings.TrimSpace(string(output))
	if strings.Contains(url, "/") {
		parts := strings.Split(url, "/")
		if len(parts) > 0 {
			repo := parts[len(parts)-1]
			if strings.HasSuffix(repo, ".git") {
				repo = repo[:len(repo)-4]
			}
			return repo
		}
	}
	return "unknown"
}

func parseInt(s string) int {
	if s == "-" {
		return 0
	}
	
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}
