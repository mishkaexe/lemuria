package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mishkaexe/lemuria/pkg/helmrender"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRenderer(t *testing.T) {
	t.Run("should create new renderer successfully", func(t *testing.T) {
		renderer := helmrender.NewRenderer()
		assert.NotNil(t, renderer, "NewRenderer should return non-nil renderer")
	})
}

func TestRender_ValidChart(t *testing.T) {
	renderer := helmrender.NewRenderer()
	require.NotNil(t, renderer)

	testDataDir := getTestDataDir(t)
	chartPath := filepath.Join(testDataDir, "valid-chart")

	t.Run("should render chart with default values", func(t *testing.T) {
		opts := helmrender.RenderOptions{
			ChartPath:   chartPath,
			ReleaseName: "test-release",
			Namespace:   "default",
		}

		result, err := renderer.Render(opts)
		assert.NoError(t, err, "Render should not return error for valid chart")
		assert.NotNil(t, result, "Result should not be nil")
		assert.NotEmpty(t, result.Manifests, "Manifests should not be empty")

		var containsDeployment bool = false
		var containsReleaseName bool = false
		for _, manifest := range result.Manifests {
			if strings.Contains(manifest, "kind: Deployment") {
				containsDeployment = true
			}
			if strings.Contains(manifest, "test-release") {
				containsReleaseName = true
			}
		}

		assert.True(t, containsDeployment, "Should contain Deployment manifest")
		assert.True(t, containsReleaseName, "Should contain release name in manifest")
	})

	t.Run("should render chart with custom values file", func(t *testing.T) {
		valuesFile := filepath.Join(chartPath, "values-dev.yaml")
		opts := helmrender.RenderOptions{
			ChartPath:   chartPath,
			ValuesFiles: []string{valuesFile},
			ReleaseName: "dev-release",
			Namespace:   "development",
		}

		result, err := renderer.Render(opts)
		assert.NoError(t, err, "Render should not return error with custom values")
		assert.NotNil(t, result, "Result should not be nil")
		assert.NotEmpty(t, result.Manifests, "Manifests should not be empty")

		deployManifest := getDeploymentManifest(result.Manifests)
		// Check that dev values are applied
		assert.Contains(t, deployManifest, "replicas: 2", "Should use dev replica count")
		assert.Contains(t, deployManifest, "dev-version", "Should use dev image tag")
	})
}

func TestRender_MultipleValuesFiles(t *testing.T) {
	renderer := helmrender.NewRenderer()
	require.NotNil(t, renderer)

	testDataDir := getTestDataDir(t)
	chartPath := filepath.Join(testDataDir, "valid-chart")

	t.Run("should merge multiple values files with correct precedence", func(t *testing.T) {
		baseValues := filepath.Join(chartPath, "values.yaml")
		devValues := filepath.Join(chartPath, "values-dev.yaml")

		opts := helmrender.RenderOptions{
			ChartPath:   chartPath,
			ValuesFiles: []string{baseValues, devValues},
			ReleaseName: "multi-values-release",
			Namespace:   "default",
		}

		result, err := renderer.Render(opts)
		assert.NoError(t, err, "Render should not return error with multiple values files")
		assert.NotNil(t, result, "Result should not be nil")
		assert.NotEmpty(t, result.Manifests, "Manifests should not be empty")

		// Dev values should override base values
		deployManifest := getDeploymentManifest(result.Manifests)
		assert.Contains(t, deployManifest, "replicas: 2", "Dev values should override base values")
		assert.Contains(t, deployManifest, "dev-version", "Should use dev image tag")
	})
}

func TestRender_InlineValues(t *testing.T) {
	renderer := helmrender.NewRenderer()
	require.NotNil(t, renderer)

	testDataDir := getTestDataDir(t)
	chartPath := filepath.Join(testDataDir, "valid-chart")

	t.Run("should apply inline values", func(t *testing.T) {
		inlineValues := map[string]interface{}{
			"replicaCount": 5,
			"image": map[string]interface{}{
				"tag": "custom-tag",
			},
			"config": map[string]interface{}{
				"message":     "Hello from inline values",
				"environment": "custom",
			},
		}

		opts := helmrender.RenderOptions{
			ChartPath:   chartPath,
			Values:      inlineValues,
			ReleaseName: "inline-values-release",
			Namespace:   "default",
		}

		result, err := renderer.Render(opts)
		assert.NoError(t, err, "Render should not return error with inline values")
		assert.NotNil(t, result, "Result should not be nil")
		assert.NotEmpty(t, result.Manifests, "Manifests should not be empty")

		deployManifest := getDeploymentManifest(result.Manifests)
		assert.Contains(t, deployManifest, "replicas: 5", "Should use inline replica count")
		assert.Contains(t, deployManifest, "custom-tag", "Should use inline image tag")
	})

	t.Run("should merge inline values with values files", func(t *testing.T) {
		valuesFile := filepath.Join(chartPath, "values-dev.yaml")
		inlineValues := map[string]interface{}{
			"replicaCount": 7, // Override dev value
		}

		opts := helmrender.RenderOptions{
			ChartPath:   chartPath,
			ValuesFiles: []string{valuesFile},
			Values:      inlineValues,
			ReleaseName: "mixed-values-release",
			Namespace:   "default",
		}

		result, err := renderer.Render(opts)
		assert.NoError(t, err, "Render should not return error with mixed values")
		assert.NotNil(t, result, "Result should not be nil")
		assert.NotEmpty(t, result.Manifests, "Manifests should not be empty")

		deployManifest := getDeploymentManifest(result.Manifests)
		assert.Contains(t, deployManifest, "replicas: 7", "Inline values should override file values")
		assert.Contains(t, deployManifest, "dev-version", "Non-overridden file values should still apply")
	})
}

