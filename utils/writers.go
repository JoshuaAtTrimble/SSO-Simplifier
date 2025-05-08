package utils

import (
	"os"
	"path/filepath"
)

// WriteSimplifiedSSO writes a ServerSideObject to a simplified .java file with a default constructor and minimal method bodies.
func WriteSimplifiedSSO(outputDir string, sso *ServerSideObject) error {
	// Ensure the output directory exists
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return err
	}

	// Construct the output file path
	outputFilePath := filepath.Join(outputDir, sso.ClassName+".java")

	// Open the file for writing
	file, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the simplified SSO content
	if _, err := file.WriteString("package " + sso.PackageLine + ";\n\n"); err != nil {
		return err
	}
	if _, err := file.WriteString("public class " + sso.ClassName + " {\n\n"); err != nil {
		return err
	}

	// Write the empty public constructor
	if _, err := file.WriteString("    public " + sso.ClassName + "() {}\n\n"); err != nil {
		return err
	}

	for _, method := range sso.DeclaredMethods {
		methodSignature := "    public " + method.ReturnType + " " + method.MethodName + "("
		for i, param := range method.Parameters {
			if i > 0 {
				methodSignature += ", "
			}
			methodSignature += param.Type + " " + param.Name
		}
		methodSignature += ") {\n"

		// Simplify the method body with a return statement for the simplest form of the return type
		if method.ReturnType != "void" {
			methodBody := "        return "
			if defaultValue, ok := allowedTypes[method.ReturnType]; ok {
				methodBody += defaultValue + ";"
			} else {
				methodBody += "null;" // Fallback for unsupported types
			}
			methodSignature += methodBody + "\n"
		}
		methodSignature += "    }\n\n"

		if _, err := file.WriteString(methodSignature); err != nil {
			return err
		}
	}
	if _, err := file.WriteString("}\n"); err != nil {
		return err
	}

	return nil
}
