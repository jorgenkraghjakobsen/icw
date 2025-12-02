package hdl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverFiles(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create some test files
	testFiles := map[string]string{
		"module1.v":      "module module1();",
		"module2_tb.v":   "module module2_tb();",
		"module3.sv":     "module module3();",
		"module4_tb.sv":  "module module4_tb();",
		"header.svh":     "`define TEST",
		"entity1.vhd":    "architecture rtl of entity1 is",
		"entity2.vhd":    "architecture testbench of entity2 is",
		"package1.vhd":   "package my_pkg is",
	}

	for filename, content := range testFiles {
		filepath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Discover files
	hdlFiles, err := DiscoverFiles(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverFiles failed: %v", err)
	}

	// Verify Verilog files
	expectedRTL := 2 // module1.v, module3.sv
	if len(hdlFiles.RTL) < expectedRTL {
		t.Errorf("Expected at least %d RTL files, got %d", expectedRTL, len(hdlFiles.RTL))
	}

	// Verify behavioral files
	expectedBehav := 2 // module2_tb.v, module4_tb.sv
	if len(hdlFiles.Behav) < expectedBehav {
		t.Errorf("Expected at least %d behavioral files, got %d", expectedBehav, len(hdlFiles.Behav))
	}

	// Headers go to RTL
	hasHeader := false
	for _, file := range hdlFiles.RTL {
		if filepath.Base(file) == "header.svh" {
			hasHeader = true
			break
		}
	}
	if !hasHeader {
		t.Error("Expected header.svh in RTL files")
	}
}

func TestDiscoverFilesEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	hdlFiles, err := DiscoverFiles(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverFiles failed: %v", err)
	}

	if len(hdlFiles.RTL) != 0 {
		t.Errorf("Expected 0 RTL files, got %d", len(hdlFiles.RTL))
	}
	if len(hdlFiles.Behav) != 0 {
		t.Errorf("Expected 0 behavioral files, got %d", len(hdlFiles.Behav))
	}
	if len(hdlFiles.Package) != 0 {
		t.Errorf("Expected 0 package files, got %d", len(hdlFiles.Package))
	}
}

func TestDiscoverFilesNonExistent(t *testing.T) {
	hdlFiles, err := DiscoverFiles("/nonexistent/path")
	if err != nil {
		t.Fatalf("DiscoverFiles should not error on non-existent path: %v", err)
	}

	// Should return empty lists
	if len(hdlFiles.RTL) != 0 || len(hdlFiles.Behav) != 0 || len(hdlFiles.Package) != 0 {
		t.Error("Expected empty results for non-existent path")
	}
}

func TestClassifyVHDLFile(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		name        string
		content     string
		expectedType FileType
		isPackage   bool
	}{
		{
			name:        "rtl_architecture",
			content:     "architecture rtl of my_entity is\nbegin\nend rtl;",
			expectedType: RTL,
			isPackage:   false,
		},
		{
			name:        "testbench_architecture",
			content:     "architecture testbench of my_entity_tb is\nbegin\nend testbench;",
			expectedType: Behav,
			isPackage:   false,
		},
		{
			name:        "package",
			content:     "package my_package is\nend package;",
			expectedType: "",
			isPackage:   true,
		},
		{
			name:        "package_body",
			content:     "package body my_package is\nend package body;",
			expectedType: "",
			isPackage:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tc.name+".vhd")
			if err := os.WriteFile(testFile, []byte(tc.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			fileType, isPackage, err := classifyVHDLFile(testFile)
			if err != nil {
				t.Fatalf("classifyVHDLFile failed: %v", err)
			}

			if isPackage != tc.isPackage {
				t.Errorf("Expected isPackage=%v, got %v", tc.isPackage, isPackage)
			}

			if !tc.isPackage && fileType != tc.expectedType {
				t.Errorf("Expected type %s, got %s", tc.expectedType, fileType)
			}
		})
	}
}
