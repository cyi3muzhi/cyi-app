package toon

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// Node represents a TOON data node
type Node interface {
	// Value returns the underlying value
	Value() interface{}
}

// Object represents a TOON object (dictionary)
type Object map[string]Node

// Value returns the underlying map
func (o Object) Value() interface{} {
	result := make(map[string]interface{})
	for k, v := range o {
		result[k] = v.Value()
	}
	return result
}

// Array represents a TOON array (list)
type Array []Node

// Value returns the underlying slice
func (a Array) Value() interface{} {
	result := make([]interface{}, len(a))
	for i, v := range a {
		result[i] = v.Value()
	}
	return result
}

// String represents a TOON string value
type String string

// Value returns the underlying string
func (s String) Value() interface{} {
	return string(s)
}

// Number represents a TOON number value
type Number float64

// Value returns the underlying float64
func (n Number) Value() interface{} {
	return float64(n)
}

// Boolean represents a TOON boolean value
type Boolean bool

// Value returns the underlying bool
func (b Boolean) Value() interface{} {
	return bool(b)
}

// ParseError represents an error during parsing
type ParseError struct {
	Line int
	Msg  string
}

// Error returns the error message
func (e *ParseError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Msg)
}

// Parse parses TOON data from a string
func Parse(data string) (Node, error) {
	lines := strings.Split(data, "\n")
	return parseRoot(lines)
}

// ParseReader parses TOON data from a reader
func ParseReader(r io.Reader) (Node, error) {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, r)
	if err != nil {
		return nil, err
	}
	return Parse(buf.String())
}

// parseRoot parses the root of the TOON document
func parseRoot(lines []string) (Node, error) {
	// Skip empty lines at the beginning
	currentLine := 0
	for currentLine < len(lines) && strings.TrimSpace(lines[currentLine]) == "" {
		currentLine++
	}
	
	if currentLine >= len(lines) {
		return nil, errors.New("empty input")
	}
	
	// Check if it's a single primitive value
	lineText := strings.TrimSpace(lines[currentLine])
	if !strings.Contains(lineText, " ") {
		// Check if there are no more lines or next line is not indented
		if currentLine+1 >= len(lines) {
			return parsePrimitive(lineText)
		}
		
		nextLine := lines[currentLine+1]
		nextIndent := getIndentation(nextLine)
		if nextIndent <= 0 {
			return parsePrimitive(lineText)
		}
	}
	
	// Parse the entire document as a series of key-value pairs
	obj := Object{}
	
	for currentLine < len(lines) {
		key, value, nextLine, err := parseItem(lines, currentLine, 0)
		if err != nil {
			return nil, err
		}
		
		if key != "" {
			obj[key] = value
		}
		
		currentLine = nextLine
	}
	
	return obj, nil
}

// parseItem parses a single TOON item (key-value pair or nested structure)
func parseItem(lines []string, startLine int, parentIndent int) (string, Node, int, error) {
	// Skip empty lines
	for startLine < len(lines) && strings.TrimSpace(lines[startLine]) == "" {
		startLine++
	}
	
	if startLine >= len(lines) {
		return "", nil, startLine, nil
	}
	
	// Get current line's indentation and text
	currentLine := lines[startLine]
	indent := getIndentation(currentLine)
	lineText := strings.TrimSpace(currentLine)
	
	// If indentation is less than parent, we're done with this level
	if indent < parentIndent {
		return "", nil, startLine, nil
	}
	
	// Split into key and value
	parts := strings.SplitN(lineText, " ", 2)
	key := parts[0]
	
	// Check if this is a nested structure
	if startLine+1 < len(lines) {
		nextLine := lines[startLine+1]
		nextIndent := getIndentation(nextLine)
		
		if nextIndent > indent {
			// This is a nested structure
			if !strings.Contains(strings.TrimSpace(nextLine), " ") {
				// It's an array
				arr, endLine, err := parseArray(lines, startLine, indent)
				return key, arr, endLine, err
			} else {
				// It's an object
				obj, endLine, err := parseObject(lines, startLine, indent)
				return key, obj, endLine, err
			}
		}
	}
	
	// It's a simple key-value pair
	var value Node
	var err error
	
	if len(parts) == 2 {
		// Has a value on the same line
		valueStr := strings.TrimSpace(parts[1])
		value, err = parsePrimitive(valueStr)
	} else {
		// No value
		value = String("")
	}
	
	if err != nil {
		return "", nil, startLine, err
	}
	
	return key, value, startLine + 1, nil
}

