// Package version contains the version for the whole of Qvain and associated commands.
package version

// Set an appropriate SemVer for "official releases".
// CommitHash is set by the build process as a linker flag. Do not edit.
var (
	SemVer     = "0.0.1"
	CommitTag  = "v???"
	CommitHash = "(unknown)"
)

// github link to hash:
//   https://github.com/wvh/helloworld/commit/76f4ee6123c9584d77a37c783b5fc4addafe14f2
// github link to release:
//   https://github.com/wvh/helloworld/releases/tag/v1.0
