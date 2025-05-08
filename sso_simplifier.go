package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/JoshuaAtTrimble/SSO-Simplifier/utils"
)

// printHelp prints the help message for the program, indicating required flags and available options.
func printHelp() {
	fmt.Println()
	fmt.Println("sso_simplifier simplifies SSO Java class files for the VIP SSO Gallery by extracting the package line, class signature, and public method signatures with minimal method code.")
	fmt.Println("Usage: sso_simplifier [options]")
	fmt.Println("Options:")
	fmt.Println("  --help          Display help information.")
	fmt.Println("  --inputPath     (Required) Path to search for ServerSideObjects (SSOs) to simplify.")
	fmt.Println("  --outputPath    (Required) Path to save simplified SSOs.")
	fmt.Println("  --compile       Compile simplified SSOs into a single Java archive.")
	fmt.Println()
}

func main() {
	// If no arguments or flags are provided, behave as if the user entered the help flag
	if len(os.Args) == 1 {
		printHelp()
		os.Exit(0)
	}

	// Define command-line flags
	help := flag.Bool("help", false, "Display help information.")
	inputPath := flag.String("inputPath", "", "Path to search for ServerSideObjects (SSOs) to simplify.")
	outputPath := flag.String("outputPath", "", "Path to save simplified SSOs.")
	compile := flag.String("compile", "", "Compile simplified SSOs into a single Java archive.")

	flag.Parse()

	// After parsing flags, check if inputPath and outputPath are provided
	if *inputPath == "" || *outputPath == "" {
		fmt.Println("Error: Both --inputPath and --outputPath flags are required.")
		os.Exit(1)
	}

	if *help {
		printHelp()
		os.Exit(0)
	}

	// Retrieve a list of ServerSideObjects from the specified directory
	serverSideObjects, err := utils.ScanForSSOs(*inputPath)
	if err != nil {
		fmt.Printf("Error parsing directory: %v\n", err)
		os.Exit(1)
	}

	// Check if there are any matching ServerSideObjects and print the result
	if len(serverSideObjects) == 0 {
		fmt.Println("No matching files found.")
	} else {
		fmt.Printf("Parsed %d matching files.\n", len(serverSideObjects))
	}

	// Write each ServerSideObject to the determined output directory
	for _, sso := range serverSideObjects {
		err := utils.WriteSimplifiedSSO(*outputPath, &sso)
		if err != nil {
			fmt.Printf("Error writing simplified SSO for %s: %v\n", sso.ClassName, err)
		}
	}
	fmt.Printf("Simplified SSOs have been written to the output directory: %s\n", *outputPath)

	// Handle the compile flag
	if *compile != "" {
		compiledJarName := *compile
		if !strings.HasSuffix(compiledJarName, ".jar") {
			compiledJarName += ".jar"
		}

		// Output statement to indicate the start of the compilation process
		fmt.Printf("Compiling the simplified SSOs into: %s\n", compiledJarName)

		// Path to the compiled JAR file
		compiledJarPath := filepath.Join(*outputPath, compiledJarName)

		// Compile .java files into .class files
		javaFiles := []string{}
		err := filepath.Walk(*outputPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".java") {
				javaFiles = append(javaFiles, path)
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error finding .java files: %v\n", err)
			os.Exit(1)
		}

		if len(javaFiles) == 0 {
			fmt.Println("No .java files found to compile.")
			os.Exit(1)
		}

		// Compile the .java files
		cmd := exec.Command("javac", append([]string{"-d", *outputPath}, javaFiles...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error compiling .java files: %v\n", err)
			os.Exit(1)
		}

		// Create the .jar file
		cmd = exec.Command("jar", "cf", compiledJarPath, "-C", *outputPath, ".")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error creating .jar file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Compiled .jar file created at: %s\n", compiledJarPath)
	}
}
