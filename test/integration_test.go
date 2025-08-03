package test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/mishkaexe/lemuria/pkg/helmrender"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_RealChart_DevValues(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	renderer := helmrender.NewRenderer()
	require.NotNil(t, renderer)

	testDataDir := getTestDataDir(t)
	chartPath := filepath.Join(testDataDir, "valid-chart")
	valuesFile := filepath.Join(chartPath, "values-dev.yaml")

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ValuesFiles: []string{valuesFile},
		ReleaseName: "integration-dev",
		Namespace:   "development",
	}

	result, err := renderer.Render(opts)
	require.NoError(t, err, "Integration test should render successfully")
	require.NotNil(t, result, "Result should not be nil")

	// Validate all expected manifests are present
	expectedKinds := []string{"Deployment", "Service", "ConfigMap"}
	manifestKinds := make(map[string]bool)

	for _, manifest := range result.Manifests {
		for _, kind := range expectedKinds {
			if strings.Contains(manifest, "kind: "+kind) {
				manifestKinds[kind] = true
			}
		}
	}

	for _, kind := range expectedKinds {
		assert.True(t, manifestKinds[kind], "Should contain %s manifest", kind)
	}

	// Validate specific dev environment values
	deploymentManifest := findManifestByKind(result.Manifests, "Deployment")
	require.NotEmpty(t, deploymentManifest, "Should have Deployment manifest")

	assert.Contains(t, deploymentManifest, "replicas: 2", "Should use dev replica count")
	assert.Contains(t, deploymentManifest, "development", "Should contain development environment")
	assert.Contains(t, deploymentManifest, "Hello from development", "Should contain dev message")

	serviceManifest := findManifestByKind(result.Manifests, "Service")
	require.NotEmpty(t, serviceManifest, "Should have Service manifest")
	assert.Contains(t, serviceManifest, "type: NodePort", "Should use NodePort service type")
}

func TestIntegration_RealChart_ProdValues(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	renderer := helmrender.NewRenderer()
	require.NotNil(t, renderer)

	testDataDir := getTestDataDir(t)
	chartPath := filepath.Join(testDataDir, "valid-chart")
	valuesFile := filepath.Join(chartPath, "values-prod.yaml")

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ValuesFiles: []string{valuesFile},
		ReleaseName: "integration-prod",
		Namespace:   "production",
	}

	result, err := renderer.Render(opts)
	require.NoError(t, err, "Integration test should render successfully")
	require.NotNil(t, result, "Result should not be nil")

	// Validate production-specific values
	deploymentManifest := findManifestByKind(result.Manifests, "Deployment")
	require.NotEmpty(t, deploymentManifest, "Should have Deployment manifest")

	assert.Contains(t, deploymentManifest, "replicas: 3", "Should use prod replica count")
	assert.Contains(t, deploymentManifest, "production", "Should contain production environment")
	assert.Contains(t, deploymentManifest, "Hello from production", "Should contain prod message")

	serviceManifest := findManifestByKind(result.Manifests, "Service")
	require.NotEmpty(t, serviceManifest, "Should have Service manifest")
	assert.Contains(t, serviceManifest, "type: LoadBalancer", "Should use LoadBalancer service type")
}

func TestIntegration_ComplexChart_HugeValues(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	renderer := helmrender.NewRenderer()
	require.NotNil(t, renderer)

	testDataDir := getTestDataDir(t)
	chartPath := filepath.Join(testDataDir, "complex-chart")
	valuesFile := filepath.Join(chartPath, "values-huge.yaml")

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ValuesFiles: []string{valuesFile},
		ReleaseName: "integration-complex",
		Namespace:   "production",
	}

	result, err := renderer.Render(opts)
	require.NoError(t, err, "Complex chart should render successfully")
	require.NotNil(t, result, "Result should not be nil")

	// Should have multiple manifests including HPA
	assert.GreaterOrEqual(t, len(result.Manifests), 3, "Complex chart should have multiple manifests")

	// Check for HPA manifest (autoscaling is enabled)
	hpaManifest := findManifestByKind(result.Manifests, "HorizontalPodAutoscaler")
	assert.NotEmpty(t, hpaManifest, "Should have HPA manifest when autoscaling is enabled")
	assert.Contains(t, hpaManifest, "minReplicas: 5", "Should use correct min replicas")
	assert.Contains(t, hpaManifest, "maxReplicas: 100", "Should use correct max replicas")

	// Check deployment with complex values
	deploymentManifest := findManifestByKind(result.Manifests, "Deployment")
	require.NotEmpty(t, deploymentManifest, "Should have Deployment manifest")
	assert.Contains(t, deploymentManifest, "production", "Should contain production environment")
}

func TestIntegration_ComplexChart_MinimalValues(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	renderer := helmrender.NewRenderer()
	require.NotNil(t, renderer)

	testDataDir := getTestDataDir(t)
	chartPath := filepath.Join(testDataDir, "complex-chart")
	valuesFile := filepath.Join(chartPath, "values-minimal.yaml")

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ValuesFiles: []string{valuesFile},
		ReleaseName: "integration-minimal",
		Namespace:   "default",
	}

	result, err := renderer.Render(opts)
	require.NoError(t, err, "Complex chart with minimal values should render successfully")
	require.NotNil(t, result, "Result should not be nil")

	// Should have basic manifests but no HPA (autoscaling not enabled)
	hpaManifest := findManifestByKind(result.Manifests, "HorizontalPodAutoscaler")
	assert.Empty(t, hpaManifest, "Should not have HPA manifest when autoscaling is disabled")

	deploymentManifest := findManifestByKind(result.Manifests, "Deployment")
	require.NotEmpty(t, deploymentManifest, "Should have Deployment manifest")
	assert.Contains(t, deploymentManifest, "replicas: 1", "Should use minimal replica count")
}

