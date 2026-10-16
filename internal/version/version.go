package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func init() {
	if !Release.valid {
		panic(fmt.Sprintf("release version is not valid: %#v", Release))
	}
}

type Version struct {
	Major, Minor, Patch int
	Identifier          string

	valid bool
}

// dirty, non-complaint validator
var identifierValidator = regexp.MustCompile("^[a-zA-Z0-9-]+$")

func NewVersion(major, minor, patch int, identifier string) Version {
	var identifierValid bool
	var valid bool

	if identifier != "" {
		identifierValid = identifierValidator.MatchString(identifier)
	} else {
		identifierValid = true
	}

	if major >= 0 && minor >= 0 && patch >= 0 && identifierValid {
		valid = true
	}

	return Version{
		Major: major, Minor: minor, Patch: patch,
		Identifier: identifier,
		valid:      valid,
	}
}

func NewVersionFromString(version string) (Version, error) {
	trimmed := strings.TrimPrefix(version, "v")

	var identifier string

	if strings.Contains(trimmed, "-") {
		parts := strings.SplitN(trimmed, "-", 2)

		trimmed = parts[0]
		identifier = parts[1]
	}

	parts := strings.Split(trimmed, ".")

	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version string: %s", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major value: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor value: %s", parts[1])
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch value: %s", parts[2])
	}

	return NewVersion(major, minor, patch, identifier), nil
}

func (v Version) IsDev() bool {
	return v.Identifier == "dev"
}

func (v Version) IsValid() bool {
	return v.valid
}

func (v Version) FormatMMP() string {
	if !v.valid {
		return "0.0.0"
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v Version) FormatFull() string {
	if !v.valid {
		return "v0.0.0"
	}
	if v.Identifier != "" {
		return fmt.Sprintf("v%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.Identifier)
	}
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v Version) String() string {
	return v.FormatFull()
}

//goland:noinspection GoMixedReceiverTypes
func (v *Version) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, v.FormatFull())), nil
}

//goland:noinspection GoMixedReceiverTypes
func (v *Version) UnmarshalJSON(val []byte) error {
	s := string(val)

	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return fmt.Errorf("invalid value: %s", s)
	}

	parsed, err := NewVersionFromString(s[1 : len(s)-1])
	if err != nil {
		return err
	}

	*v = parsed

	return nil
}

func (v Version) GreaterThan(other Version) bool {
	switch {
	case v.Major > other.Major:
		return true
	case v.Major < other.Major:
		return false
	case v.Minor > other.Minor:
		return true
	case v.Minor < other.Minor:
		return false
	case v.Patch > other.Patch:
		return true
	case v.Patch < other.Patch:
		return false
	default:
		return false
	}
}
