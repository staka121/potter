package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/staka121/potter/pkg/types"
)

// ExecutionResult represents the result of implementing a service
type ExecutionResult struct {
	ObjectName   string
	Success      bool
	Error        error
	Response     string
	Duration     time.Duration
	InputTokens  int
	OutputTokens int
}

// Runner executes implementation tasks
type Runner struct {
	client      *ClaudeClient
	generator   *PromptGenerator
	plan        *types.ImplementationPlan
	concurrency int    // 0 = unlimited
	tempDir     string // temporary directory for this run
}

// NewRunner creates a new execution runner
func NewRunner(plan *types.ImplementationPlan) (*Runner, error) {
	client, err := NewClaudeClient()
	if err != nil {
		return nil, err
	}

	// Create timestamped temp directory
	// Format: /tmp/potter/{app-name}/yyyymmddhhmmss
	timestamp := time.Now().Format("20060102150405")
	tempDir := filepath.Join("/tmp", "potter", plan.Tsubo, timestamp)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &Runner{
		client:      client,
		generator:   NewPromptGenerator(plan),
		plan:        plan,
		concurrency: 0, // unlimited by default
		tempDir:     tempDir,
	}, nil
}

// SetConcurrency sets the maximum number of parallel executions
func (r *Runner) SetConcurrency(n int) {
	r.concurrency = n
}

// GetTempDir returns the temporary directory for this run
func (r *Runner) GetTempDir() string {
	return r.tempDir
}

// ExecuteAll executes all waves in the implementation plan
func (r *Runner) ExecuteAll() ([]ExecutionResult, error) {
	var allResults []ExecutionResult
	totalWaves := len(r.plan.Waves)
	totalObjects := 0
	for _, wave := range r.plan.Waves {
		totalObjects += len(wave.Objects)
	}

	fmt.Printf("\n📋 Total: %d wave(s), %d object(s)\n", totalWaves, totalObjects)

	completedObjects := 0

	for waveIdx, wave := range r.plan.Waves {
		fmt.Printf("\n🌊 Wave %d/%d (Wave ID: %d)\n", waveIdx+1, totalWaves, wave.Wave)
		fmt.Printf("   Objects: %d\n", len(wave.Objects))
		fmt.Printf("   Mode: ")
		if wave.Parallel {
			fmt.Println("Parallel")
		} else {
			fmt.Println("Sequential")
		}
		fmt.Println()

		results, err := r.executeWave(wave, &completedObjects, totalObjects)
		if err != nil {
			return allResults, fmt.Errorf("wave %d failed: %w", wave.Wave, err)
		}

		allResults = append(allResults, results...)

		// Check for failures
		for _, result := range results {
			if !result.Success {
				return allResults, fmt.Errorf("wave %d: object %s failed: %v", wave.Wave, result.ObjectName, result.Error)
			}
		}

		fmt.Printf("✅ Wave %d/%d completed\n", waveIdx+1, totalWaves)
	}

	return allResults, nil
}

// executeWave executes all objects in a wave
func (r *Runner) executeWave(wave types.Wave, completedObjects *int, totalObjects int) ([]ExecutionResult, error) {
	if wave.Parallel {
		return r.executeParallel(wave.Objects, completedObjects, totalObjects)
	}
	return r.executeSequential(wave.Objects, completedObjects, totalObjects)
}

// executeParallel executes objects in parallel with optional concurrency limit
func (r *Runner) executeParallel(objects []types.ObjectInWave, completedObjects *int, totalObjects int) ([]ExecutionResult, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]ExecutionResult, len(objects))
	errors := make([]error, len(objects))

	// Create semaphore channel for concurrency control
	var semaphore chan struct{}
	if r.concurrency > 0 {
		semaphore = make(chan struct{}, r.concurrency)
	}

	for i, obj := range objects {
		wg.Add(1)

		// Acquire semaphore if concurrency limit is set
		if semaphore != nil {
			semaphore <- struct{}{}
		}

		go func(index int, object types.ObjectInWave) {
			defer wg.Done()
			if semaphore != nil {
				defer func() { <-semaphore }() // Release semaphore
			}

			result, err := r.executeObject(object, completedObjects, totalObjects, &mu)
			results[index] = result
			errors[index] = err
		}(i, obj)
	}

	wg.Wait()

	// Check for errors
	for i, err := range errors {
		if err != nil {
			results[i].Success = false
			results[i].Error = err
		}
	}

	return results, nil
}

// executeSequential executes objects sequentially
func (r *Runner) executeSequential(objects []types.ObjectInWave, completedObjects *int, totalObjects int) ([]ExecutionResult, error) {
	results := make([]ExecutionResult, 0, len(objects))
	var mu sync.Mutex

	for _, obj := range objects {
		result, err := r.executeObject(obj, completedObjects, totalObjects, &mu)
		if err != nil {
			result.Success = false
			result.Error = err
		}
		results = append(results, result)

		if !result.Success {
			return results, fmt.Errorf("object %s failed: %v", obj.Name, err)
		}
	}

	return results, nil
}

