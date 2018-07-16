package fixtures

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const envLocalTest = "BBLFSH_TEST_LOCAL"

var runInDocker = os.Getenv(envLocalTest) != "true"

const (
	dockerBinary = "docker"
	dockerSocket = "/var/run/docker.sock"
)

func checkDockerInstalled(t testing.TB) {
	path, err := exec.LookPath(dockerBinary)
	if err == nil {
		t.Logf("found docker binary: %s", path)
		return
	}
	t.Errorf("cannot find docker: %s", err)
	t.Logf("BBLFSH_TEST_LOCAL: %s", os.Getenv(envLocalTest))
	t.Logf("PATH: %q", filepath.SplitList(os.Getenv("PATH")))
	if _, err := os.Stat(dockerSocket); err != nil {
		t.Errorf("docker socket is not available: %v", err)
	} else {
		t.Logf("docker socket is available!")
	}
	t.FailNow()
}

func (s *Suite) runTestsDocker(t *testing.T) {
	checkDockerInstalled(t)

	pkgDir, err := os.Getwd()
	require.NoError(t, err)
	root := filepath.Join(pkgDir, "../..")

	t.Logf("Running tests in Docker")

	dir, err := ioutil.TempDir("", "fixtures-test-")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	const bin = "fixtures.test"

	compileTest(t, "./", filepath.Join(dir, bin))

	t.Log("Building test container...")
	s.genDockerfile(t, dir)
	imageid := dockerBuild(t, dir)
	t.Log(imageid)

	t.Log("Running tests...")
	testOut := dockerRunFixtures(t, root, imageid, s.Docker.Debug)
	reconstructTestLog(t, testOut)
}

const dockerFile = `FROM %s

WORKDIR /test/driver/fixtures

ADD fixtures.test ./

VOLUME /test/fixtures
VOLUME /test/build
VOLUME /test/native

ENV ` + envLocalTest + `=true

ENTRYPOINT ./fixtures.test -test.v
`

func (s *Suite) genDockerfile(t testing.TB, dir string) {
	err := ioutil.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(fmt.Sprintf(dockerFile, s.Docker.Image)), 0644)
	require.NoError(t, err)
}

func compileTest(t testing.TB, path, dst string) {
	out, err := exec.Command("go", "test", "-c", "-o", dst, path).CombinedOutput()
	require.NoError(t, err, "failed to compile tests; output:\n%s", out)
}

func dockerBuild(t testing.TB, dir string) string {
	outBuf := bytes.NewBuffer(nil)
	errBuf := bytes.NewBuffer(nil)
	cmd := exec.Command(dockerBinary, "build", "-q", dir)
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf
	err := cmd.Run()
	require.NoError(t, err, "%s", errBuf)

	imageid := string(bytes.TrimSpace(outBuf.Bytes()))
	if imageid == "" {
		require.Fail(t, "failed to get image id", "output:\n%s\nerror:\n%s", outBuf, errBuf)
	}
	return imageid
}

func dockerRunFixtures(t testing.TB, root, image string, debug bool) io.Reader {
	outBuf := bytes.NewBuffer(nil)
	errBuf := bytes.NewBuffer(nil)
	errc := make(chan error, 1)
	pr, pw := io.Pipe()
	go func() {
		defer pr.Close()
		cmd := exec.Command("go", "tool", "test2json")
		cmd.Stdout = outBuf
		cmd.Stdin = pr
		errc <- cmd.Run()
	}()

	args := []string{
		"run", "--rm",
	}
	for _, d := range []string{
		"fixtures",
		"native",
		"build",
	} {
		args = append(args,
			"-v", filepath.Join(root, d)+":/test/"+d,
		)
	}
	args = append(args, image)

	outWriter := func(w io.Writer) io.Writer {
		return w
	}

	if debug {
		logf, err := os.Create("docker_test.log")
		require.NoError(t, err)
		defer logf.Close()

		outWriter = func(w io.Writer) io.Writer {
			return io.MultiWriter(w, logf)
		}
	}

	t.Log(strings.Join(append([]string{dockerBinary}, args...), " "))
	cmd := exec.Command(dockerBinary, args...)
	cmd.Stdout = outWriter(pw)
	cmd.Stderr = outWriter(errBuf)
	err := cmd.Run()
	pw.Close()
	if err != nil {
		t.Error(errBuf.String())
	}
	if err = <-errc; err != nil {
		t.Error(err)
	}
	return outBuf
}

func gopath(t testing.TB) []string {
	paths := os.Getenv("GOPATH")
	if paths == "" {
		data, err := exec.Command("go", "env", "GOPATH").Output()
		require.NoError(t, err)
		paths = strings.TrimSpace(string(data))
	}
	return filepath.SplitList(paths)
}

func findPackage(t testing.TB, dir string) string {
	for _, pref := range gopath(t) {
		pref = filepath.Join(pref, "src")
		if strings.HasPrefix(dir, pref) {
			return strings.Trim(strings.TrimPrefix(dir, pref), string(filepath.Separator))
		}
	}
	t.Fatal("cannot find the test package in GOPATH")
	return ""
}

type testEvent struct {
	Time    time.Time
	Action  string
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

type testNode struct {
	Test   string
	Events []testEvent
	Sub    []*testNode
}

func (root *testNode) AddEvent(e testEvent) {
	sub := strings.SplitN(e.Test, "/", 3)
	root.Test = sub[0]
	if len(sub) == 1 {
		root.Events = append(root.Events, e)
		return
	}
	sub = sub[1:]
	e.Test = strings.Join(sub, "/")
	for _, s := range root.Sub {
		if s.Test == sub[0] {
			s.AddEvent(e)
			return
		}
	}
	s := &testNode{}
	root.Sub = append(root.Sub, s)
	s.AddEvent(e)
}

func (root *testNode) RunTests(t *testing.T) {
replay:
	for _, e := range root.Events {
		switch e.Action {
		case "run":
			for _, s := range root.Sub {
				s := s
				t.Run(s.Test, s.RunTests)
			}
		case "output":
			trimmed := strings.TrimSpace(e.Output)
			for _, pref := range []string{
				"=== RUN",
				"--- FAIL:",
				"--- PASS:",
			} {
				if strings.HasPrefix(trimmed, pref) {
					continue replay
				}
			}
			t.Log(strings.TrimSuffix(e.Output, "\n"))
		case "fail":
			t.Error()
		}
	}
}

func reconstructTestLog(t *testing.T, r io.Reader) {
	var root testNode
	dec := json.NewDecoder(r)
	for {
		var e testEvent
		err := dec.Decode(&e)
		if err == io.EOF {
			break
		} else if err != nil {
			require.NoError(t, err, "decoding error")
		}
		root.AddEvent(e)
	}
	root.RunTests(t)
}
