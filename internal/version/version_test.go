package version

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {

	type testCase struct {
		input       string
		expectedErr bool
		expected    Version
	}

	for i, tc := range []testCase{
		// numbers only
		{
			input: "v0.0.0", expected: Version{
				Major: 0, Minor: 0, Patch: 0,
				Identifier: "",
				valid:      true,
			},
		},
		{
			input: "v0.0.1", expected: Version{
				Major: 0, Minor: 0, Patch: 1,
				Identifier: "",
				valid:      true,
			},
		},
		{
			input: "v0.1.0", expected: Version{
				Major: 0, Minor: 1, Patch: 0,
				Identifier: "",
				valid:      true,
			},
		},
		{
			input: "v0.1.1", expected: Version{
				Major: 0, Minor: 1, Patch: 1,
				Identifier: "",
				valid:      true,
			},
		},
		{
			input: "v1.0.0", expected: Version{
				Major: 1, Minor: 0, Patch: 0,
				Identifier: "",
				valid:      true,
			},
		},
		{
			input: "v1.0.1", expected: Version{
				Major: 1, Minor: 0, Patch: 1,
				Identifier: "",
				valid:      true,
			},
		},
		{
			input: "v1.1.0", expected: Version{
				Major: 1, Minor: 1, Patch: 0,
				Identifier: "",
				valid:      true,
			},
		},
		{
			input: "v1.1.1", expected: Version{
				Major: 1, Minor: 1, Patch: 1,
				Identifier: "",
				valid:      true,
			},
		},
		// without "v"
		{
			input: "1.1.1", expected: Version{
				Major: 1, Minor: 1, Patch: 1,
				Identifier: "",
				valid:      true,
			},
		},
		// with ident
		{
			input: "v1.1.1-ident", expected: Version{
				Major: 1, Minor: 1, Patch: 1,
				Identifier: "ident",
				valid:      true,
			},
		},
		// double hypen
		{
			input: "v1.1.1-ident-ident", expected: Version{
				Major: 1, Minor: 1, Patch: 1,
				Identifier: "ident-ident",
				valid:      true,
			},
		},
		// negative numbers
		{input: "1.1.-1", expectedErr: true},
		{input: "1.-1.1", expectedErr: true},
		{input: "1.-1.-1", expectedErr: true},
		{input: "-1.1.1", expectedErr: true},
		{input: "-1.-1.1", expectedErr: true},
		{input: "-1.-1.-1", expectedErr: true},
	} {
		t.Run(fmt.Sprintf("%d_%s", i, tc.input), func(t *testing.T) {
			v, err := NewVersionFromString(tc.input)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, v)
			}
		})
	}

}
