package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/staka121/potter/pkg/types"
)

// PromptGenerator generates implementation prompts for AI agents
type PromptGenerator struct {
	plan *types.ImplementationPlan
}

// NewPromptGenerator creates a new prompt generator
func NewPromptGenerator(plan *types.ImplementationPlan) *PromptGenerator {
	return &PromptGenerator{plan: plan}
}

// GeneratePrompt generates a complete implementation prompt for an object
func (pg *PromptGenerator) GeneratePrompt(obj types.ObjectInWave) (string, error) {
	var prompt strings.Builder

	// Header
	prompt.WriteString(fmt.Sprintf("# Implementation Task: %s\n\n", obj.Name))
	prompt.WriteString("You are implementing a microservice based on Tsubo philosophy and contracts.\n\n")

	// Critical instructions
	prompt.WriteString("**CRITICAL: Follow these steps in order:**\n\n")

	// Step 1: Read context files
	prompt.WriteString("## Step 1: Read and understand context files\n\n")
	prompt.WriteString("Read these files to understand Tsubo philosophy and principles:\n\n")

	for i, contextFile := range pg.plan.ContextFiles {
		content, err := readFileContent(contextFile)
		if err != nil {
			// If file doesn't exist, just list it
			prompt.WriteString(fmt.Sprintf("%d. %s\n", i+1, contextFile))
			continue
		}

		baseName := filepath.Base(contextFile)
		prompt.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, baseName))
		prompt.WriteString("```\n")
		prompt.WriteString(content)
		prompt.WriteString("\n```\n\n")
	}

	// Step 2: Read the contract
	prompt.WriteString("## Step 2: Read the contract\n\n")
	prompt.WriteString(fmt.Sprintf("Read the complete contract specification for %s:\n\n", obj.Name))

	contractContent, err := readFileContent(obj.Contract)
	if err != nil {
		return "", fmt.Errorf("failed to read contract %s: %w", obj.Contract, err)
	}

	prompt.WriteString(fmt.Sprintf("**Contract: %s**\n", filepath.Base(obj.Contract)))
	prompt.WriteString("```yaml\n")
	prompt.WriteString(contractContent)
	prompt.WriteString("\n```\n\n")

	// Step 3: Understand dependencies
	if len(obj.Dependencies) > 0 {
		prompt.WriteString("## Step 3: Understand dependencies\n\n")
		prompt.WriteString(fmt.Sprintf("This service depends on: %v\n\n", obj.Dependencies))
		prompt.WriteString("**IMPORTANT: Use correct service URLs for dependencies:**\n")
		prompt.WriteString("- Service URLs are based on container names and their assigned ports\n")
		prompt.WriteString("- Example: If user-service uses port 8084, the URL is http://user-service:8084\n")
		prompt.WriteString("- Check the tsubo.yaml for correct port assignments\n")
		prompt.WriteString("- DO NOT assume default ports like 8080\n\n")
		prompt.WriteString("Make sure to:\n")
		prompt.WriteString("- Implement service-to-service communication\n")
		prompt.WriteString("- Handle dependency failures gracefully\n")
		prompt.WriteString("- Use proper service discovery (environment variables or Docker network)\n\n")
	}

	// Step 4: Implementation task
	stepNum := 3
	if len(obj.Dependencies) > 0 {
		stepNum = 4
	}

	prompt.WriteString(fmt.Sprintf("## Step %d: Implement the service\n\n", stepNum))
	prompt.WriteString("**Your task:**\n")
	prompt.WriteString("- Implement " + obj.Name + " in Go language (Go 1.22) following the contract exactly\n")
	prompt.WriteString("- Create all necessary files in: " + filepath.Join(pg.plan.ImplementationsDir, obj.Name) + "\n")
	prompt.WriteString(fmt.Sprintf("- **IMPORTANT: Use port %d for this service** (defined in tsubo.yaml)\n", obj.Port))
	prompt.WriteString("- Required files:\n")
	prompt.WriteString("  - Go source files (main.go, handlers, models, storage, etc.)\n")
	prompt.WriteString("  - go.mod\n")
	prompt.WriteString("  - Dockerfile (multi-stage build with golang:1.22-alpine)\n")
	prompt.WriteString("  - docker-compose.yml\n")
	prompt.WriteString("  - .dockerignore\n")
	prompt.WriteString("  - README.md (brief implementation notes)\n")
	prompt.WriteString("  - test script (test.sh or similar)\n\n")

	prompt.WriteString("**Important principles:**\n")
	prompt.WriteString("- Follow Docker First: Everything runs in Docker\n")
	prompt.WriteString("- Do NOT ask questions during implementation (contract is complete)\n")
	prompt.WriteString("- Implement exactly what the contract specifies - no more, no less\n")
	prompt.WriteString("- Use in-memory storage as specified in the contract\n")
	prompt.WriteString("- All API endpoints must match the contract specification\n")
	prompt.WriteString("- Handle all edge cases specified in the contract\n")
	prompt.WriteString("- Use UUIDv4 for IDs\n")
	prompt.WriteString("- Follow Go best practices (standard library, simple code)\n")
	prompt.WriteString("- **CRITICAL**: Do NOT import unused packages (Go will fail to compile)\n")
	prompt.WriteString("- Only import packages that are actually used in the code\n\n")
	prompt.WriteString("**Port configuration:**\n")
	prompt.WriteString(fmt.Sprintf("- The service MUST listen on port %d\n", obj.Port))
	prompt.WriteString(fmt.Sprintf("- In Dockerfile, use EXPOSE %d\n", obj.Port))
	prompt.WriteString(fmt.Sprintf("- In docker-compose.yml, map port %d:%d\n", obj.Port, obj.Port))
	prompt.WriteString("- This port is allocated to avoid conflicts with other services\n\n")
	prompt.WriteString("**Docker network configuration:**\n")
	prompt.WriteString("- Use network name: tsubo-network\n")
	prompt.WriteString("- In docker-compose.yml, declare the network as external:\n")
	prompt.WriteString("  ```yaml\n")
	prompt.WriteString("  networks:\n")
	prompt.WriteString("    tsubo-network:\n")
	prompt.WriteString("      external: true\n")
	prompt.WriteString("  ```\n")
	prompt.WriteString("- This allows all services to communicate via the shared network\n\n")
	prompt.WriteString("**Docker Compose format:**\n")
	prompt.WriteString("- DO NOT include 'version' field in docker-compose.yml (it's obsolete)\n")
	prompt.WriteString("- Start directly with 'services:' at the top level\n\n")

	serviceDir := filepath.Join(pg.plan.ImplementationsDir, obj.Name)
	prompt.WriteString(fmt.Sprintf("**Output directory:** %s\n\n", serviceDir))

	// Output format instructions
	prompt.WriteString("## Output Format\n\n")
	prompt.WriteString("**CRITICAL:** You MUST output each file using the following exact format:\n\n")
	prompt.WriteString("```\n")
	prompt.WriteString("<create_file>\n")
	prompt.WriteString("<path>relative/path/to/file.go</path>\n")
	prompt.WriteString("<content>\n")
	prompt.WriteString("// File content here\n")
	prompt.WriteString("</content>\n")
	prompt.WriteString("</create_file>\n")
	prompt.WriteString("```\n\n")
	prompt.WriteString("**Important notes about file paths:**\n")
	prompt.WriteString(fmt.Sprintf("- All paths should be relative to the service directory\n"))
	prompt.WriteString("- Example: `main.go` (for top-level files)\n")
	prompt.WriteString("- Example: `handlers/user.go` (for nested files)\n")
	prompt.WriteString("- DO NOT include the full path like `poc/implementations/user-service/main.go`\n\n")
	prompt.WriteString("Start implementation now.\n")

	return prompt.String(), nil
}

// GenerateAllPrompts generates prompts for all objects in the plan
func (pg *PromptGenerator) GenerateAllPrompts() (map[string]string, error) {
	prompts := make(map[string]string)

	for _, wave := range pg.plan.Waves {
		for _, obj := range wave.Objects {
			prompt, err := pg.GeneratePrompt(obj)
			if err != nil {
				return nil, fmt.Errorf("failed to generate prompt for %s: %w", obj.Name, err)
			}
			prompts[obj.Name] = prompt
		}
	}

	return prompts, nil
}

// readFileContent reads and returns file content
func readFileContent(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
