package buildmanifest

import "gopkg.in/yaml.v2"

const (
	// Filename of a build manifest.
	Filename = "build.yml"
	// CurrentVersion os the manifest file format. See Manifest.SDK.
	CurrentVersion = "2"
)

// Artifact is a file copied from one path to another during the build.
type Artifact struct {
	Path string `yaml:"path"`
	Dest string `yaml:"dest"`
}

// Manifest used for the declarative build system for drivers.
type Manifest struct {
	// SDK version. Used to track future changes to this file. Current version is '2'.
	SDK string `yaml:"sdk"`

	Native struct {
		// Image is a Docker image used as the native driver runtime.
		Image string `yaml:"image"`
		// Static is a list of files that will be copied from native driver's source directory
		// to the final driver image. Note that those files cannot be modified by the build.
		Static []Artifact `yaml:"static"`
		// Deps is a list of apt/apk commands executed in the final driver image.
		// This directive should be avoided since a different Docker image may be used instead of it.
		Deps  []string `yaml:"deps"`
		Build struct {
			// Gopath directive sets a new gopath for the driver build. Only used by Go drivers.
			Gopath string `yaml:"gopath"`
			// Image is a Docker image used to build the native driver.
			Image string `yaml:"image"`
			// Deps is a list of shell commands to pull native driver dependencies.
			// Note that those commands are executed before copying the driver files to the
			// container, so they can be cached. See Run also.
			Deps []string `yaml:"deps"`
			// Add is a list of files that are copied from the native driver source to the
			// native build image.
			Add []Artifact `yaml:"add"`
			// Run is a list of shell commands to build the native driver. Those commands
			// can access files copied by Add directives. Files produced by the build should
			// be mentioned in the Artifacts directive to be copied to the final image.
			Run []string `yaml:"run"`
			// Artifacts is a list of files copied from the native build image to the final
			// driver image.
			Artifacts []Artifact `yaml:"artifacts"`
		} `yaml:"build"`
		Test struct {
			Deps []string `yaml:"deps"`
			Run  []string `yaml:"run"`
		} `yaml:"test"`
	} `yaml:"native"`
	Runtime struct {
		// Version of Go used to build the driver server.
		Version string `yaml:"version"`
	} `yaml:"go-runtime"`
}

// Decode the manifest file.
func (m *Manifest) Decode(data []byte) error {
	return yaml.Unmarshal(data, m)
}
