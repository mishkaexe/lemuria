package test

import (
	"path/filepath"
	"testing"

	"github.com/mishkaexe/lemuria/pkg/helmrender"
	"github.com/stretchr/testify/require"
)

func BenchmarkRender_SmallChart(b *testing.B) {
	renderer := helmrender.NewRenderer()
	require.NotNil(b, renderer)

	testDataDir := getTestDataDir(b)
	chartPath := filepath.Join(testDataDir, "valid-chart")

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ReleaseName: "benchmark-small",
		Namespace:   "default",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := renderer.Render(opts)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
		if result == nil || len(result.Manifests) == 0 {
			b.Fatal("Expected non-empty result")
		}
	}
}

func BenchmarkRender_SmallChartWithValues(b *testing.B) {
	renderer := helmrender.NewRenderer()
	require.NotNil(b, renderer)

	testDataDir := getTestDataDir(b)
	chartPath := filepath.Join(testDataDir, "valid-chart")
	valuesFile := filepath.Join(chartPath, "values-dev.yaml")

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ValuesFiles: []string{valuesFile},
		ReleaseName: "benchmark-small-values",
		Namespace:   "default",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := renderer.Render(opts)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
		if result == nil || len(result.Manifests) == 0 {
			b.Fatal("Expected non-empty result")
		}
	}
}

func BenchmarkRender_LargeChart(b *testing.B) {
	renderer := helmrender.NewRenderer()
	require.NotNil(b, renderer)

	testDataDir := getTestDataDir(b)
	chartPath := filepath.Join(testDataDir, "complex-chart")
	valuesFile := filepath.Join(chartPath, "values-huge.yaml")

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ValuesFiles: []string{valuesFile},
		ReleaseName: "benchmark-large",
		Namespace:   "production",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := renderer.Render(opts)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
		if result == nil || len(result.Manifests) == 0 {
			b.Fatal("Expected non-empty result")
		}
	}
}

func BenchmarkRender_MultipleValues(b *testing.B) {
	renderer := helmrender.NewRenderer()
	require.NotNil(b, renderer)

	testDataDir := getTestDataDir(b)
	chartPath := filepath.Join(testDataDir, "valid-chart")
	baseValues := filepath.Join(chartPath, "values.yaml")
	devValues := filepath.Join(chartPath, "values-dev.yaml")
	prodValues := filepath.Join(chartPath, "values-prod.yaml")

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ValuesFiles: []string{baseValues, devValues, prodValues},
		ReleaseName: "benchmark-multiple-values",
		Namespace:   "default",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := renderer.Render(opts)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
		if result == nil || len(result.Manifests) == 0 {
			b.Fatal("Expected non-empty result")
		}
	}
}

func BenchmarkRender_InlineValues(b *testing.B) {
	renderer := helmrender.NewRenderer()
	require.NotNil(b, renderer)

	testDataDir := getTestDataDir(b)
	chartPath := filepath.Join(testDataDir, "valid-chart")

	inlineValues := map[string]interface{}{
		"replicaCount": 5,
		"image": map[string]interface{}{
			"repository": "nginx",
			"tag":        "benchmark-tag",
			"pullPolicy": "Always",
		},
		"service": map[string]interface{}{
			"type": "LoadBalancer",
			"port": 8080,
		},
		"config": map[string]interface{}{
			"message":     "Benchmark test message",
			"environment": "benchmark",
			"debug":       true,
		},
		"resources": map[string]interface{}{
			"limits": map[string]interface{}{
				"cpu":    "1000m",
				"memory": "1Gi",
			},
			"requests": map[string]interface{}{
				"cpu":    "500m",
				"memory": "512Mi",
			},
		},
	}

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		Values:      inlineValues,
		ReleaseName: "benchmark-inline",
		Namespace:   "default",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := renderer.Render(opts)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
		if result == nil || len(result.Manifests) == 0 {
			b.Fatal("Expected non-empty result")
		}
	}
}

func BenchmarkRender_MixedValues(b *testing.B) {
	renderer := helmrender.NewRenderer()
	require.NotNil(b, renderer)

	testDataDir := getTestDataDir(b)
	chartPath := filepath.Join(testDataDir, "complex-chart")
	valuesFile := filepath.Join(chartPath, "values-huge.yaml")

	inlineValues := map[string]interface{}{
		"replicaCount": 15, // Override huge values
		"config": map[string]interface{}{
			"environment": "benchmark-mixed",
			"logLevel":    "debug",
		},
		"resources": map[string]interface{}{
			"limits": map[string]interface{}{
				"cpu":    "3000m",
				"memory": "6Gi",
			},
		},
	}

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ValuesFiles: []string{valuesFile},
		Values:      inlineValues,
		ReleaseName: "benchmark-mixed",
		Namespace:   "benchmark",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := renderer.Render(opts)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
		if result == nil || len(result.Manifests) == 0 {
			b.Fatal("Expected non-empty result")
		}
	}
}

func BenchmarkRenderer_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer := helmrender.NewRenderer()
		if renderer == nil {
			b.Fatal("Expected non-nil renderer")
		}
	}
}

func BenchmarkRender_ParallelSmallChart(b *testing.B) {
	renderer := helmrender.NewRenderer()
	require.NotNil(b, renderer)

	testDataDir := getTestDataDir(b)
	chartPath := filepath.Join(testDataDir, "valid-chart")

	opts := helmrender.RenderOptions{
		ChartPath:   chartPath,
		ReleaseName: "benchmark-parallel",
		Namespace:   "default",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result, err := renderer.Render(opts)
			if err != nil {
				b.Fatalf("Render failed: %v", err)
			}
			if result == nil || len(result.Manifests) == 0 {
				b.Fatal("Expected non-empty result")
			}
		}
	})
}

