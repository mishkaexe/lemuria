# Helm Render Library - Basic Example

This example demonstrates how to use the `helmrender` library to render Helm charts programmatically.

## Running the Example

From the `examples/basic` directory:

```bash
go run main.go
```

## What This Example Shows

### 1. Default Values Rendering
Renders a chart using only the default `values.yaml` file.

### 2. Custom Values File
Shows how to use a specific values file (e.g., `values-dev.yaml`) to override default values.

### 3. Inline Values
Demonstrates passing values directly as a Go map structure, useful for dynamic value generation.

### 4. Mixed Values
Combines a values file with inline values, where inline values take precedence.

### 5. Error Handling
Shows different types of errors that can occur and how to handle them:
- `ChartNotFoundError`: When the chart path doesn't exist
- `InvalidValuesError`: When a values file contains invalid YAML
- General validation errors for missing required fields

## Expected Output

The example will render manifests for a simple application including:
- Deployment
- Service  
- ConfigMap

Each example shows different configurations based on the values provided.

## Notes

- This example assumes the test charts exist in `../../test/testdata/`
- The actual implementation needs to be completed in `pkg/helmrender/` for this to work
- Error handling demonstrates the custom error types defined in the library