// parseObject parses an object
func parseObject(lines []string, startLine int, parentIndent int) (Object, int, error) {
	obj := Object{}
	currentLine := startLine + 1
	
	// Get the indentation level of this object's children
	// It should be the indentation of the first child line
	childIndent := parentIndent + 2 // Assume 2 spaces indentation for children
	
	// Find the actual child indentation from the first non-empty line
	for currentLine < len(lines) {
		if strings.TrimSpace(lines[currentLine]) != "" {
			childIndent = getIndentation(lines[currentLine])
			break
		}
		currentLine++
	}
	
	for currentLine < len(lines) {
		// Skip empty lines
		if strings.TrimSpace(lines[currentLine]) == "" {
			currentLine++
			continue
		}
		
		// Get current line's indentation
		lineIndent := getIndentation(lines[currentLine])
		
		// If current line's indentation is less than or equal to parent's indent, we're done with this object
		if lineIndent <= parentIndent {
			return obj, currentLine, nil
		}
		
		// Parse key-value pair
		key, value, nextLine, err := parseItem(lines, currentLine, childIndent)
		if err != nil {
			return nil, currentLine, err
		}
		
		if key == "" {
			// End of object
			return obj, nextLine, nil
		}
		
		obj[key] = value
		currentLine = nextLine
	}
	
	return obj, currentLine, nil
}





// parseArray parses an array
func parseArray(lines []string, startLine int, parentIndent int) (Node, int, error) {
	arr := Array{}
	currentLine := startLine + 1
	
	for currentLine < len(lines) {
		// Skip empty lines
		if strings.TrimSpace(lines[currentLine]) == "" {
			currentLine++
			continue
		}
		
		// Check indentation - if it's less than or equal to parent, we're done
		if getIndentation(lines[currentLine]) <= parentIndent {
			break
		}
		
		// Parse array item as primitive value
		line := strings.TrimSpace(lines[currentLine])
		item, err := parsePrimitive(line)
		if err != nil {
			return nil, currentLine, &ParseError{currentLine + 1, err.Error()}
		}
		
		arr = append(arr, item)
		currentLine++
	}
	
	return arr, currentLine, nil
}

// parsePrimitive parses a primitive value
func parsePrimitive(s string) (Node, error) {
	// Trim whitespace
	s = strings.TrimSpace(s)
	
	if s == "" {
		return String(""), nil
	}
	
	// Try boolean
	if s == "true" {
		return Boolean(true), nil
	}
	if s == "false" {
		return Boolean(false), nil
	}
	
	// Try number
	if num, err := strconv.ParseFloat(s, 64); err == nil {
		return Number(num), nil
	}
	
	// Default to string
	return String(s), nil
}

// getIndentation returns the indentation level of a line
func getIndentation(line string) int {
	return len(line) - len(strings.TrimLeft(line, " \t"))
}

// findObjectEnd finds the end of an object by looking for the next line with the same or less indentation than the parent
func findObjectEnd(lines []string, startLine int, parentIndent int) int {
	for i := startLine; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], " \t")
		if line == "" {
			continue
		}
		
		currentIndent := getIndentation(lines[i])
		if currentIndent <= parentIndent {
			return i
		}
	}
	
	return len(lines)
}

