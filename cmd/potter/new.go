package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func runNew(args []string) error {
	fs := flag.NewFlagSet("new", flag.ExitOnError)
	helpFlag := fs.Bool("help", false, "Show help for new command")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *helpFlag {
		printNewUsage()
		return nil
	}

	// Get service name (default: "example")
	args = fs.Args()
	serviceName := "example"
	if len(args) > 0 {
		serviceName = args[0]
	}

	return createServiceTemplate(serviceName)
}

func createServiceTemplate(serviceName string) error {
	fmt.Printf("Creating service template: %s\n", serviceName)

	// Create contracts directory if it doesn't exist
	contractsDir := "contracts"
	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		return fmt.Errorf("failed to create contracts directory: %w", err)
	}

	// Generate .object.yaml template
	objectFile := filepath.Join(contractsDir, fmt.Sprintf("%s.object.yaml", serviceName))

	template := generateObjectTemplate(serviceName)

	if err := os.WriteFile(objectFile, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write object file: %w", err)
	}

	fmt.Printf("\n%s✓ Service template created:%s\n", colorGreen, colorReset)
	fmt.Printf("  %s\n", objectFile)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit the contract to define your service")
	fmt.Println("  2. Add it to your .tsubo.yaml file")
	fmt.Printf("  3. Run: potter build <tsubo-file> --ai\n")
	fmt.Println()

	return nil
}

func generateObjectTemplate(serviceName string) string {
	return fmt.Sprintf(`version: "1.0"

service:
  name: %s
  description: Description of %s service
  runtime:
    language: go
    version: "1.22"

api:
  type: rest
  port: 8080
  endpoints:
    - path: /health
      method: GET
      description: Health check endpoint
      response:
        type: object
        properties:
          status:
            type: string
            example: "healthy"

    - path: /%s
      method: POST
      description: Create a new %s
      request:
        type: object
        properties:
          name:
            type: string
            required: true
      response:
        type: object
        properties:
          id:
            type: string
          name:
            type: string

    - path: /%s/{id}
      method: GET
      description: Get %s by ID
      response:
        type: object
        properties:
          id:
            type: string
          name:
            type: string

dependencies:
  services: []
    # Example:
    # - name: other-service
    #   endpoint: http://other-service:8080

  databases:
    - type: postgres
      name: %s_db
      schema:
        tables:
          - name: %ss
            columns:
              - name: id
                type: uuid
                primary_key: true
              - name: name
                type: varchar(255)
                required: true
              - name: created_at
                type: timestamp
                default: now()
`, serviceName, serviceName, serviceName, serviceName, serviceName, serviceName, serviceName, serviceName)
}

func printNewUsage() {
	fmt.Println("Usage: potter new [service-name]")
	fmt.Println()
	fmt.Println("Creates a new service contract template (.object.yaml)")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  service-name    Name of the service (default: example)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  potter new              # Creates example.object.yaml")
	fmt.Println("  potter new user         # Creates user.object.yaml")
	fmt.Println("  potter new todo         # Creates todo.object.yaml")
}