func TestIntegration_ValuesOverride(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	renderer := helmrender.NewRenderer()
	require.NotNil(t, renderer)

	testDataDir := getTestDataDir(t)
	chartPath := filepath.Join(testDataDir, "valid-chart")
	devValuesFile := filepath.Join(chartPath, "values-dev.yaml")

	// Override dev values with inline values
	inlineValues := map[string]interface{}{
		"replicaCount": 10, // Override dev's 2 replicas
		"service": map[string]interface{}{
			"type": "LoadBalancer", // Override dev's NodePort
		},
		"config": map[string]interface{}{
			"message": "Overridden by integration test", // Override dev message
		},
	}

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ValuesFiles: []string{devValuesFile},
		Values:      inlineValues,
		ReleaseName: "integration-override",
		Namespace:   "testing",
	}

	result, err := renderer.Render(opts)
	require.NoError(t, err, "Values override should work successfully")
	require.NotNil(t, result, "Result should not be nil")

	deploymentManifest := findManifestByKind(result.Manifests, "Deployment")
	require.NotEmpty(t, deploymentManifest, "Should have Deployment manifest")

	// Inline values should override file values
	assert.Contains(t, deploymentManifest, "replicas: 10", "Inline values should override file values")
	assert.Contains(t, deploymentManifest, "Overridden by integration test", "Inline message should override file message")

	serviceManifest := findManifestByKind(result.Manifests, "Service")
	require.NotEmpty(t, serviceManifest, "Should have Service manifest")
	assert.Contains(t, serviceManifest, "type: LoadBalancer", "Inline service type should override file service type")

	// Non-overridden values from dev file should still be present
	assert.Contains(t, deploymentManifest, "dev-version", "Non-overridden dev image tag should be preserved")
}

func TestIntegration_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	renderer := helmrender.NewRenderer()
	require.NotNil(t, renderer)

	testDataDir := getTestDataDir(t)

	t.Run("should handle chart with template errors gracefully", func(t *testing.T) {
		chartPath := filepath.Join(testDataDir, "invalid-chart")

		opts := helmrender.RenderOptions{
			ChartPath:   chartPath,
			ReleaseName: "integration-error",
			Namespace:   "default",
		}

		result, err := renderer.Render(opts)
		assert.Error(t, err, "Should return error for chart with template errors")
		assert.Nil(t, result, "Result should be nil on error")

		// Should be a render error
		var renderErr *helmrender.RenderError
		assert.ErrorAs(t, err, &renderErr, "Should return RenderError for template errors")
	})

	t.Run("should handle invalid values file gracefully", func(t *testing.T) {
		chartPath := filepath.Join(testDataDir, "valid-chart")
		invalidValuesFile := filepath.Join(testDataDir, "invalid-chart", "values-invalid.yaml")

		opts := helmrender.RenderOptions{
			ChartPath:   chartPath,
			ValuesFiles: []string{invalidValuesFile},
			ReleaseName: "integration-invalid-values",
			Namespace:   "default",
		}

		result, err := renderer.Render(opts)
		assert.Error(t, err, "Should return error for invalid values file")
		assert.Nil(t, result, "Result should be nil on error")

		// Should be an invalid values error
		var invalidValuesErr *helmrender.InvalidValuesError
		assert.ErrorAs(t, err, &invalidValuesErr, "Should return InvalidValuesError")
	})
}

func TestIntegration_ManifestStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	renderer := helmrender.NewRenderer()
	require.NotNil(t, renderer)

	testDataDir := getTestDataDir(t)
	chartPath := filepath.Join(testDataDir, "valid-chart")

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ReleaseName: "integration-structure",
		Namespace:   "default",
	}

	result, err := renderer.Render(opts)
	require.NoError(t, err, "Should render successfully")
	require.NotNil(t, result, "Result should not be nil")

	t.Run("each manifest should be valid YAML", func(t *testing.T) {
		for i, manifest := range result.Manifests {
			assert.NotEmpty(t, strings.TrimSpace(manifest), "Manifest %d should not be empty", i)
			assert.Contains(t, manifest, "apiVersion:", "Manifest %d should contain apiVersion", i)
			assert.Contains(t, manifest, "kind:", "Manifest %d should contain kind", i)
			assert.Contains(t, manifest, "metadata:", "Manifest %d should contain metadata", i)
		}
	})

	t.Run("manifests should contain correct labels", func(t *testing.T) {
		for i, manifest := range result.Manifests {
			assert.Contains(t, manifest, "integration-structure", "Manifest %d should contain release name", i)
			assert.Contains(t, manifest, "app.kubernetes.io/name:", "Manifest %d should contain app name label", i)
			assert.Contains(t, manifest, "app.kubernetes.io/instance:", "Manifest %d should contain instance label", i)
		}
	})
}

// Helper function to find a manifest by kind
func findManifestByKind(manifests []string, kind string) string {
	for _, manifest := range manifests {
		if strings.Contains(manifest, "kind: "+kind) {
			return manifest
		}
	}
	return ""
}
