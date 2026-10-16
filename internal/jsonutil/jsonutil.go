// JSON unmarshal wrapper that raises an error on required fields
// required field is being marked by the user in a JSON "required" tag, example:
//
// type MyStruct struct {
//     MyField         string `json:"my_field"`
//     MyRequiredField string `json:"my_required_field,required"`
// }

package jsonutil

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/gethiox/EDCITR/internal/utils"
)

type MissingRequiredFieldError struct {
	Field string
}

func (m MissingRequiredFieldError) Error() string {
	return fmt.Sprintf("missing required field: %q", m.Field)
}

// structMetadata holds pre-calculated info to avoid reflection in the hot path
type structMetadata struct {
	hasRequired    bool
	requiredFields []string
	nested         map[string]*structMetadata
	uniquePaths    [][]int
}

var metadataCache sync.Map // map[reflect.Type]*structMetadata

func getMetadata(t reflect.Type) *structMetadata {
	if val, ok := metadataCache.Load(t); ok {
		return val.(*structMetadata)
	}

	meta := &structMetadata{
		nested: make(map[string]*structMetadata),
	}
	buildMetadata(t, meta, nil)
	metadataCache.Store(t, meta)
	return meta
}

func buildMetadata(t reflect.Type, meta *structMetadata, currentPath []int) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "-" {
			continue
		}

		parts := strings.Split(tag, ",")
		jsonName := parts[0]

		// Track deep paths for unique string interning (handles embeds)
		path := append([]int(nil), currentPath...)
		path = append(path, i)

		if field.Anonymous && jsonName == "" {
			buildMetadata(field.Type, meta, path)
			continue
		}

		if jsonName == "" {
			jsonName = field.Name
		}

		isRequired, isUnique := false, false
		for _, opt := range parts[1:] {
			if opt == "required" {
				isRequired = true
			}
			if opt == "unique" {
				isUnique = true
			}
		}

		if isRequired {
			meta.requiredFields = append(meta.requiredFields, jsonName)
			meta.hasRequired = true
		}

		if isUnique && field.Type.Kind() == reflect.String {
			meta.uniquePaths = append(meta.uniquePaths, path)
		}

		// Recurse for nested objects
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct {
			nestedMeta := &structMetadata{nested: make(map[string]*structMetadata)}
			buildMetadata(fieldType, nestedMeta, nil) // Nested tokens start a fresh local path
			if nestedMeta.hasRequired {
				meta.nested[jsonName] = nestedMeta
				meta.hasRequired = true
			}
		}
	}
}

func UnmarshalStrictV3(data []byte, v any) error {
	//if err := json.Unmarshal(data, v); err != nil {
	if err := json.Unmarshal(data, v, jsontext.AllowDuplicateNames(true)); err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("v must be a non-nil pointer")
	}

	elem := rv.Elem()
	meta := getMetadata(elem.Type())

	// Fast string interning using paths (supports embedded structs natively)
	for _, path := range meta.uniquePaths {
		f := elem.FieldByIndex(path)
		s := f.String()
		if s != "" {
			f.SetString(utils.UniqueString(s))
		}
	}

	if !meta.hasRequired {
		return nil
	}

	dec := jsontext.NewDecoder(bytes.NewReader(data))
	return validateTokens(dec, meta)
}

func validateTokens(dec *jsontext.Decoder, meta *structMetadata) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	if tok.Kind() != '{' {
		return nil // Not an object, ignore
	}

	// Use a slice instead of uint64 mask to avoid 64-field limit
	seen := make([]bool, len(meta.requiredFields))
	seenCount := 0

	for dec.PeekKind() != '}' && dec.PeekKind() != 0 {
		keyTok, err := dec.ReadToken()
		if err != nil {
			return err
		}
		key := keyTok.String()

		// Check required fields at current depth
		for i, req := range meta.requiredFields {
			if req == key && !seen[i] {
				seen[i] = true
				seenCount++
				break
			}
		}

		// Recurse or skip
		if nestedMeta, ok := meta.nested[key]; ok {
			if err := validateTokens(dec, nestedMeta); err != nil {
				var e MissingRequiredFieldError
				if errors.As(err, &e) {
					return MissingRequiredFieldError{Field: key + "." + e.Field}
				}
				return err
			}
		} else {
			if err := dec.SkipValue(); err != nil {
				return err
			}
		}
	}

	dec.ReadToken() // Consume '}'

	if seenCount < len(meta.requiredFields) {
		for i, req := range meta.requiredFields {
			if !seen[i] {
				return MissingRequiredFieldError{Field: req}
			}
		}
	}
	return nil
}