// findArrayEnd finds the end of an array by looking for the next line with the same or less indentation than the parent
func findArrayEnd(lines []string, startLine int, parentIndent int) int {
	for i := startLine; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], " \t")
		if line == "" {
			continue
		}
		
		currentIndent := getIndentation(lines[i])
		if currentIndent <= parentIndent {
			return i
		}
	}
	
	return len(lines)
}

// Marshal converts a Go value to TOON format, supporting any Go type including structs
func Marshal(v interface{}) (string, error) {
	var sb strings.Builder
	err := marshalValue(v, &sb, 0)
	return sb.String(), err
}

// marshalValue converts a Go value to TOON format recursively
func marshalValue(v interface{}, sb *strings.Builder, indentLevel int) error {
	indent := strings.Repeat("  ", indentLevel)

	switch val := v.(type) {
	// Handle maps
	case map[string]interface{}:
		for k, v := range val {
			sb.WriteString(indent)
			sb.WriteString(k)
			sb.WriteString(" ")

			// Check if value is a complex type
			switch v.(type) {
			case map[string]interface{}:
				sb.WriteString("\n")
				if err := marshalValue(v, sb, indentLevel+1); err != nil {
					return err
				}
			case []interface{}:
				sb.WriteString("\n")
				if err := marshalValue(v, sb, indentLevel+1); err != nil {
					return err
				}
			default:
				if err := marshalPrimitive(v, sb); err != nil {
					return err
				}
				sb.WriteString("\n")
			}
		}
	
	// Handle slices and arrays
	case []interface{}:
		for _, item := range val {
			sb.WriteString(indent)
			if err := marshalPrimitive(item, sb); err != nil {
				return err
			}
			sb.WriteString("\n")
		}
	
	// Handle string slices specially
	case []string:
		for _, item := range val {
			sb.WriteString(indent)
			sb.WriteString(item)
			sb.WriteString("\n")
		}
	
	// Handle structs using reflection
	default:
		// Check if it's a struct or pointer to struct
		v := reflect.ValueOf(val)
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		
		if v.Kind() == reflect.Struct {
			return marshalStruct(v, sb, indentLevel)
		}
		
		// Check if it's a slice or array
		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			// Handle slices and arrays of any type
			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i).Interface()
				sb.WriteString(indent)
				if err := marshalPrimitive(elem, sb); err != nil {
					return err
				}
				sb.WriteString("\n")
			}
			return nil
		}
		
		// Handle primitive types
		if err := marshalPrimitive(val, sb); err != nil {
			return err
		}
		if indentLevel > 0 {
			sb.WriteString("\n")
		}
	}

	return nil
}

// marshalStruct marshals a struct to TOON format using reflection
func marshalStruct(v reflect.Value, sb *strings.Builder, indentLevel int) error {
	indent := strings.Repeat("  ", indentLevel)
	vType := v.Type()
	
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := vType.Field(i)
		
		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}
		
		// Get field name from tag or use struct field name
		fieldName := fieldType.Tag.Get("toon")
		if fieldName == "" {
			fieldName = fieldType.Name
		}
		
		// Write field name with indent
		sb.WriteString(indent)
		sb.WriteString(fieldName)
		sb.WriteString(" ")
		
		// Marshal field value
		fieldValue := field.Interface()
		fieldKind := field.Kind()
		
		// Check if field is a complex type that needs nested formatting
		isComplex := false
		switch fieldKind {
		case reflect.Map, reflect.Struct, reflect.Slice, reflect.Array:
			isComplex = true
		}
		
		if isComplex {
			// Write newline and indent for nested structures
			sb.WriteString("\n")
			if err := marshalValue(fieldValue, sb, indentLevel+1); err != nil {
				return err
			}
		} else {
			// Write primitive value on the same line
			if err := marshalPrimitive(fieldValue, sb); err != nil {
				return err
			}
			sb.WriteString("\n")
		}
	}
	
	return nil
}