// executeObject executes implementation for a single object
func (r *Runner) executeObject(obj types.ObjectInWave, completedObjects *int, totalObjects int, mu *sync.Mutex) (ExecutionResult, error) {
	result := ExecutionResult{
		ObjectName: obj.Name,
		Success:    false,
	}

	start := time.Now()

	// Show progress
	mu.Lock()
	*completedObjects++
	currentCount := *completedObjects
	mu.Unlock()

	fmt.Printf("\n[%d/%d] 🔨 %s\n", currentCount, totalObjects, obj.Name)
	if len(obj.Dependencies) > 0 {
		fmt.Printf("   Dependencies: %v\n", obj.Dependencies)
	}

	// Generate prompt
	fmt.Printf("   ⏳ Generating prompt...\n")
	prompt, err := r.generator.GeneratePrompt(obj)
	if err != nil {
		return result, fmt.Errorf("failed to generate prompt: %w", err)
	}

	// Save prompt to file for debugging
	promptFile := filepath.Join(r.tempDir, fmt.Sprintf("tsubo-prompt-%s.md", obj.Name))
	if err := os.WriteFile(promptFile, []byte(prompt), 0644); err != nil {
		fmt.Printf("   ⚠️  Warning: failed to save prompt file: %v\n", err)
	}

	// Execute with Claude API
	fmt.Printf("   🤖 Calling Claude API (this may take a while)...\n")
	response, err := r.client.Implement(prompt)
	if err != nil {
		result.Duration = time.Since(start)
		return result, fmt.Errorf("API call failed: %w", err)
	}

	result.Response = response
	result.Duration = time.Since(start)

	// Save response to file
	responseFile := filepath.Join(r.tempDir, fmt.Sprintf("tsubo-response-%s.md", obj.Name))
	if err := os.WriteFile(responseFile, []byte(response), 0644); err != nil {
		fmt.Printf("[%s] Warning: failed to save response file: %v\n", obj.Name, err)
	}

	// Extract files from response
	files := extractFiles(response)
	if len(files) == 0 {
		fmt.Printf("   ⚠️  Warning: no files extracted from response\n")
		fmt.Printf("   💡 Response preview (first 500 chars):\n")
		preview := response
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		fmt.Printf("   %s\n", strings.ReplaceAll(preview, "\n", "\n   "))
	} else {
		fmt.Printf("   📦 Extracted %d file(s) from response\n", len(files))

		// Save implementation to implementations directory
		serviceDir := filepath.Join(r.plan.ImplementationsDir, obj.Name)
		if err := saveImplementation(serviceDir, files); err != nil {
			result.Duration = time.Since(start)
			return result, fmt.Errorf("failed to save implementation: %w", err)
		}

		fmt.Printf("   💾 Saved to: %s\n", serviceDir)
	}

	fmt.Printf("   ⏱️  Completed in %s\n", result.Duration)
	fmt.Printf("   ✅ %s implemented successfully\n", obj.Name)

	result.Success = true
	return result, nil
}

// extractFiles extracts files from Claude's response
// Supports multiple patterns:
// - <file path="filename">```lang\ncontent\n```</file>
// - `filename`:\n```lang\ncontent\n```
// - ```lang:filename\ncontent\n```
func extractFiles(response string) map[string]string {
	files := make(map[string]string)

	// Pattern 1: <file path="filename">```lang\ncontent\n```</file>
	pattern1 := regexp.MustCompile(`<file\s+path="([^"]+)">(?:\s*)` + "```[a-z]*\\s*([\\s\\S]*?)```" + `(?:\s*)</file>`)
	matches1 := pattern1.FindAllStringSubmatch(response, -1)
	for _, match := range matches1 {
		if len(match) >= 3 {
			filename := strings.TrimSpace(match[1])
			content := strings.TrimSpace(match[2])
			files[filename] = content
		}
	}

	// Pattern 2: `filename`:\n```lang\ncontent\n```
	pattern2 := regexp.MustCompile("`([^`]+)`:\\s*```[a-z]*\\s*([\\s\\S]*?)```")
	matches2 := pattern2.FindAllStringSubmatch(response, -1)
	for _, match := range matches2 {
		if len(match) >= 3 {
			filename := strings.TrimSpace(match[1])
			content := strings.TrimSpace(match[2])
			if _, exists := files[filename]; !exists {
				files[filename] = content
			}
		}
	}

	// Pattern 3: ```lang:filename\ncontent\n```
	pattern3 := regexp.MustCompile("```[a-z]*:([^\\s]+)\\s*([\\s\\S]*?)```")
	matches3 := pattern3.FindAllStringSubmatch(response, -1)
	for _, match := range matches3 {
		if len(match) >= 3 {
			filename := strings.TrimSpace(match[1])
			content := strings.TrimSpace(match[2])
			if _, exists := files[filename]; !exists {
				files[filename] = content
			}
		}
	}

	return files
}

// saveImplementation saves extracted files to the implementations directory
func saveImplementation(serviceDir string, files map[string]string) error {
	// Create service directory
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return fmt.Errorf("failed to create service directory: %w", err)
	}

	// Save each file
	for filename, content := range files {
		filePath := filepath.Join(serviceDir, filename)

		// Create subdirectories if needed
		dir := filepath.Dir(filePath)
		if dir != serviceDir {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}

		// Write file
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filename, err)
		}
	}

	return nil
}

// PrintSummary prints a summary of execution results
func PrintSummary(results []ExecutionResult) {
	fmt.Println("\n=== Execution Summary ===")
	fmt.Printf("Total objects: %d\n", len(results))

	successful := 0
	failed := 0
	totalDuration := time.Duration(0)

	for _, result := range results {
		if result.Success {
			successful++
		} else {
			failed++
		}
		totalDuration += result.Duration
	}

	fmt.Printf("Successful: %d\n", successful)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Total duration: %s\n", totalDuration)
	fmt.Println()

	fmt.Println("Details:")
	for _, result := range results {
		status := "✓"
		if !result.Success {
			status = "✗"
		}
		fmt.Printf("  %s %s (%s)\n", status, result.ObjectName, result.Duration)
		if !result.Success {
			fmt.Printf("    Error: %v\n", result.Error)
		}
	}
}
