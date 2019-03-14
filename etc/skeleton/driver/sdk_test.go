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
	err := build.SDKUpdate("../", &build.UpdateOptions{
		DryRun:   true,
		Debugf:   printf,
		Noticef:  printf,
		Warningf: printf,
	})
	if err != nil {
		t.Fatal(err)
	}
}
