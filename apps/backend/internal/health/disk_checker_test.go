package health

import (
	"context"
	"testing"
	"time"
)

func TestDiskChecker(t *testing.T) {
	t.Run("Checker name is 'disk'", func(t *testing.T) {
		checker := NewDiskChecker("/")
		if checker.Name() != "disk" {
			t.Errorf("Expected checker name 'disk', got '%s'", checker.Name())
		}
	})

	t.Run("Check disk health with valid path", func(t *testing.T) {
		checker := NewDiskChecker("/")
		ctx := context.Background()
		result := checker.Check(ctx)

		// Should return a result
		if result.Status == "" {
			t.Error("Expected status to be set")
		}

		// Status should be one of the valid values
		if result.Status != CheckPass && result.Status != CheckWarn && result.Status != CheckFail {
			t.Errorf("Invalid status: %s", result.Status)
		}

		// Response time should be populated
		if result.ResponseTime == "" {
			t.Error("Expected response time to be set")
		}

		// Message should be populated
		if result.Message == "" {
			t.Error("Expected message to be set")
		}
	})

	t.Run("Check disk health metrics accuracy", func(t *testing.T) {
		checker := NewDiskChecker("/")
		ctx := context.Background()
		result := checker.Check(ctx)

		// Verify result structure
		if result.Details == nil {
			t.Error("Expected details to be populated with disk metrics")
		}

		// If we have details, verify they contain expected fields
		if details, ok := result.Details.(map[string]interface{}); ok {
			// Check for common disk metrics fields
			expectedFields := []string{"used_gb", "total_gb", "free_percentage", "free_gb"}
			for _, field := range expectedFields {
				if _, exists := details[field]; !exists {
					t.Logf("Warning: Expected field '%s' not found in disk details", field)
				}
			}
		}
	})

	t.Run("Check completes within timeout", func(t *testing.T) {
		checker := NewDiskChecker("/")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		start := time.Now()
		result := checker.Check(ctx)
		duration := time.Since(start)

		// Check should complete quickly
		if result.Status == "" {
			t.Error("Check did not complete")
		}

		// Should complete well within timeout
		if duration > 1*time.Second {
			t.Logf("Warning: Disk check took %v, consider optimization", duration)
		}
	})

	t.Run("Handle invalid path gracefully", func(t *testing.T) {
		checker := NewDiskChecker("/nonexistent/path/that/should/not/exist")
		ctx := context.Background()
		result := checker.Check(ctx)

		// Should return a failure status rather than panic
		if result.Status != CheckFail && result.Status != CheckWarn {
			t.Logf("Expected CheckFail or CheckWarn for invalid path, got %s", result.Status)
		}

		// Should still provide a message
		if result.Message == "" {
			t.Error("Expected error message for invalid path")
		}
	})
}

func TestDiskMetricsThresholds(t *testing.T) {
	t.Run("Check low disk space warning", func(t *testing.T) {
		checker := NewDiskChecker("/")
		ctx := context.Background()
		result := checker.Check(ctx)

		// Check if details contain disk usage information
		if details, ok := result.Details.(map[string]interface{}); ok {
			if freePercent, ok := details["free_percentage"].(float64); ok {
				// If free space is below 20%, status should be CheckWarn or CheckFail
				if freePercent < 20.0 {
					if result.Status == CheckPass {
						t.Error("Expected warning or failure status when disk space below 20%")
					}
				}
			}
		}
	})

	t.Run("Check critical disk space threshold", func(t *testing.T) {
		checker := NewDiskChecker("/")
		ctx := context.Background()
		result := checker.Check(ctx)

		// Check if details contain disk usage information
		if details, ok := result.Details.(map[string]interface{}); ok {
			if freePercent, ok := details["free_percentage"].(float64); ok {
				// If free space is below 10%, status should be CheckFail (critical)
				if freePercent < 10.0 {
					if result.Status != CheckFail {
						t.Error("Expected critical failure status when disk space below 10%")
					}
				}
			}
		}
	})
}
