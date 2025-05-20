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
	// methodPattern matches public method declarations in normalized content, allowing for extra whitespace
	methodPattern = regexp.MustCompile(`public\s+([a-zA-Z0-9_$<>\[\]]+)\s+([a-zA-Z0-9_$]+)\s*\(([^)]*)\)`)
	// publicFieldPattern matches public field declarations with optional modifiers, type, name, and optional initializer
	publicFieldPattern = regexp.MustCompile(`public(?:\s+(?:static|final|transient|volatile))*\s+([a-zA-Z0-9_$\[\]]+)\s+([a-zA-Z0-9_$]+)(?:\s*=\s*[^;]+)?;`)
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

				// Remove any private classes from classContent before extracting public methods
				classContent = removePrivateClasses(classContent)

				// Extract public methods within the class definition
				methodMatches := methodPattern.FindAllStringSubmatch(classContent, -1)
				var declaredMethods []PublicMethod
				for _, match := range methodMatches {
					if len(match) >= 4 {
						// Check if return type is allowed
						if _, ok := allowedTypes[match[1]]; !ok {
							continue // Skip this method if return type is not allowed
						}
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

				// Extract public fields within the class definition
				fieldMatches := publicFieldPattern.FindAllStringSubmatch(classContent, -1)
				var declaredFields []PublicField
				for _, match := range fieldMatches {
					if len(match) >= 3 {
						declaredFields = append(declaredFields, PublicField{
							Type: match[1],
							Name: match[2],
						})
					}
				}

				// Append superclass methods to declaredMethods from sso_super.go
				declaredMethods = append(declaredMethods, SuperclassMethods...)

				// Create a new ServerSideObject and append it to the list
				matchingFiles = append(matchingFiles, ServerSideObject{
					FilePath:        path,
					ClassName:       className,
					PackageLine:     packageLine,
					DeclaredMethods: declaredMethods,
					DeclaredFields:  declaredFields,
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
		if len(parts) >= 2 {
			// Remove allowed parameter modifiers (final, annotations)
			j := 0
			for j < len(parts)-2 {
				if parts[j] == "final" || strings.HasPrefix(parts[j], "@") {
					j++
				} else {
					break
				}
			}
			// The type is at parts[j], the name is at parts[j+1]
			parameters = append(parameters, Parameter{
				Type: parts[j],
				Name: parts[j+1],
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

// removePrivateClasses removes all private class definitions (with nested braces) from the input string.
func removePrivateClasses(input string) string {
	for {
		startIdx := strings.Index(input, "private class ")
		if startIdx == -1 {
			break
		}
		// Find the opening brace for the class
		braceIdx := strings.Index(input[startIdx:], "{")
		if braceIdx == -1 {
			break
		}
		braceIdx += startIdx
		// Use a counter to find the matching closing brace
		count := 1
		endIdx := braceIdx + 1
		for endIdx < len(input) && count > 0 {
			if input[endIdx] == '{' {
				count++
			} else if input[endIdx] == '}' {
				count--
			}
			endIdx++
		}
		// Remove the private class definition
		if count == 0 {
			input = input[:startIdx] + input[endIdx:]
		} else {
			// Unmatched braces, break to avoid infinite loop
			break
		}
	}
	return input
}
