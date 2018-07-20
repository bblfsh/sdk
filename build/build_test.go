package build

import "testing"

func TestIsApkOrApt(t *testing.T) {
	for _, c := range []struct {
		cmd string
		exp bool
	}{
		{`echo`, false},
		{`apk add xxx`, true},
		{`apt install xxx`, true},
		{`apt-get install xxx`, true},
		{`apt update && apt install xxx`, true},
		{`apt update && echo`, false},
		{`apt update || echo`, false},
		{`apt update ; echo`, false},
		{`apt install x || apt install y`, true},
		{`apt install x; apt install y`, true},
	} {
		t.Run(c.cmd, func(t *testing.T) {
			if c.exp != isApkOrApt(c.cmd) {
				t.Fatal("should be", c.exp)
			}
		})
	}
}
