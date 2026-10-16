// note: make sure this file is before version.go in lexical order
// to make init() in version.go correctly validate against updated variable here

//go:build !release

package version

func init() {
	Release = NewVersion(Release.Major, Release.Minor, Release.Patch, "dev")
}
