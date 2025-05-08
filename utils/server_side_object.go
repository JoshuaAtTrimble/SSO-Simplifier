package utils

import (
	"encoding/json"
	"fmt"
)

// ServerSideObject represents a Java file with its path, name, and declared methods.
type ServerSideObject struct {
	FilePath        string         // The absolute or relative path of the file
	ClassName       string         // The name of the class
	PackageLine     string         // The package line of the Java file
	DeclaredMethods []PublicMethod // The declared methods of the class
}

// PublicMethod represents a Java method signature broken into elements.
type PublicMethod struct {
	AccessModifier string      // The access modifier of the method (e.g., public, private, protected)
	ReturnType     string      // The return type of the method
	MethodName     string      // The name of the method
	Parameters     []Parameter // The parameters of the method
}

// Parameter represents a parameter in a Java method signature.
type Parameter struct {
	Type string // The type of the parameter (e.g., int, String)
	Name string // The name of the parameter
}

// allowedTypes defines the list of allowed parameter types and their default return values.
var allowedTypes = map[string]string{
	"boolean": "false",
	"byte":    "0",
	"char":    "'\\0'",
	"short":   "0",
	"int":     "0",
	"long":    "0L",
	"float":   "0.0f",
	"double":  "0.0",
	"String":  "null",
}

// ServerSideObjectList is a custom type that implements sort.Interface for []ServerSideObject.
type ServerSideObjectList []ServerSideObject

// Len returns the length of the list.
func (s ServerSideObjectList) Len() int {
	return len(s)
}

// Less compares two ServerSideObjects by ClassName for sorting.
func (s ServerSideObjectList) Less(i, j int) bool {
	return s[i].ClassName < s[j].ClassName
}

// Swap swaps two ServerSideObjects in the list.
func (s ServerSideObjectList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// PrettyPrintStruct prints a struct in a nested, hierarchical format.
func PrettyPrintStruct(s interface{}) {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling struct:", err)
		return
	}
	fmt.Println(string(data))
}
