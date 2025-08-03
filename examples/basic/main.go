package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/mishkaexe/lemuria/pkg/helmrender"
)

func main() {
	// Create a new Helm chart renderer
	renderer := helmrender.NewRenderer()
	if renderer == nil {
		log.Fatal("Failed to create renderer")
	}

	// Example 1: Render chart with default values
	fmt.Println("=== Example 1: Default Values ===")
	renderWithDefaults(renderer)

	// Example 2: Render chart with custom values file
	fmt.Println("\n=== Example 2: Custom Values File ===")
	renderWithValuesFile(renderer)

	// Example 3: Render chart with inline values
	fmt.Println("\n=== Example 3: Inline Values ===")
	renderWithInlineValues(renderer)

	// Example 4: Render chart with mixed values (file + inline)
	fmt.Println("\n=== Example 4: Mixed Values ===")
	renderWithMixedValues(renderer)

	// Example 5: Error handling
	fmt.Println("\n=== Example 5: Error Handling ===")
	demonstrateErrorHandling(renderer)
}

func renderWithDefaults(renderer *helmrender.ChartRenderer) {
	// Assuming test chart exists in ../../test/testdata/valid-chart
	chartPath := "../../test/testdata/valid-chart"

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ReleaseName: "example-default",
		Namespace:   "default",
	}

	result, err := renderer.Render(opts)
	if err != nil {
		fmt.Printf("Error rendering chart: %v\n", err)
		return
	}

	fmt.Printf("Rendered %d manifests:\n", len(result.Manifests))
	for i, manifest := range result.Manifests {
		fmt.Printf("--- Manifest %d ---\n", i+1)
		fmt.Println(manifest)
	}

	if result.Notes != "" {
		fmt.Printf("Notes:\n%s\n", result.Notes)
	}
}

func renderWithValuesFile(renderer *helmrender.ChartRenderer) {
	// Render with development values
	chartPath := "../../test/testdata/valid-chart"
	valuesFile := filepath.Join(chartPath, "values-dev.yaml")

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ValuesFiles: []string{valuesFile},
		ReleaseName: "example-dev",
		Namespace:   "development",
	}

	result, err := renderer.Render(opts)
	if err != nil {
		fmt.Printf("Error rendering chart with values file: %v\n", err)
		return
	}

	fmt.Printf("Rendered chart with dev values - %d manifests generated\n", len(result.Manifests))
	
	// Show just the first manifest as an example
	if len(result.Manifests) > 0 {
		fmt.Println("First manifest:")
		fmt.Println(result.Manifests[0])
	}
}

func renderWithInlineValues(renderer *helmrender.ChartRenderer) {
	chartPath := "../../test/testdata/valid-chart"

	// Define custom values inline
	inlineValues := map[string]interface{}{
		"replicaCount": 5,
		"image": map[string]interface{}{
			"repository": "nginx",
			"tag":        "custom-tag",
			"pullPolicy": "Always",
		},
		"service": map[string]interface{}{
			"type": "LoadBalancer",
			"port": 8080,
		},
		"config": map[string]interface{}{
			"message":     "Hello from inline values!",
			"environment": "custom",
			"debug":       true,
		},
	}

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		Values:      inlineValues,
		ReleaseName: "example-inline",
		Namespace:   "custom",
	}

	result, err := renderer.Render(opts)
	if err != nil {
		fmt.Printf("Error rendering chart with inline values: %v\n", err)
		return
	}

	fmt.Printf("Rendered chart with inline values - %d manifests generated\n", len(result.Manifests))
	
	// Show deployment manifest to see custom values
	for _, manifest := range result.Manifests {
		if contains(manifest, "kind: Deployment") {
			fmt.Println("Deployment manifest with custom values:")
			fmt.Println(manifest)
			break
		}
	}
}

func renderWithMixedValues(renderer *helmrender.ChartRenderer) {
	chartPath := "../../test/testdata/valid-chart"
	valuesFile := filepath.Join(chartPath, "values-dev.yaml")

	// Start with dev values, then override some with inline values
	inlineValues := map[string]interface{}{
		"replicaCount": 8, // Override dev's 2 replicas
		"config": map[string]interface{}{
			"message": "Mixed values example", // Override dev message
		},
		// service.type will remain as NodePort from dev values
	}

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ValuesFiles: []string{valuesFile},
		Values:      inlineValues,
		ReleaseName: "example-mixed",
		Namespace:   "mixed",
	}

	result, err := renderer.Render(opts)
	if err != nil {
		fmt.Printf("Error rendering chart with mixed values: %v\n", err)
		return
	}

	fmt.Printf("Rendered chart with mixed values - %d manifests generated\n", len(result.Manifests))
	fmt.Println("This combines dev values file with inline overrides")
}

func demonstrateErrorHandling(renderer *helmrender.ChartRenderer) {
	// Example 1: Non-existent chart
	fmt.Println("1. Non-existent chart:")
	opts1 := helmrender.RenderOptions{
		ChartPath:   "/non/existent/path",
		ReleaseName: "error-example",
		Namespace:   "default",
	}

	_, err := renderer.Render(opts1)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
		
		// Check error type
		if chartErr, ok := err.(*helmrender.ChartNotFoundError); ok {
			fmt.Printf("Chart not found at path: %s\n", chartErr.Path)
		}
	}

	// Example 2: Invalid values file
	fmt.Println("\n2. Invalid values file:")
	chartPath := "../../test/testdata/valid-chart"
	invalidValues := "../../test/testdata/invalid-chart/values-invalid.yaml"
	
	opts2 := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ValuesFiles: []string{invalidValues},
		ReleaseName: "error-example-2",
		Namespace:   "default",
	}

	_, err = renderer.Render(opts2)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
		
		// Check error type
		if valuesErr, ok := err.(*helmrender.InvalidValuesError); ok {
			fmt.Printf("Invalid values file: %s\n", valuesErr.File)
		}
	}

	// Example 3: Empty release name
	fmt.Println("\n3. Empty release name:")
	opts3 := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ReleaseName: "", // Empty release name
		Namespace:   "default",
	}

	_, err = renderer.Render(opts3)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}