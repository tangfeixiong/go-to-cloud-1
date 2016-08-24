package origin

import (
	"os"
	"testing"
)

func TestNewBuild(t *testing.T) {
	cmd := newOsoNewBuildCommand()

	overrideStringFlag(cmd.NewBuildCmd, "context-dir", "edge", "", "", "latest")
	overrideStringFlag(cmd.NewBuildCmd, "name", "nc-docker", "", "", "nc-alpine")
	overrideStringFlag(cmd.NewBuildCmd, "strategy", "docker", "", "", "docker")
	overrideBoolFlag(cmd.NewBuildCmd, "to-docker", true, "", "", true)
	overrideStringFlag(cmd.NewBuildCmd, "to", "172.17.4.50:30005/tangfx/nc-alpine:latest", "", "", "tangfx/osobuilds:latest")
	overrideIntFlag(cmd.NewBuildCmd, "loglevel", 99, "", "", 10)

	err := cmd.Execute([]string{fakeGitURI}, "tangfx", os.Stdout, os.Stdin)

	/*err := cmd.Execute([]string{fakeGitURI, "--context-dir=/edge",
	"--strategy=docker",
	"--to-docker=true", "--to=172.17.4.50:30005/tangfx/netcat-alpine:edge",
	"--loglevel=10"}, os.Stdout, os.Stdin)*/

	/*
		//overrideBoolFlag(cmd.NewBuildCmd, "allow-missing-images", true, "", "", true)
		overrideStringFlag(cmd.NewBuildCmd, "name", "httpd-centos", "", "", "httpd-centos")
		overrideStringFlag(cmd.NewBuildCmd, "dockerfile", "FROM docker.io/centos:latest\nRUN yum install -y httpd", "", "", "")
		//overrideBoolFlag(cmd.NewBuildCmd, "insecure-registry", "true", "", "", true)
		overrideStringFlag(cmd.NewBuildCmd, "strategy", "docker", "", "", "docker")
		//overrideStringFlag(cmd.NewBuildCmd, "docker-image", "docker.io/openshift/origin-docker-builder:v1.3.0-alpha.1", "", "", "")
		overrideBoolFlag(cmd.NewBuildCmd, "to-docker", true, "", "", true)
		overrideStringFlag(cmd.NewBuildCmd, "to", "172.17.4.50:30005/tangfx/httpd-centos7:latest", "", "", "")
		overrideIntFlag(cmd.NewBuildCmd, "loglevel", 99, "", "", 10)

		err := cmd.Execute([]string{}, "tangfx", os.Stdout, os.Stdin)
	*/

	/*err := cmd.Execute([]string{"--allow-missing-images=true",
	"--dockerfile='FROM docker.io/centos:7\nRUN yum install -y httpd'",
	"--strategy=docker", "--docker-image=docker.io/centos:7",
	"--to-docker=true", "--to=172.17.4.50:30005/tangfx/httpd-centos7:latest",
	"--loglevel=99"}, os.Stdout, os.Stdin)*/

	if err != nil {
		t.Fatal(err)
	}
}
