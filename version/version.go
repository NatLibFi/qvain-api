// Package version contains the version for the whole of Qvain and associated commands.
package version

// Set an appropriate SemVer for "official releases".
// CommitHash is set by the build process as a linker flag. Do not edit.
var (

	// ident for internal use
	Id = "qvain"

	// program name for end user
	Name = "Qvain"

	// program description
	Description = "Qvain API"

	// semver string – set this manually
	SemVer = "0.0.1"

	// nearest commit tag – set by linker
	CommitTag = "v???"

	// exact commit hash – set by linker
	CommitHash = "(unknown)"

	// branch – set by linker
	CommitBranch = "master"
)
