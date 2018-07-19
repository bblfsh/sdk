package build

import "os"

func isCI() bool {
	return os.Getenv("CI") != "" ||
		os.Getenv("CONTINUOUS_INTEGRATION") != "" ||
		os.Getenv("TRAVIS") != ""
}

func ciBranch() string {
	return os.Getenv("TRAVIS_BRANCH")
}

func ciTag() string {
	return os.Getenv("TRAVIS_TAG")
}

func ciIsPR() bool {
	pr := os.Getenv("TRAVIS_PULL_REQUEST")
	return pr != "false" && pr != ""
}
