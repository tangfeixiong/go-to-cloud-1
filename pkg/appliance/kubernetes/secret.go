package kubernetes

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/helm/helm-classic/codec"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapiv1 "k8s.io/kubernetes/pkg/api/v1"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	utilerrors "k8s.io/kubernetes/pkg/util/errors"
)

func createSecret(c *kclient.Client, namespace string, secret *kapiv1.Secret) ([]byte, *kapi.Secret, *kapiv1.Secret, error) {
	data, err := c.RESTClient.Verb("POST").Namespace(namespace).Resource("Secrets").Body(secret).DoRaw()
	if err != nil {
		glog.Errorf("Could not access kubernetes: %+v", err)
		return nil, nil, nil, err
	}
	glog.V(10).Infof("received from creation: %+v", data)

	hco, err := codec.JSON.Decode(data).One()
	if err != nil {
		glog.Errorf("Could not setup hco: %+v\n", err)
		return nil, nil, nil, err
	}
	result := &kapi.Secret{}
	if err := hco.Object(result); err != nil {
		glog.Errorf("Could not decode runtime object: %+v\n", err)
		return data, nil, nil, err
	}
	v1 := &kapiv1.Secret{}
	if err := hco.Object(v1); err != nil {
		glog.Errorf("Could not decode runtime object: %+v\n", err)
		return data, result, nil, err
	}
	return data, result, v1, nil
}

func retrieveSecret(c *kclient.Client, namespace, secret string) (*kapi.Secret, error) {
	data, err := c.RESTClient.Verb("GET").Namespace(namespace).Resource("Secrets").Name(secret).DoRaw()
	if err != nil {
		glog.Errorf("Could not access kubernetes: %+v", err)
		return nil, err
	}

	hco, err := codec.JSON.Decode(data).One()
	if err != nil {
		glog.Errorf("Could not setup hco: %+v\n", err)
		return nil, err
	}
	meta := &unversioned.TypeMeta{}
	if err := hco.Object(meta); err != nil {
		glog.Errorf("Could not decode into TypeMeta: %+v\n", err)
		return nil, err
	}
	if strings.EqualFold("Status", meta.Kind) {
		return nil, nil
	}

	obj := &kapi.Secret{}
	if err := hco.Object(obj); err != nil {
		glog.Errorf("Could not decode runtime object: %+v\n", err)
		return nil, err
	}
	return obj, nil
}

func updateSecret(c *kclient.Client, namespace string, obj *kapi.Secret) (*kapi.Secret, error) {
	result, err := c.Secrets(namespace).Update(obj)
	if err != nil {
		glog.Errorf("Could not update secret: %+v\n", err)
		return nil, err
	}
	glog.V(10).Infof("Secret updated: %+v\n", result)
	return result, nil
}

func deleteSecret(c *kclient.Client, namespace, secret string) error {
	err := c.Secrets(namespace).Delete(secret)
	if err != nil {
		glog.Errorf("Could not delete secret: %+v\n", err)
		return err
	}
	glog.V(10).Infof("Secret deleted")
	return nil
}

type debugError interface {
	DebugError() (msg string, args []interface{})
}

var fatalErrHandler = fatal

// fatal prints the message and then exits. If V(2) or greater, glog.Fatal
// is invoked for extended information.
func fatal(msg string) {
	// add newline if needed
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}

	if glog.V(2) {
		glog.FatalDepth(2, msg)
	}
	//fmt.Fprint(os.Stderr, msg)
	//os.Exit(1)
	fmt.Fprint(os.Stdout, msg)
}

// CheckErr prints a user friendly error to STDERR and exits with a non-zero
// exit code. Unrecognized errors will be printed with an "error: " prefix.
//
// This method is generic to the command in use and may be used by non-Kubectl
// commands.
func CheckErr(err error) {
	checkErr(err, fatalErrHandler)
}

