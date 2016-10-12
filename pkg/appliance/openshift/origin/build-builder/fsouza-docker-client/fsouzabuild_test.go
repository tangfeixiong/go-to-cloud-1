package builder

import (
	"archive/tar"
	"bytes"

	"os"
	"testing"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

func TestFsouzaBuild(t *testing.T) {
	client, err := docker.NewClientFromEnv()
	if len(os.Getenv("DOCKER_HOST")) > 0 {
		t.Logf("endpoint > %s", os.Getenv("DOCKER_HOST"))
	} else {
		t.Log("endpoint > ", "unix:///var/run/docker.sock")
	}

	//client, err := docker.NewClient("http://localhost:4243")
	if err != nil {
		t.Fatal(err)
	}

	dockerfile := `
FROM alpine:latest
RUN apk add --update netcat-openbsd && rm -rf /var/cache/apk/*
`
	now := time.Now()
	inputbuf, outputbuf := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	tr := tar.NewWriter(inputbuf)
	tr.WriteHeader(&tar.Header{Name: "Dockerfile", Size: int64(len(dockerfile)), ModTime: now, AccessTime: now, ChangeTime: now})
	tr.Write([]byte(dockerfile))
	tr.Close()
	opts := docker.BuildImageOptions{
		Name:        "tangfeixiong/alpine-with-netcat1",
		NoCache:     true,
		InputStream: inputbuf,
		//OutputStream: outputbuf,
		OutputStream: os.Stdout,
	}
	if err := client.BuildImage(opts); err != nil {
		t.Fatal(err)
	}
	t.Log(outputbuf.String())
}
