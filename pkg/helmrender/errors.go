package helmrender

import "fmt"

// ChartNotFoundError is returned when a chart cannot be found
type ChartNotFoundError struct {
	Path string
}

func (e ChartNotFoundError) Error() string {
	return fmt.Sprintf("chart not found at path: %s", e.Path)
}

// InvalidValuesError is returned when values file is invalid
type InvalidValuesError struct {
	File string
	Err  error
}

func (e InvalidValuesError) Error() string {
	return fmt.Sprintf("invalid values file %s: %v", e.File, e.Err)
}

// RenderError is returned when chart rendering fails
type RenderError struct {
	Chart string
	Err   error
}

func (e RenderError) Error() string {
	return fmt.Sprintf("failed to render chart %s: %v", e.Chart, e.Err)
}