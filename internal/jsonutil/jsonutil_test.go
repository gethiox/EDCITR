package jsonutil

import (
	"encoding/json/v2"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testFuncs = []struct {
	f    func(data []byte, v any) error
	name string
}{
	{
		f:    UnmarshalStrict,
		name: "UnmarshalStrict",
	}, {
		f:    UnmarshalStrictV3,
		name: "UnmarshalStrictV3",
	},
}

func TestUnmarshal(t *testing.T) {
	type SubStruct struct {
		Field  string `json:"field"`
		FieldR string `json:"field_r,required"`
	}

	type MyStruct struct {
		Field      string    `json:"field"`
		FieldR     string    `json:"field_r,required"`
		SubStruct  SubStruct `json:"sub"`
		SubStructR SubStruct `json:"sub_r,required"`
	}

	var registry = map[string]reflect.Type{
		"my_struct": reflect.TypeFor[MyStruct](),
	}

	type TestCase struct {
		name      string
		inputData string
		expected  any
		expectErr bool
	}

	testCases := []TestCase{
		{
			name:      "all fields populated",
			inputData: `{"field": "1", "field_r": "2", "sub": {"field": "3", "field_r": "4"}, "sub_r": {"field": "5", "field_r": "6"}}`,
			expected: MyStruct{
				Field:      "1",
				FieldR:     "2",
				SubStruct:  SubStruct{Field: "3", FieldR: "4"},
				SubStructR: SubStruct{Field: "5", FieldR: "6"},
			},
			expectErr: false,
		},
		{
			name:      "missing top-level optional field",
			inputData: `{"field_r": "2", "sub": {"field": "3", "field_r": "4"}, "sub_r": {"field": "5", "field_r": "6"}}`,
			expected: MyStruct{
				Field:      "", // Should be zero-value
				FieldR:     "2",
				SubStruct:  SubStruct{Field: "3", FieldR: "4"},
				SubStructR: SubStruct{Field: "5", FieldR: "6"},
			},
			expectErr: false,
		},
		{
			name:      "missing top-level required field",
			inputData: `{"field": "1", "sub": {"field": "3", "field_r": "4"}, "sub_r": {"field": "5", "field_r": "6"}}`,
			expected:  MyStruct{}, // Irrelevant if it errors
			expectErr: true,
		},
		{
			name: "missing optional struct entirely",
			// 'sub' is omitted. Because 'sub' itself isn't required, this should pass,
			// even though 'sub' has an internal field that is required.
			inputData: `{"field": "1", "field_r": "2", "sub_r": {"field": "5", "field_r": "6"}}`,
			expected: MyStruct{
				Field:      "1",
				FieldR:     "2",
				SubStruct:  SubStruct{}, // Remains zero-value
				SubStructR: SubStruct{Field: "5", FieldR: "6"},
			},
			expectErr: false,
		},
		{
			name:      "missing required struct entirely",
			inputData: `{"field": "1", "field_r": "2", "sub": {"field": "3", "field_r": "4"}}`,
			expected:  MyStruct{},
			expectErr: true,
		},
		{
			name: "optional struct present but missing its required field",
			// 'sub' is present, meaning the parser enters it. It is missing 'field_r'.
			inputData: `{"field": "1", "field_r": "2", "sub": {"field": "3"}, "sub_r": {"field": "5", "field_r": "6"}}`,
			expected:  MyStruct{},
			expectErr: true,
		},
		{
			name:      "required struct present but missing its required field",
			inputData: `{"field": "1", "field_r": "2", "sub": {"field": "3", "field_r": "4"}, "sub_r": {"field": "5"}}`,
			expected:  MyStruct{},
			expectErr: true,
		},
		{
			name: "required field present but empty string",
			// An empty string ("") means the key was present in the JSON.
			// A good 'required' implementation should PASS this unless you also validate for non-zero length.
			inputData: `{"field": "1", "field_r": "", "sub": {"field": "3", "field_r": "4"}, "sub_r": {"field": "5", "field_r": "6"}}`,
			expected: MyStruct{
				Field:      "1",
				FieldR:     "",
				SubStruct:  SubStruct{Field: "3", FieldR: "4"},
				SubStructR: SubStruct{Field: "5", FieldR: "6"},
			},
			expectErr: false,
		},
		{
			name:      "empty JSON object",
			inputData: `{}`,
			expected:  MyStruct{},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		for _, fn := range testFuncs {
			t.Run(fmt.Sprintf("%s_%s", tc.name, fn.name), func(t *testing.T) {
				t.Run("direct", func(t *testing.T) {
					var actual MyStruct
					err := fn.f([]byte(tc.inputData), &actual)

					if tc.expectErr {
						if err == nil {
							t.Errorf("expect err but got nil")
						}
					} else {
						if err != nil {
							t.Errorf("expect no error but got %v", err)
						}
						assert.Equal(t, tc.expected, actual)
					}
				})

				t.Run("indirect", func(t *testing.T) {
					evType, ok := registry["my_struct"]
					if !ok {
						t.Fatalf("no event type registered")
					}
					targetPtr := reflect.New(evType).Interface()

					err := fn.f([]byte(tc.inputData), targetPtr)

					if tc.expectErr {
						if err == nil {
							t.Errorf("expect err but got nil")
						}
					} else {
						if err != nil {
							t.Errorf("expect no error but got %v", err)
						}
						v, ok := targetPtr.(*MyStruct)
						if !ok {
							t.Fatalf("expected MyStruct, got %T", targetPtr)
						}
						assert.Equal(t, tc.expected, *v)
					}
				})
			})
		}
	}
}

func Benchmark(b *testing.B) {
	type MyStruct struct {
		MyField         string `json:"my_field"`
		MyRequiredField string `json:"my_required_field,required"`
	}

	data := `{"my_field": "hello 1", "my_required_field": "hello 2"}`

	for _, fn := range testFuncs {
		b.Run(fmt.Sprintf("%s", fn.name), func(b *testing.B) {
			var v MyStruct
			for i := 0; i < b.N; i++ {
				fn.f([]byte(data), &v)
			}
		})
	}

	b.Run("json.Unmarshal", func(b *testing.B) {
		var v MyStruct
		for i := 0; i < b.N; i++ {
			json.Unmarshal([]byte(data), &v)
		}
	})
}
