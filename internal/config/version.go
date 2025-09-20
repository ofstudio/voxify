package config

var version string

// Version returns the version of the build.
// If the version is not set, it returns "dev".
//
// The version is set by the build system:
//
//	go build \
//	  -ldflags "-X '${MODULE}/internal/config.version=${BUILD_VERSION}'" \
//	  -o ${DEST} ${SRC}
func Version() string {
	if version == "" {
		return "dev"
	}
	return version
}
