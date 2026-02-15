package executor

import (
	"fmt"
	"os"
	"path/filepath"
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
	concurrency int // 0 = unlimited
}

// NewRunner creates a new execution runner
func NewRunner(plan *types.ImplementationPlan) (*Runner, error) {
	client, err := NewClaudeClient()
	if err != nil {
		return nil, err
	}

	return &Runner{
		client:      client,
		generator:   NewPromptGenerator(plan),
		plan:        plan,
		concurrency: 0, // unlimited by default
	}, nil
}

// SetConcurrency sets the maximum number of parallel executions
func (r *Runner) SetConcurrency(n int) {
	r.concurrency = n
}

// ExecuteAll executes all waves in the implementation plan
func (r *Runner) ExecuteAll() ([]ExecutionResult, error) {
	var allResults []ExecutionResult

	for _, wave := range r.plan.Waves {
		fmt.Printf("\n=== Executing Wave %d ===\n", wave.Wave)
		fmt.Printf("Objects: %d\n", len(wave.Objects))
		fmt.Printf("Mode: ")
		if wave.Parallel {
			fmt.Println("Parallel")
		} else {
			fmt.Println("Sequential")
		}
		fmt.Println()

		results, err := r.executeWave(wave)
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

		fmt.Printf("Wave %d completed successfully\n", wave.Wave)
	}

	return allResults, nil
}

// executeWave executes all objects in a wave
func (r *Runner) executeWave(wave types.Wave) ([]ExecutionResult, error) {
	if wave.Parallel {
		return r.executeParallel(wave.Objects)
	}
	return r.executeSequential(wave.Objects)
}

// executeParallel executes objects in parallel with optional concurrency limit
func (r *Runner) executeParallel(objects []types.ObjectInWave) ([]ExecutionResult, error) {
	var wg sync.WaitGroup
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

			result, err := r.executeObject(object)
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
func (r *Runner) executeSequential(objects []types.ObjectInWave) ([]ExecutionResult, error) {
	results := make([]ExecutionResult, 0, len(objects))

	for _, obj := range objects {
		result, err := r.executeObject(obj)
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
func (r *Runner) executeObject(obj types.ObjectInWave) (ExecutionResult, error) {
	result := ExecutionResult{
		ObjectName: obj.Name,
		Success:    false,
	}

	start := time.Now()

	fmt.Printf("[%s] Starting implementation...\n", obj.Name)

	// Generate prompt
	prompt, err := r.generator.GeneratePrompt(obj)
	if err != nil {
		return result, fmt.Errorf("failed to generate prompt: %w", err)
	}

	// Save prompt to file for debugging
	promptFile := filepath.Join("/tmp", fmt.Sprintf("tsubo-prompt-%s.md", obj.Name))
	if err := os.WriteFile(promptFile, []byte(prompt), 0644); err != nil {
		fmt.Printf("[%s] Warning: failed to save prompt file: %v\n", obj.Name, err)
	}

	// Execute with Claude API
	fmt.Printf("[%s] Calling Claude API...\n", obj.Name)
	response, err := r.client.Implement(prompt)
	if err != nil {
		result.Duration = time.Since(start)
		return result, fmt.Errorf("API call failed: %w", err)
	}

	result.Response = response
	result.Duration = time.Since(start)

	// Save response to file
	responseFile := filepath.Join("/tmp", fmt.Sprintf("tsubo-response-%s.md", obj.Name))
	if err := os.WriteFile(responseFile, []byte(response), 0644); err != nil {
		fmt.Printf("[%s] Warning: failed to save response file: %v\n", obj.Name, err)
	}

	fmt.Printf("[%s] Implementation completed in %s\n", obj.Name, result.Duration)
	fmt.Printf("[%s] Response saved to: %s\n", obj.Name, responseFile)

	result.Success = true
	return result, nil
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
