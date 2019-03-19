package main_test

import (
	"testing"

	"gopkg.in/bblfsh/sdk.v2/build"
)

func TestSDKUpToDate(t *testing.T) {
	printf := func(format string, args ...interface{}) (int, error) {
		t.Logf(format, args...)
		return 0, nil
	}
	err := build.UpdateSDK("../", &build.UpdateOptions{
		DryRun:  true,
		Debug:   printf,
		Notice:  printf,
		Warning: printf,
	})
	if err != nil {
		t.Fatal(err)
	}
}
