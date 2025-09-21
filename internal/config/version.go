package config

var version string

const VersionDev = "dev"

// Version returns the version of the build.
// If the version is not set, it returns VersionDev
//
// The version is set by the build system:
//
//	go build \
//	  -ldflags "-X '${MODULE}/internal/config.version=${BUILD_VERSION}'" \
//	  -o ${DEST} ${SRC}
func Version() string {
	if version == "" {
		return VersionDev
	}
	return version
}
