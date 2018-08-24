package build

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
)

var reVers = regexp.MustCompile(`^v\d+`)

func pushEnabled() bool {
	return !ciIsPR() && (ciBranch() == "master" || ciTag() != "")
}

func pushTagEnabled() bool {
	return pushEnabled() && (reVers.MatchString(ciBranch()) || reVers.MatchString(ciTag()))
}

func pushLatestEnabled() bool {
	return pushEnabled() && ciTag() != ""
}

func (d *Driver) Push(image string) error {
	if isCI() {
		log.Printf("CI variables: branch=%q, tag=%q, pr=%v", ciBranch(), ciTag(), ciIsPR())
	}
	if !pushTagEnabled() && !pushLatestEnabled() {
		return fmt.Errorf("push disabled")
	} else if image == "" {
		return fmt.Errorf("image should be specified")
	}
	tag, err := d.VersionTag()
	if err != nil {
		return err
	}
	m, err := d.readManifest()
	if err != nil {
		return err
	}
	if user := os.Getenv("DOCKER_USERNAME"); user != "" {
		cmd := exec.Command("docker", "login", "-u="+user, "-p="+os.Getenv("DOCKER_PASSWORD"))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	push := func(id, name string) error {
		err = execIn("", nil, "docker", "tag", id, name)
		if err != nil {
			return err
		}
		return execIn("", nil, "docker", "push", name)
	}
	imageName := "bblfsh/" + m.Language + "-driver"
	if pushTagEnabled() {
		if err := push(image, imageName+":"+tag); err != nil {
			return err
		}
	}
	if pushLatestEnabled() {
		if err := push(image, imageName+":latest"); err != nil {
			return err
		}
	}
	return nil
}
