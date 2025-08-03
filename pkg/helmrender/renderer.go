package helmrender

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

// ChartRenderer handles rendering of Helm charts
type ChartRenderer struct {
	actionConfig *action.Configuration
	settings     *cli.EnvSettings
}

// RenderOptions contains options for rendering a Helm chart
type RenderOptions struct {
	ChartPath   string
	ValuesFiles []string
	Values      map[string]interface{}
	Namespace   string
	ReleaseName string
}

// RenderResult contains the result of rendering a Helm chart
type RenderResult struct {
	Manifests []string
	Notes     string
}

// NewRenderer creates a new ChartRenderer
func NewRenderer() *ChartRenderer {
	settings := cli.New()
	actionConfig := new(action.Configuration)

	// Initialize with in-memory storage to avoid external dependencies
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), "memory", func(format string, v ...interface{}) {}); err != nil {
		// For this implementation, we'll continue even if init fails
		// as we're only using template rendering which doesn't require full cluster access
	}

	return &ChartRenderer{
		actionConfig: actionConfig,
		settings:     settings,
	}
}

// Render renders a Helm chart with the given options
func (r *ChartRenderer) Render(opts RenderOptions) (*RenderResult, error) {
	// Input validation
	if err := r.validateOptions(opts); err != nil {
		return nil, err
	}

	// Load chart from filesystem
	chart, err := r.loadChart(opts.ChartPath)
	if err != nil {
		return nil, err
	}

	// Parse and merge values
	values, err := r.mergeValues(opts)
	if err != nil {
		return nil, err
	}

	// Render templates
	manifests, notes, err := r.renderTemplates(chart, opts, values)
	if err != nil {
		return nil, err
	}

	return &RenderResult{
		Manifests: manifests,
		Notes:     notes,
	}, nil
}

// validateOptions validates the render options
func (r *ChartRenderer) validateOptions(opts RenderOptions) error {
	if opts.ChartPath == "" {
		return fmt.Errorf("chart path cannot be empty")
	}

	if opts.ReleaseName == "" {
		return fmt.Errorf("release name cannot be empty")
	}

	// Check if chart path exists
	if _, err := os.Stat(opts.ChartPath); os.IsNotExist(err) {
		return &ChartNotFoundError{Path: opts.ChartPath}
	}

	return nil
}

// loadChart loads a Helm chart from the filesystem
func (r *ChartRenderer) loadChart(chartPath string) (*chart.Chart, error) {
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, &ChartNotFoundError{Path: chartPath}
	}
	return chart, nil
}

// mergeValues parses and merges values from files and inline values
func (r *ChartRenderer) mergeValues(opts RenderOptions) (map[string]interface{}, error) {
	values := make(map[string]interface{})

	// First, load values from files
	for _, valuesFile := range opts.ValuesFiles {
		fileValues, err := r.loadValuesFile(valuesFile)
		if err != nil {
			return nil, err
		}
		values = r.mergeMaps(values, fileValues)
	}

	// Then, merge inline values (they take precedence)
	if opts.Values != nil {
		values = r.mergeMaps(values, opts.Values)
	}

	return values, nil
}

// loadValuesFile loads and parses a YAML values file
func (r *ChartRenderer) loadValuesFile(filename string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, &InvalidValuesError{File: filename, Err: err}
	}

	var values map[string]interface{}
	if err := yaml.Unmarshal(data, &values); err != nil {
		return nil, &InvalidValuesError{File: filename, Err: err}
	}

	return values, nil
}

// mergeMaps recursively merges two maps, with the second map taking precedence
func (r *ChartRenderer) mergeMaps(base, override map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy base map
	for k, v := range base {
		result[k] = v
	}

	// Merge override map
	for k, v := range override {
		if existing, exists := result[k]; exists {
			if existingMap, ok := existing.(map[string]interface{}); ok {
				if overrideMap, ok := v.(map[string]interface{}); ok {
					result[k] = r.mergeMaps(existingMap, overrideMap)
					continue
				}
			}
		}
		result[k] = v
	}

	return result
}

// renderTemplates renders the chart templates using Helm
func (r *ChartRenderer) renderTemplates(chart *chart.Chart, opts RenderOptions, values map[string]interface{}) ([]string, string, error) {
	// Create template action
	client := action.NewInstall(r.actionConfig)
	client.DryRun = true
	client.ReleaseName = opts.ReleaseName
	client.Namespace = opts.Namespace
	if client.Namespace == "" {
		client.Namespace = "default"
	}
	client.Replace = true
	client.ClientOnly = true

	// Render the templates
	release, err := client.Run(chart, values)
	if err != nil {
		return nil, "", &RenderError{Chart: opts.ChartPath, Err: err}
	}

	// Separate manifests
	manifests := r.separateManifests(release.Manifest)

	return manifests, release.Info.Notes, nil
}

// separateManifests splits a multi-document YAML string into individual manifests
func (r *ChartRenderer) separateManifests(manifestString string) []string {
	if manifestString == "" {
		return []string{}
	}

	// Split by YAML document separator
	documents := strings.Split(manifestString, "---")
	var manifests []string

	for _, doc := range documents {
		trimmed := strings.TrimSpace(doc)
		if trimmed == "" {
			continue
		}

		// TODO Do we want to check comments?
		// if strings.HasPrefix(trimmed, "#") {
		// 	manifests = append(manifests, trimmed)
		// }

		manifests = append(manifests, trimmed)
	}

	return manifests
}
