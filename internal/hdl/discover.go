package hdl

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FileType represents the type of HDL file
type FileType string

const (
	Package FileType = "package" // VHDL packages
	RTL     FileType = "rtl"      // Synthesizable RTL
	Behav   FileType = "behav"    // Behavioral/testbench
)

// HDLFiles contains categorized HDL files for a component
type HDLFiles struct {
	Package []string // VHDL package files
	RTL     []string // Synthesizable RTL files
	Behav   []string // Behavioral/testbench files
}

// Architecture name mappings for VHDL
var archMap = map[string]FileType{
	"rtl":        RTL,
	"impl":       RTL,
	"structural": RTL,
	"behavioral": RTL,
	"testbench":  Behav,
	"asim":       Behav,
	"sim":        Behav,
}

// DiscoverFiles finds and categorizes HDL files in a component directory
func DiscoverFiles(componentPath string) (*HDLFiles, error) {
	files := &HDLFiles{
		Package: make([]string, 0),
		RTL:     make([]string, 0),
		Behav:   make([]string, 0),
	}

	// Check if directory exists
	if _, err := os.Stat(componentPath); os.IsNotExist(err) {
		return files, nil // Return empty if component not checked out
	}

	// Find all HDL files
	verilogFiles, err := filepath.Glob(filepath.Join(componentPath, "*.v"))
	if err != nil {
		return nil, err
	}

	systemVerilogFiles, err := filepath.Glob(filepath.Join(componentPath, "*.sv"))
	if err != nil {
		return nil, err
	}

	systemVerilogHeaders, err := filepath.Glob(filepath.Join(componentPath, "*.svh"))
	if err != nil {
		return nil, err
	}

	vhdlFiles, err := filepath.Glob(filepath.Join(componentPath, "*.vhd"))
	if err != nil {
		return nil, err
	}

	// Process Verilog files
	for _, file := range verilogFiles {
		if strings.HasSuffix(file, "_tb.v") {
			files.Behav = append(files.Behav, file)
		} else {
			files.RTL = append(files.RTL, file)
		}
	}

	// Process SystemVerilog files
	for _, file := range systemVerilogFiles {
		if strings.HasSuffix(file, "_tb.sv") {
			files.Behav = append(files.Behav, file)
		} else {
			files.RTL = append(files.RTL, file)
		}
	}

	// Process SystemVerilog headers
	for _, file := range systemVerilogHeaders {
		files.RTL = append(files.RTL, file)
	}

	// Process VHDL files
	packages := make(map[string]bool)
	for _, file := range vhdlFiles {
		fileType, isPackage, err := classifyVHDLFile(file)
		if err != nil {
			// If we can't classify, skip it
			continue
		}

		if isPackage {
			packages[file] = true
		} else {
			switch fileType {
			case RTL:
				files.RTL = append(files.RTL, file)
			case Behav:
				files.Behav = append(files.Behav, file)
			}
		}
	}

	// Add packages
	for pkg := range packages {
		files.Package = append(files.Package, pkg)
	}

	return files, nil
}

// classifyVHDLFile reads a VHDL file and determines its type based on architecture name
func classifyVHDLFile(filePath string) (FileType, bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Regex patterns
	archPattern := regexp.MustCompile(`(?i)architecture\s+(\w+)\s+of\s+(\w+)\s+is`)
	packagePattern := regexp.MustCompile(`(?i)package\s+(body\s+)?(\w+)\s+is`)

	isPackage := false
	var fileType FileType

	for scanner.Scan() {
		line := scanner.Text()

		// Check for package
		if packagePattern.MatchString(line) {
			isPackage = true
			continue
		}

		// Check for architecture
		if matches := archPattern.FindStringSubmatch(line); matches != nil {
			archName := strings.ToLower(matches[1])
			if mappedType, ok := archMap[archName]; ok {
				fileType = mappedType
			} else {
				// Unknown architecture, default to RTL
				fileType = RTL
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", false, err
	}

	return fileType, isPackage, nil
}