func TestRender_InvalidChart(t *testing.T) {
	renderer := helmrender.NewRenderer()
	require.NotNil(t, renderer)

	testDataDir := getTestDataDir(t)

	t.Run("should return error for non-existent chart", func(t *testing.T) {
		opts := helmrender.RenderOptions{
			ChartPath:   "/non/existent/path",
			ReleaseName: "test-release",
			Namespace:   "default",
		}

		result, err := renderer.Render(opts)
		assert.Error(t, err, "Should return error for non-existent chart")
		assert.Nil(t, result, "Result should be nil on error")

		// Check if it's the expected error type
		var chartNotFoundErr *helmrender.ChartNotFoundError
		assert.ErrorAs(t, err, &chartNotFoundErr, "Should return ChartNotFoundError")
	})

	t.Run("should return error for invalid values file", func(t *testing.T) {
		chartPath := filepath.Join(testDataDir, "invalid-chart")
		valuesFile := filepath.Join(chartPath, "values-invalid.yaml")

		opts := helmrender.RenderOptions{
			ChartPath:   chartPath,
			ValuesFiles: []string{valuesFile},
			ReleaseName: "invalid-release",
			Namespace:   "default",
		}

		result, err := renderer.Render(opts)
		assert.Error(t, err, "Should return error for invalid values file")
		assert.Nil(t, result, "Result should be nil on error")

		// Check if it's the expected error type
		var invalidValuesErr *helmrender.InvalidValuesError
		assert.ErrorAs(t, err, &invalidValuesErr, "Should return InvalidValuesError")
	})

	t.Run("should return error for non-existent values file", func(t *testing.T) {
		chartPath := filepath.Join(testDataDir, "valid-chart")
		nonExistentValues := filepath.Join(chartPath, "non-existent-values.yaml")

		opts := helmrender.RenderOptions{
			ChartPath:   chartPath,
			ValuesFiles: []string{nonExistentValues},
			ReleaseName: "test-release",
			Namespace:   "default",
		}

		result, err := renderer.Render(opts)
		assert.Error(t, err, "Should return error for non-existent values file")
		assert.Nil(t, result, "Result should be nil on error")
	})
}

func TestRender_EmptyChart(t *testing.T) {
	renderer := helmrender.NewRenderer()
	require.NotNil(t, renderer)

	t.Run("should handle empty release name", func(t *testing.T) {
		testDataDir := getTestDataDir(t)
		chartPath := filepath.Join(testDataDir, "valid-chart")

		opts := helmrender.RenderOptions{
			ChartPath:   chartPath,
			ReleaseName: "", // Empty release name
			Namespace:   "default",
		}

		result, err := renderer.Render(opts)
		assert.Error(t, err, "Should return error for empty release name")
		assert.Nil(t, result, "Result should be nil on error")
	})

	t.Run("should handle empty namespace", func(t *testing.T) {
		testDataDir := getTestDataDir(t)
		chartPath := filepath.Join(testDataDir, "valid-chart")

		opts := helmrender.RenderOptions{
			ChartPath:   chartPath,
			ReleaseName: "test-release",
			Namespace:   "", // Empty namespace should default to "default"
		}

		result, err := renderer.Render(opts)
		// This should either work with default namespace or return a specific error
		if err != nil {
			t.Logf("Render with empty namespace returned error: %v", err)
		} else {
			assert.NotNil(t, result, "Result should not be nil")
			assert.NotEmpty(t, result.Manifests, "Manifests should not be empty")
		}
	})
}

func TestRender_ManifestSeparation(t *testing.T) {
	renderer := helmrender.NewRenderer()
	require.NotNil(t, renderer)

	testDataDir := getTestDataDir(t)
	chartPath := filepath.Join(testDataDir, "valid-chart")

	t.Run("should separate manifests correctly", func(t *testing.T) {
		opts := helmrender.RenderOptions{
			ChartPath:   chartPath,
			ReleaseName: "separation-test",
			Namespace:   "default",
		}

		result, err := renderer.Render(opts)
		assert.NoError(t, err, "Render should not return error")
		assert.NotNil(t, result, "Result should not be nil")
		assert.NotEmpty(t, result.Manifests, "Manifests should not be empty")

		// Should have multiple manifests (deployment, service, configmap)
		assert.GreaterOrEqual(t, len(result.Manifests), 3, "Should have at least 3 manifests")

		// Check that each manifest contains only one Kind
		for i, manifest := range result.Manifests {
			kindCount := 0
			lines := []string{"kind: Deployment", "kind: Service", "kind: ConfigMap"}
			for _, line := range lines {
				if strings.Contains(manifest, line) {
					kindCount++
				}
			}
			assert.LessOrEqual(t, kindCount, 1, "Manifest %d should contain at most one Kind", i)
		}
	})
}

// Helper function to get test data directory
func getTestDataDir(tb testing.TB) string {
	wd, err := os.Getwd()
	require.NoError(tb, err, "Should be able to get working directory")
	return filepath.Join(wd, "testdata")
}

func getDeploymentManifest(manifests []string) string {
	var deployManifest string
	for _, manifest := range manifests {
		if strings.Contains(manifest, "kind: Deployment") {
			deployManifest = manifest
		}
	}

	return deployManifest
}
