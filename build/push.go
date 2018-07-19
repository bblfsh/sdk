package build

import (
	"fmt"
	"os"
	"os/exec"
)

func PushEnabled() bool {
	return ciBranch() == "master" && ciTag() != "" && !ciIsPR()
}

func (d *Driver) Push(image string) error {
	if !PushEnabled() {
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
	imageName := "bblfsh/" + m.Language + "-driver"
	latestImage := imageName + ":latest"
	versImage := imageName + ":" + tag
	if user := os.Getenv("DOCKER_USERNAME"); user != "" {
		cmd := exec.Command("docker", "login", "-u="+user, "-p="+os.Getenv("DOCKER_PASSWORD"))
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	err = execIn("", nil, "docker", "tag", image, versImage)
	if err != nil {
		return err
	}
	err = execIn("", nil, "docker", "tag", versImage, latestImage)
	if err != nil {
		return err
	}
	err = execIn("", nil, "docker", "push", versImage)
	if err != nil {
		return err
	}
	err = execIn("", nil, "docker", "push", latestImage)
	if err != nil {
		return err
	}
	return nil
}