// UnmarshalStrict unmarshals JSON into v, handling both direct structs and
// interfaces that might hold a Type or a Struct instance.
func UnmarshalStrict(data []byte, v any) error {
	rv := reflect.ValueOf(v)

	// Validate we got a pointer
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("v must be a non-nil pointer, got %T", v)
	}

	// --- 1. Indirect Case: Interface holding a Value OR a Type ---
	// We check if the pointer 'v' points to an Interface (e.g., var actual any)
	if rv.Elem().Kind() == reflect.Interface {
		elem := rv.Elem() // This is the interface variable (e.g., actual)
		if !elem.IsNil() {
			inner := elem.Elem() // This is the value INSIDE the interface

			var targetType reflect.Type

			// Check A: Does the interface hold a reflect.Type?
			// (e.g., registry["key"] which returns reflect.Type)
			if t, ok := inner.Interface().(reflect.Type); ok {
				targetType = t
			} else if inner.Kind() == reflect.Struct {
				// Check B: Does the interface hold a Struct Value?
				// (e.g., MyStruct{})
				targetType = inner.Type()
			}

			// If we successfully identified a struct type to unmarshal into
			if targetType != nil {
				// Handle Pointers (e.g., reflect.TypeOf(&MyStruct{}))
				if targetType.Kind() == reflect.Ptr {
					targetType = targetType.Elem()
				}

				// Ensure it's a struct
				if targetType.Kind() == reflect.Struct {
					// 1. Create a new pointer to the concrete struct type
					// newPtr is *MyStruct
					newPtr := reflect.New(targetType)

					// 2. Unmarshal JSON into this new pointer
					if err := json.Unmarshal(data, newPtr.Interface()); err != nil {
						return err
					}

					// 3. Strict Validation
					if err := validateStruct(newPtr.Elem(), data); err != nil {
						return err
					}

					// 4. Update the original interface with the concrete STRUCT VALUE
					// actual becomes MyStruct{...} instead of reflect.Type
					elem.Set(newPtr.Elem())
					return nil
				}
			}
		}
	}

	// --- 2. Direct Case: Standard pointer to struct ---
	// Handles: UnmarshalStrict(..., &myStructInstance)
	if err := json.Unmarshal(data, v); err != nil {
		return err
	}

	// Validate the element pointed to by v
	return validateStruct(rv.Elem(), data)
}

func validateStruct(v reflect.Value, data []byte) error {
	// Parse JSON into map to check for existence of keys
	var rawMap map[string]any
	if err := json.Unmarshal(data, &rawMap); err != nil {
		// If data isn't a JSON object, we skip strict validation
		return nil
	}

	return validateRecursive(v, rawMap)
}

func validateRecursive(val reflect.Value, rawMap map[string]any) error {
	// Dereference pointers/interfaces to get to the struct
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		if val.Kind() == reflect.Ptr && val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		tag := field.Tag.Get("json")
		if tag == "-" {
			continue
		}

		parts := strings.Split(tag, ",")
		jsonName := parts[0]

		// Handle Embedded Structs (Anonymous fields)
		if field.Anonymous {
			// If inlined (empty json name), recurse using the same map
			if jsonName == "" {
				if err := validateRecursive(fieldVal, rawMap); err != nil {
					return err
				}
				continue
			}
		}

		if jsonName == "" {
			jsonName = field.Name
		}

		// Check "required"
		isRequired := false
		for _, opt := range parts[1:] {
			if opt == "required" {
				isRequired = true
				break
			}
		}

		// Check existence in map
		rawValue, exists := rawMap[jsonName]
		if isRequired && !exists {
			return MissingRequiredFieldError{Field: field.Name}
		}

		// Recurse for Nested Structs (NOT embedded/anonymous)
		if exists && (fieldVal.Kind() == reflect.Struct || (fieldVal.Kind() == reflect.Ptr && fieldVal.Elem().Kind() == reflect.Struct)) {
			if nestedMap, ok := rawValue.(map[string]any); ok {
				err := validateRecursive(fieldVal, nestedMap)
				if err != nil {
					var e MissingRequiredFieldError
					if errors.As(err, &e) {
						return MissingRequiredFieldError{Field: fmt.Sprintf("%s.%s", jsonName, e.Field)}
					}
					return err
				}
			}
		}
	}
	return nil
}
