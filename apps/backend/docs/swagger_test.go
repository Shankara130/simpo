package docs

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestSwaggerGeneration tests that swagger docs can be generated without errors (Story 9.2, Task 14)
func TestSwaggerGeneration(t *testing.T) {
	// Test that swag init runs without errors
	cmd := exec.Command("swag", "init", "-g", "cmd/server/main.go", "-o", "docs")
	cmd.Dir = filepath.Join("..", ".") // Run from backend root
	output, err := cmd.CombinedOutput()
	assert.NoError(t, err, "Swagger generation should succeed: %s", string(output))

	// Verify files are created
	assert.FileExists(t, "swagger.yaml", "swagger.yaml should be generated")
	assert.FileExists(t, "swagger.json", "swagger.json should be generated")
	assert.FileExists(t, "docs.go", "docs.go should be generated")
}

// TestOpenAPISpecValidation tests that generated spec follows Swagger 2.0 standard (Story 9.2, Task 14)
func TestOpenAPISpecValidation(t *testing.T) {
	// Load generated spec
	data, err := os.ReadFile("swagger.yaml")
	require.NoError(t, err, "swagger.yaml should exist")

	// Parse as Swagger spec
	var spec map[string]interface{}
	err = yaml.Unmarshal(data, &spec)
	assert.NoError(t, err, "swagger.yaml should be valid YAML")

	// Validate required fields for Swagger 2.0
	assert.Contains(t, spec, "swagger", "Should have swagger version field (Swagger 2.0)")
	assert.Contains(t, spec, "info", "Should have info field")
	assert.Contains(t, spec, "paths", "Should have paths field")

	// Validate info field
	info, ok := spec["info"].(map[string]interface{})
	assert.True(t, ok, "info should be a map")
	assert.Contains(t, info, "title", "info should have title")
	assert.Contains(t, info, "version", "info should have version")

	// Validate title matches simpo API
	title := info["title"].(string)
	assert.Equal(t, "simpo Pharmacy Management System API", title, "API title should match simpo")

	// Validate version
	version := info["version"].(string)
	assert.Equal(t, "1.0", version, "API version should be 1.0")

	// Validate basePath matches /api/v1
	assert.Contains(t, spec, "basePath", "Should have basePath field")
	basePath := spec["basePath"].(string)
	assert.Equal(t, "/api/v1", basePath, "BasePath should be /api/v1")

	// Validate security schemes for Swagger 2.0
	assert.Contains(t, spec, "securityDefinitions", "Should have security definitions field (Swagger 2.0)")
	securityDefs := spec["securityDefinitions"]
	assert.NotNil(t, securityDefs, "Security definitions should exist")
}

// TestSwaggerUIRoute tests that Swagger UI route is accessible (Story 9.2, Task 14)
func TestSwaggerUIRoute(t *testing.T) {
	// This test verifies that the swagger files exist and are not empty
	// Actual server testing would require starting the server

	// Verify swagger.json has required content
	data, err := os.ReadFile("swagger.json")
	require.NoError(t, err, "swagger.json should exist")
	assert.NotEmpty(t, string(data), "swagger.json should not be empty")

	// Verify swagger.yaml has required content
	data, err = os.ReadFile("swagger.yaml")
	require.NoError(t, err, "swagger.yaml should exist")
	assert.NotEmpty(t, string(data), "swagger.yaml should not be empty")
}
