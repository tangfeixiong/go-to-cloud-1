package client

import (
	"errors"
)

const (
	loginLong = `
Log in to your server and save login for subsequent use

First-time users of the client should run this command to connect to a server,
establish an authenticated session, and save connection to the configuration file. The
default configuration will be saved to your home directory under
".kube/config".

The information required to login -- like username and password, a session token, or
the server details -- can be provided through flags. If not provided, the command will
prompt for user input as needed.`

	loginExample = `  # Log in interactively
  $ %[1]s login

  # Log in to the given server with the given certificate authority file
  $ %[1]s login localhost:8443 --certificate-authority=/path/to/cert.crt

  # Log in to the given server with the given credentials (will not prompt interactively)
  $ %[1]s login localhost:8443 --username=myuser --password=mypass`

	requestProjectLong = `
Create a new project for yourself

If your administrator allows self-service, this command will create a new project for you and assign you
as the project admin.

After your project is created it will become the default project in your config.`

	requestProjectExample = `  # Create a new project with minimal information
  $ %[1]s web-team-dev

  # Create a new project with a display name and description
  $ %[1]s web-team-dev --display-name="Web Team Development" --description="Development project for the web team."`

	newBuildLong = `
Create a new build by specifying source code

This command will try to create a build configuration for your application using images and
code that has a public repository. It will lookup the images on the local Docker installation
(if available), a Docker registry, or an image stream.

If you specify a source code URL, it will set up a build that takes your source code and converts
it into an image that can run inside of a pod. Local source must be in a git repository that has a
remote repository that the server can see.

Once the build configuration is created a new build will be automatically triggered.
You can use '%[1]s status' to check the progress.`

	newBuildExample = `
  # Create a build config based on the source code in the current git repository (with a public
  # remote) and a Docker image
  $ %[1]s new-build . --docker-image=repo/langimage

  # Create a NodeJS build config based on the provided [image]~[source code] combination
  $ %[1]s new-build openshift/nodejs-010-centos7~https://github.com/openshift/nodejs-ex.git

  # Create a build config from a remote repository using its beta2 branch
  $ %[1]s new-build https://github.com/openshift/ruby-hello-world#beta2

  # Create a build config using a Dockerfile specified as an argument
  $ %[1]s new-build -D $'FROM centos:7\nRUN yum install -y httpd'

  # Create a build config from a remote repository and add custom environment variables
  $ %[1]s new-build https://github.com/openshift/ruby-hello-world RACK_ENV=development

  # Create a build config from a remote repository and inject the npmrc into a build
  $ %[1]s new-build https://github.com/openshift/ruby-hello-world --build-secret npmrc:.npmrc

  # Create a build config that gets its input from a remote repository and another Docker image
  $ %[1]s new-build https://github.com/openshift/ruby-hello-world --source-image=openshift/jenkins-1-centos7 --source-image-path=/var/lib/jenkins:tmp`

	newBuildNoInput = `You must specify one or more images, image streams, or source code locations to create a build.

To build from an existing image stream tag or Docker image, provide the name of the image and
the source code location:

  $ %[1]s new-build openshift/nodejs-010-centos7~https://github.com/openshift/nodejs-ex.git

If you only specify the source repository location (local or remote), the command will look at
the repo to determine the type, and then look for a matching image on your server or on the
default Docker registry.

  $ %[1]s new-build https://github.com/openshift/nodejs-ex.git

will look for an image called "nodejs" in your current project, the 'openshift' project, or
on the Docker Hub.
`

	startBuildLong = `
Start a build

This command starts a new build for the provided build config or copies an existing build using
--from-build=<name>. Pass the --follow flag to see output from the build.

In addition, you can pass a file, directory, or source code repository with the --from-file,
--from-dir, or --from-repo flags directly to the build. The contents will be streamed to the build
and override the current build source settings. When using --from-repo, the --commit flag can be
used to control which branch, tag, or commit is sent to the server. If you pass --from-file, the
file is placed in the root of an empty directory with the same filename. Note that builds
triggered from binary input will not preserve the source on the server, so rebuilds triggered by
base image changes will use the source specified on the build config.
`

	startBuildExample = `  # Starts build from build config "hello-world"
  $ %[1]s start-build hello-world

  # Starts build from a previous build "hello-world-1"
  $ %[1]s start-build --from-build=hello-world-1

  # Use the contents of a directory as build input
  $ %[1]s start-build hello-world --from-dir=src/

  # Send the contents of a Git repository to the server from tag 'v2'
  $ %[1]s start-build hello-world --from-repo=../hello-world --commit=v2

  # Start a new build for build config "hello-world" and watch the logs until the build
  # completes or fails.
  $ %[1]s start-build hello-world --follow

  # Start a new build for build config "hello-world" and wait until the build completes. It
  # exits with a non-zero return code if the build fails.
  $ %[1]s start-build hello-world --wait`
)

var (
	errNotFound       error = errors.New("Not found")
	errNotImplemented error = errors.New("Not implemented")
	errUnexpected     error = errors.New("Unexpected")
)
