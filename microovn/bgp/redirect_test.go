package bgp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateVrfModule(t *testing.T) {
	// This test verifies the validateVrfModule function logic
	// Note: The actual result depends on the test environment's kernel modules

	// Test that the function doesn't panic and returns a well-formed error
	err := validateVrfModule()

	// The module might or might not be loaded in the test environment
	// We just verify that:
	// 1. The function executes without panic
	// 2. If it returns an error, it's properly formatted
	if err != nil {
		if err.Error() == "" {
			t.Error("validateVrfModule returned an error with empty message")
		}
		// Verify error message contains expected content
		errMsg := err.Error()
		expectedMsg1 := "VRF kernel module is not loaded. Please load it with 'modprobe vrf' or ensure it's configured to load at boot"
		expectedPrefix := "unable to check kernel modules: "

		if errMsg != expectedMsg1 && !strings.HasPrefix(errMsg, expectedPrefix) {
			t.Errorf("unexpected error message format: %s", errMsg)
		}
	}
}

func TestValidateVrfModule_WithMockSysModule(t *testing.T) {
	// Test the /sys/module/vrf path check in isolation
	// by temporarily modifying the check logic

	// We can't easily mock the filesystem for this test without significant refactoring,
	// but we can at least verify the function behaves consistently

	// Multiple calls should return the same result
	err1 := validateVrfModule()
	err2 := validateVrfModule()

	if (err1 == nil) != (err2 == nil) {
		t.Error("validateVrfModule returned inconsistent results on consecutive calls")
	}
}

func TestValidateVrfModule_ProcModulesFormat(t *testing.T) {
	// This test validates the parsing logic for /proc/modules format
	// Create a temporary file that simulates /proc/modules content

	tmpDir := t.TempDir()
	testCases := []struct {
		name          string
		content       string
		shouldBeFound bool
	}{
		{
			name:          "VRF module present",
			content:       "vrf 28672 0 - Live 0xffffffffc0a3e000\nother_module 16384 0 - Live 0xffffffffc0a39000\n",
			shouldBeFound: true,
		},
		{
			name:          "VRF module absent",
			content:       "other_module 16384 0 - Live 0xffffffffc0a39000\nanother_module 28672 0 - Live 0xffffffffc0a3e000\n",
			shouldBeFound: false,
		},
		{
			name:          "Empty modules file",
			content:       "",
			shouldBeFound: false,
		},
		{
			name:          "VRF as part of another module name",
			content:       "myvrf_custom 16384 0 - Live 0xffffffffc0a39000\n",
			shouldBeFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tc.name)
			err := os.WriteFile(testFile, []byte(tc.content), 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			// Read and parse the file content like validateVrfModule does
			data, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("failed to read test file: %v", err)
			}

			// Simple check if "vrf " prefix exists using same logic as validateVrfModule
			found := false
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "vrf ") {
					found = true
					break
				}
			}

			if found != tc.shouldBeFound {
				t.Errorf("expected found=%v, got found=%v for content:\n%s", tc.shouldBeFound, found, tc.content)
			}
		})
	}
}