// marshalPrimitive converts a primitive Go value to TOON string
func marshalPrimitive(v interface{}, sb *strings.Builder) error {
	switch val := v.(type) {
	case string:
		sb.WriteString(val)
	case bool:
		if val {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		sb.WriteString(fmt.Sprintf("%d", val))
	case float32, float64:
		sb.WriteString(fmt.Sprintf("%g", val))
	// Handle slices and arrays of primitive types
	case []string:
		for i, s := range val {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(s)
		}
	case []int, []int8, []int16, []int32, []int64, []uint, []uint8, []uint16, []uint32, []uint64:
		// Convert to []interface{} and marshal
		slice := reflect.ValueOf(val)
		for i := 0; i < slice.Len(); i++ {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("%v", slice.Index(i).Interface()))
		}
	case []bool:
		// Convert to []interface{} and marshal
		slice := reflect.ValueOf(val)
		for i := 0; i < slice.Len(); i++ {
			if i > 0 {
				sb.WriteString(" ")
			}
			if slice.Index(i).Bool() {
				sb.WriteString("true")
			} else {
				sb.WriteString("false")
			}
		}
	// Handle nil values
	case nil:
		sb.WriteString("")
	default:
		// Try to convert to string
		sb.WriteString(fmt.Sprintf("%v", val))
	}
	return nil
}

// Unmarshal converts TOON string to a Go value, supporting any Go type including structs
func Unmarshal(data string, v interface{}) error {
	node, err := Parse(data)
	if err != nil {
		return err
	}

	// Convert Node to interface{}
	value := node.Value()

	// Get reflect.Value of the target
	target := reflect.ValueOf(v)
	if target.Kind() != reflect.Ptr {
		return errors.New("target must be a pointer")
	}

	// Dereference the pointer to get the actual target value
	target = target.Elem()

	// Assign the parsed value to the target
	return assignValue(target, value)
}

// assignValue assigns a value to a reflect.Value, supporting structs, maps, slices, and primitives
func assignValue(target reflect.Value, value interface{}) error {
	// Get the type of the target
	targetType := target.Type()

	// Check if target is nil and can be set (only for pointers, maps, slices, channels, functions, interfaces)
	if (target.Kind() == reflect.Ptr || target.Kind() == reflect.Map || target.Kind() == reflect.Slice || 
		target.Kind() == reflect.Chan || target.Kind() == reflect.Func || target.Kind() == reflect.Interface) && 
		target.IsNil() && target.CanSet() {
		// Create a new value of the target type
		switch target.Kind() {
		case reflect.Ptr:
			// Create a new instance of the pointer type
			target.Set(reflect.New(targetType.Elem()))
		case reflect.Map:
			// Create a new map
			target.Set(reflect.MakeMap(targetType))
		case reflect.Slice:
			// Create a new slice with length 0
			target.Set(reflect.MakeSlice(targetType, 0, 0))
		}
	}

	// Dereference pointers
	if target.Kind() == reflect.Ptr {
		if target.IsNil() {
			// Create a new instance of the pointer type
			target.Set(reflect.New(targetType.Elem()))
		}
		return assignValue(target.Elem(), value)
	}

	// Handle different target types
	switch target.Kind() {
	case reflect.Interface:
		// For interfaces, just assign the value directly
		target.Set(reflect.ValueOf(value))

	case reflect.Struct:
		// Assign to struct fields
		return assignToStruct(target, value)

	case reflect.Map:
		// Assign to map
		return assignToMap(target, value)

	case reflect.Slice, reflect.Array:
		// Assign to slice or array
		return assignToSlice(target, value)

	default:
		// Assign primitive types
		return assignPrimitive(target, value)
	}

	return nil
}