func checkErr(err error, handleErr func(string)) {
	if err == nil {
		return
	}

	if errors.IsInvalid(err) {
		details := err.(*errors.StatusError).Status().Details
		prefix := fmt.Sprintf("The %s %q is invalid.\n", details.Kind, details.Name)
		errs := statusCausesToAggrError(details.Causes)
		handleErr(kcmdutil.MultilineError(prefix, errs))
	}

	if noMatch, ok := err.(*meta.NoResourceMatchError); ok {
		switch {
		case len(noMatch.PartialResource.Group) > 0 && len(noMatch.PartialResource.Version) > 0:
			handleErr(fmt.Sprintf("the server doesn't have a resource type %q in group %q and version %q", noMatch.PartialResource.Resource, noMatch.PartialResource.Group, noMatch.PartialResource.Version))
		case len(noMatch.PartialResource.Group) > 0:
			handleErr(fmt.Sprintf("the server doesn't have a resource type %q in group %q", noMatch.PartialResource.Resource, noMatch.PartialResource.Group))
		case len(noMatch.PartialResource.Version) > 0:
			handleErr(fmt.Sprintf("the server doesn't have a resource type %q in version %q", noMatch.PartialResource.Resource, noMatch.PartialResource.Version))
		default:
			handleErr(fmt.Sprintf("the server doesn't have a resource type %q", noMatch.PartialResource.Resource))
		}
		return
	}

	// handle multiline errors
	if clientcmd.IsConfigurationInvalid(err) {
		handleErr(MultilineError("Error in configuration: ", err))
	}
	if agg, ok := err.(utilerrors.Aggregate); ok && len(agg.Errors()) > 0 {
		handleErr(kcmdutil.MultipleErrors("", agg.Errors()))
	}

	msg, ok := StandardErrorMessage(err)
	if !ok {
		msg = err.Error()
		if !strings.HasPrefix(msg, "error: ") {
			msg = fmt.Sprintf("error: %s", msg)
		}
	}
	handleErr(msg)
}

func statusCausesToAggrError(scs []unversioned.StatusCause) utilerrors.Aggregate {
	errs := make([]error, len(scs))
	for i, sc := range scs {
		errs[i] = fmt.Errorf("%s: %s", sc.Field, sc.Message)
	}
	return utilerrors.NewAggregate(errs)
}

// StandardErrorMessage translates common errors into a human readable message, or returns
// false if the error is not one of the recognized types. It may also log extended
// information to glog.
//
// This method is generic to the command in use and may be used by non-Kubectl
// commands.
func StandardErrorMessage(err error) (string, bool) {
	if debugErr, ok := err.(debugError); ok {
		glog.V(4).Infof(debugErr.DebugError())
	}
	status, isStatus := err.(errors.APIStatus)
	switch {
	case isStatus:
		switch s := status.Status(); {
		case s.Reason == "Unauthorized":
			return fmt.Sprintf("error: You must be logged in to the server (%s)", s.Message), true
		default:
			return fmt.Sprintf("Error from server: %s", err.Error()), true
		}
	case errors.IsUnexpectedObjectError(err):
		return fmt.Sprintf("Server returned an unexpected response: %s", err.Error()), true
	}
	switch t := err.(type) {
	case *url.Error:
		glog.V(4).Infof("Connection error: %s %s: %v", t.Op, t.URL, t.Err)
		switch {
		case strings.Contains(t.Err.Error(), "connection refused"):
			host := t.URL
			if server, err := url.Parse(t.URL); err == nil {
				host = server.Host
			}
			return fmt.Sprintf("The connection to the server %s was refused - did you specify the right host or port?", host), true
		}
		return fmt.Sprintf("Unable to connect to the server: %v", t.Err), true
	}
	return "", false
}

// MultilineError returns a string representing an error that splits sub errors into their own
// lines. The returned string will end with a newline.
func MultilineError(prefix string, err error) string {
	if agg, ok := err.(utilerrors.Aggregate); ok {
		errs := utilerrors.Flatten(agg).Errors()
		buf := &bytes.Buffer{}
		switch len(errs) {
		case 0:
			return fmt.Sprintf("%s%v\n", prefix, err)
		case 1:
			return fmt.Sprintf("%s%v\n", prefix, messageForError(errs[0]))
		default:
			fmt.Fprintln(buf, prefix)
			for _, err := range errs {
				fmt.Fprintf(buf, "* %v\n", messageForError(err))
			}
			return buf.String()
		}
	}
	return fmt.Sprintf("%s%s\n", prefix, err)
}

// messageForError returns the string representing the error.
func messageForError(err error) string {
	msg, ok := StandardErrorMessage(err)
	if !ok {
		msg = err.Error()
	}
	return msg
}
