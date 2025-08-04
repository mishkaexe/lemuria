package test

import (
	"testing"

	"github.com/mishkaexe/lemuria/pkg/diff"
	"github.com/stretchr/testify/assert"
)

func TestCompareStrings_IdenticalStrings(t *testing.T) {
	d := diff.New()
	result, err := d.CompareStrings("hello", "hello")
	assert.NoError(t, err)
	assert.False(t, result.HasDifferences())
}

func TestCompareStrings_DifferentStrings(t *testing.T) {
	d := diff.New()
	result, err := d.CompareStrings("hello", "world")
	assert.NoError(t, err)
	assert.True(t, result.HasDifferences())
	assert.Contains(t, result.String(), "-hello")
	assert.Contains(t, result.String(), "+world")
}

func TestCompareStrings_MultilineUnifiedDiff(t *testing.T) {
	d := diff.New()
	old := `line1
line2
line3
line4
line5`
	new := `line1
line2_modified
line3
line4
line5`
	
	result, err := d.CompareStrings(old, new)
	assert.NoError(t, err)
	assert.True(t, result.HasDifferences())
	
	// Expect unified diff format with context lines
	output := result.String()
	assert.Contains(t, output, "@@")  // Line number header
	assert.Contains(t, output, "-line2")  // Removed line
	assert.Contains(t, output, "+line2_modified")  // Added line
	assert.Contains(t, output, " line1")  // Context line (unchanged)
	assert.Contains(t, output, " line3")  // Context line (unchanged)
}

func TestCompareStrings_MultipleLinesAdded(t *testing.T) {
	d := diff.New()
	old := `line1
line3
line5`
	new := `line1
line2
line3
line4
line5`
	
	result, err := d.CompareStrings(old, new)
	assert.NoError(t, err)
	assert.True(t, result.HasDifferences())
	
	output := result.String()
	assert.Contains(t, output, "+line2")
	assert.Contains(t, output, "+line4")
	assert.Contains(t, output, " line1")  // Context
	assert.Contains(t, output, " line3")  // Context
	assert.Contains(t, output, " line5")  // Context
}

func TestCompareStrings_MultipleLinesRemoved(t *testing.T) {
	d := diff.New()
	old := `line1
line2
line3
line4
line5`
	new := `line1
line3
line5`
	
	result, err := d.CompareStrings(old, new)
	assert.NoError(t, err)
	assert.True(t, result.HasDifferences())
	
	output := result.String()
	assert.Contains(t, output, "-line2")
	assert.Contains(t, output, "-line4")
	assert.Contains(t, output, " line1")  // Context
	assert.Contains(t, output, " line3")  // Context
	assert.Contains(t, output, " line5")  // Context
}

func TestCompareManifests_DeploymentImageDifference(t *testing.T) {
	d := diff.New()
	
	oldManifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  labels:
    app: my-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
        - name: my-app
          image: "nginx:1.20"
          ports:
            - containerPort: 80`

	newManifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  labels:
    app: my-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
        - name: my-app
          image: "nginx:1.21"
          ports:
            - containerPort: 80`

	result, err := d.CompareManifests(oldManifest, newManifest)
	assert.NoError(t, err)
	assert.True(t, result.HasDifferences())
	
	output := result.String()
	assert.Contains(t, output, "-          image: \"nginx:1.20\"")
	assert.Contains(t, output, "+          image: \"nginx:1.21\"")
	assert.Contains(t, output, " kind: Deployment")  // Context line
}

func TestCompareManifests_DeploymentReplicasAndEnvVars(t *testing.T) {
	d := diff.New()
	
	oldManifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 2
  template:
    spec:
      containers:
        - name: my-app
          image: "nginx:1.20"
          env:
            - name: ENV
              value: "dev"`

	newManifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 5
  template:
    spec:
      containers:
        - name: my-app
          image: "nginx:1.20"
          env:
            - name: ENV
              value: "prod"
            - name: LOG_LEVEL
              value: "info"`

	result, err := d.CompareManifests(oldManifest, newManifest)
	assert.NoError(t, err)
	assert.True(t, result.HasDifferences())
	
	output := result.String()
	assert.Contains(t, output, "-  replicas: 2")
	assert.Contains(t, output, "+  replicas: 5")
	assert.Contains(t, output, "-              value: \"dev\"")
	assert.Contains(t, output, "+              value: \"prod\"")
	assert.Contains(t, output, "+            - name: LOG_LEVEL")
	assert.Contains(t, output, "+              value: \"info\"")
}

func TestCompareManifests_IdenticalDeployments(t *testing.T) {
	d := diff.New()
	
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 1`

	result, err := d.CompareManifests(manifest, manifest)
	assert.NoError(t, err)
	assert.False(t, result.HasDifferences())
}

func TestCompareMultipleManifests_DeploymentAndService(t *testing.T) {
	d := diff.New()
	
	oldManifests := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 2
---
apiVersion: v1
kind: Service
metadata:
  name: my-app-service
spec:
  selector:
    app: my-app
  ports:
    - port: 80
      targetPort: 80`

	newManifests := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 3
---
apiVersion: v1
kind: Service
metadata:
  name: my-app-service
spec:
  selector:
    app: my-app
  ports:
    - port: 80
      targetPort: 8080
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-app-config
data:
  config.yaml: |
    key: value`

	result, err := d.CompareMultipleManifests(oldManifests, newManifests)
	assert.NoError(t, err)
	assert.True(t, result.HasDifferences())
	
	output := result.String()
	assert.Contains(t, output, "-  replicas: 2")
	assert.Contains(t, output, "+  replicas: 3")
	assert.Contains(t, output, "-      targetPort: 80")
	assert.Contains(t, output, "+      targetPort: 8080")
	assert.Contains(t, output, "+apiVersion: v1")  // New ConfigMap
	assert.Contains(t, output, "+kind: ConfigMap")
}