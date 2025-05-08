package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	// packagePattern matches package declarations in normalized content
	packagePattern = regexp.MustCompile(`package ([a-zA-Z0-9_.]+);`)
	// classPattern matches public class declarations extending ServerSideObject in normalized content
	classPattern = regexp.MustCompile(`public class [a-zA-Z0-9_$]+ extends ServerSideObject`)
	// methodPattern matches public method declarations in normalized content
	methodPattern = regexp.MustCompile(`public ([a-zA-Z0-9_$<>\[\]]+) ([a-zA-Z0-9_$]+)\(([^)]*)\)`)
)

// ScanForSSOs scans .java files in the given directory and returns a list of files that contain an SSO.
func ScanForSSOs(directory string) (ServerSideObjectList, error) {
	var matchingFiles ServerSideObjectList

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".java") {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			// Read the file content
			content, err := io.ReadAll(file)
			if err != nil {
				return err
			}

			// Normalize the content by removing newlines and extra spaces
			normalizedContent := strings.Join(strings.Fields(string(content)), " ")

			// Check if the file contains a public class extending ServerSideObject
			if classPattern.MatchString(normalizedContent) {
				className := info.Name()[:len(info.Name())-len(filepath.Ext(info.Name()))] // File name without extension

				// Output statement to indicate the SSO was found and is being parsed
				fmt.Printf("SSO found: %s.\n", className)

				// Extract package string
				packageMatch := packagePattern.FindStringSubmatch(normalizedContent)
				var packageLine string
				if len(packageMatch) > 1 {
					packageLine = packageMatch[1]
				}

				// Locate the class definition boundaries
				classStart := strings.Index(normalizedContent, "class "+className+" extends ServerSideObject")
				classEnd := strings.LastIndex(normalizedContent, "}")
				if classStart == -1 || classEnd == -1 || classStart >= classEnd {
					return nil // Invalid class definition
				}
				classContent := normalizedContent[classStart : classEnd+1]

				// Extract public methods within the class definition
				methodMatches := methodPattern.FindAllStringSubmatch(classContent, -1)
				var declaredMethods []PublicMethod
				for _, match := range methodMatches {
					if len(match) >= 4 {
						parameters := extractParameters(match[3])

						// Check if all parameter types are valid
						if !areParametersValid(parameters) {
							continue // Skip this method if an invalid parameter type is found
						}

						declaredMethods = append(declaredMethods, PublicMethod{
							AccessModifier: "public",
							ReturnType:     match[1],
							MethodName:     match[2],
							Parameters:     parameters,
						})
					}
				}

				// Create a new ServerSideObject and append it to the list
				matchingFiles = append(matchingFiles, ServerSideObject{
					FilePath:        path,
					ClassName:       className,
					PackageLine:     packageLine,
					DeclaredMethods: declaredMethods,
				})

				// Pretty print the added ServerSideObject
				// model.PrettyPrintStruct(matchingFiles[len(matchingFiles)-1])
			}
		}
		return nil
	})

	// Sort the matchingFiles by ClassName before returning
	sort.Sort(matchingFiles)

	return matchingFiles, err
}

// Helper function to extract parameters from a method signature
func extractParameters(paramString string) []Parameter {
	var parameters []Parameter
	if paramString == "" {
		return parameters // No parameters
	}

	paramPairs := strings.Split(paramString, ",")
	for _, pair := range paramPairs {
		parts := strings.Fields(strings.TrimSpace(pair))
		if len(parts) == 2 {
			parameters = append(parameters, Parameter{
				Type: parts[0],
				Name: parts[1],
			})
		}
	}
	return parameters
}

// areParametersValid checks if all parameter types are in the allowed list.
func areParametersValid(parameters []Parameter) bool {
	for _, param := range parameters {
		if _, ok := allowedTypes[param.Type]; !ok {
			return false
		}
	}
	return true
}