// assignToStruct assigns a map to a struct
func assignToStruct(target reflect.Value, value interface{}) error {
	// Convert value to map[string]interface{}
	valueMap, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("cannot assign %T to struct %s", value, target.Type().Name())
	}

	// Iterate over struct fields
	targetType := target.Type()
	for i := 0; i < target.NumField(); i++ {
		field := target.Field(i)
		fieldType := targetType.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get field name from tag or use struct field name
		fieldName := fieldType.Tag.Get("toon")
		if fieldName == "" {
			fieldName = fieldType.Name
		}

		// Check if the field exists in the map
		if val, exists := valueMap[fieldName]; exists {
			// Assign the value to the field
			if err := assignValue(field, val); err != nil {
				return fmt.Errorf("error assigning to field %s: %w", fieldName, err)
			}
		}
	}

	return nil
}

// assignToMap assigns a map to another map
func assignToMap(target reflect.Value, value interface{}) error {
	// Convert value to map[string]interface{}
	valueMap, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("cannot assign %T to map %s", value, target.Type().Name())
	}

	// Check if map is nil and can be set
	if target.IsNil() {
		// Create a new map
		newMap := reflect.MakeMap(target.Type())
		target.Set(newMap)
	}

	// Iterate over the value map and assign to target map
	for key, val := range valueMap {
		// Create key and value reflect.Value
		keyValue := reflect.ValueOf(key)
		valueValue := reflect.ValueOf(val)

		// Check if key type matches
		if keyValue.Type() != target.Type().Key() {
			continue // Skip incompatible keys
		}

		// Assign value to map
		target.SetMapIndex(keyValue, valueValue)
	}

	return nil
}

// assignToSlice assigns a slice to another slice
func assignToSlice(target reflect.Value, value interface{}) error {
	// Convert value to []interface{}
	valueSlice, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("cannot assign %T to slice %s", value, target.Type().Name())
	}

	// Create a new slice with the same length as the value slice
	newSlice := reflect.MakeSlice(target.Type(), len(valueSlice), len(valueSlice))

	// Iterate over the value slice and assign to target slice
	for i, val := range valueSlice {
		// Assign value to slice element
		elem := newSlice.Index(i)
		if err := assignPrimitive(elem, val); err != nil {
			return fmt.Errorf("error assigning to slice element %d: %w", i, err)
		}
	}

	// Set the new slice to the target
	target.Set(newSlice)

	return nil
}

// assignPrimitive assigns a primitive value to a reflect.Value
func assignPrimitive(target reflect.Value, value interface{}) error {
	targetType := target.Type()

	// Handle nil values
	if value == nil {
		return nil
	}

	// Special case: convert numeric values to string when target is string
	if targetType.Kind() == reflect.String {
		// Convert any value to string
		target.SetString(fmt.Sprintf("%v", value))
		return nil
	}

	// Special case: convert string values to numeric when target is numeric
	targetKind := targetType.Kind()
	if targetKind >= reflect.Int && targetKind <= reflect.Float64 {
		if strVal, ok := value.(string); ok {
			// Try to convert string to numeric type
			switch targetKind {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if num, err := strconv.ParseInt(strVal, 10, 64); err == nil {
					target.SetInt(num)
					return nil
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if num, err := strconv.ParseUint(strVal, 10, 64); err == nil {
					target.SetUint(num)
					return nil
				}
			case reflect.Float32, reflect.Float64:
				if num, err := strconv.ParseFloat(strVal, 64); err == nil {
					target.SetFloat(num)
					return nil
				}
			}
		}
	}

	// Get the value as reflect.Value
	valueValue := reflect.ValueOf(value)
	valueType := valueValue.Type()

	// Check if types match
	if valueType == targetType {
		target.Set(valueValue)
		return nil
	}

	// Try to convert the value to the target type
	if valueValue.CanConvert(targetType) {
		target.Set(valueValue.Convert(targetType))
		return nil
	}

	// Last resort: try to convert using string representation
	strVal := fmt.Sprintf("%v", value)
	if targetType.Kind() == reflect.Bool {
		if boolVal, err := strconv.ParseBool(strVal); err == nil {
			target.SetBool(boolVal)
			return nil
		}
	}

	return fmt.Errorf("cannot assign %T to %s", value, targetType.Name())
}
