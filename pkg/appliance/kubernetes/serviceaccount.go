package kubernetes

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/helm/helm-classic/codec"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapiv1 "k8s.io/kubernetes/pkg/api/v1"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/kubectl"
)

func createServiceAccount(c *kclient.Client, namespace string, sa *kapiv1.ServiceAccount) ([]byte, *kapi.ServiceAccount, *kapiv1.ServiceAccount, error) {
	data, err := c.RESTClient.Verb("POST").Namespace(namespace).Resource("ServiceAccounts").Body(sa).DoRaw()
	if err != nil {
		glog.Errorf("Could not access kubernetes: %+v", err)
		return nil, nil, nil, err
	}
	glog.V(10).Infof("received from creation: %+v", data)

	hco, err := codec.JSON.Decode(data).One()
	if err != nil {
		glog.Errorf("Could not setup helm codec: %+v\n", err)
		return nil, nil, nil, err
	}
	result := &kapi.ServiceAccount{}
	if err := hco.Object(result); err != nil {
		glog.Errorf("Could not decode runtime object: %+v\n", err)
		return data, nil, nil, err
	}
	v1 := &kapiv1.ServiceAccount{}
	if err := hco.Object(v1); err != nil {
		glog.Errorf("Could not decode runtime object: %+v\n", err)
		return data, result, nil, err
	}

	return data, result, v1, nil
}

func retrieveServiceAccount(c *kclient.Client, namespace, serviceaccount string) ([]byte, *kapi.ServiceAccount, *kapiv1.ServiceAccount, error) {
	data, err := c.RESTClient.Get().Namespace(namespace).Resource("ServiceAccounts").Name(serviceaccount).DoRaw()
	if err != nil {
		glog.Errorf("Could nout access kubernetes: %+v\n", err)
		return nil, nil, nil, err
	}
	glog.V(10).Infof("Received from creation: %+v", data)

	hco, err := codec.JSON.Decode(data).One()
	if err != nil {
		glog.Errorf("Could not setup helm codec: %+v\n", err)
		return nil, nil, nil, err
	}
	meta := &unversioned.TypeMeta{}
	if err := hco.Object(meta); err != nil {
		glog.Errorf("Could not decode into TypeMeta: %+v\n", err)
		return nil, nil, nil, err
	}
	if strings.EqualFold("Status", meta.Kind) {
		return nil, nil, nil, nil
	}

	result := &kapi.ServiceAccount{}
	if err := hco.Object(result); err != nil {
		glog.Errorf("Could not decode runtime object: %+v\n", err)
		return data, nil, nil, err
	}
	v1 := &kapiv1.ServiceAccount{}
	if err := hco.Object(v1); err != nil {
		glog.Errorf("Could not decode runtime object: %+v\n", err)
		return data, result, nil, err
	}
	return data, result, v1, nil
}

func updateServiceAccount(c *kclient.Client, namespace string, obj *kapi.ServiceAccount) (*kapi.ServiceAccount, error) {
	result, err := c.ServiceAccounts(namespace).Update(obj)
	if err != nil {
		glog.Errorf("Could not update serviceaccount: %+v\n", err)
		return nil, err
	}
	glog.V(10).Infof("ServiceAccount updated: %+v\n", result)
	return result, nil
}

// OutputVersion returns the preferred output version for generic content (JSON, YAML, or templates)
// defaultVersion is never mutated.  Nil simply allows clean passing in common usage from client.Config
func OutputVersion(outputversion string, defaultVersion *unversioned.GroupVersion) (unversioned.GroupVersion, error) {
	outputVersionString := outputversion /* GetFlagString(cmd, "output-version") */
	if len(outputVersionString) == 0 {
		if defaultVersion == nil {
			return unversioned.GroupVersion{}, nil
		}

		return *defaultVersion, nil
	}

	return unversioned.ParseGroupVersion(outputVersionString)
}

// PrinterForCommand returns the default printer for this command.
// Requires that printer flags have been added to cmd (see AddPrinterFlags).
func PrinterForCommand(outputFormat, templateFile, sortby string) (kubectl.ResourcePrinter, bool, error) {
	/*outputFormat := GetFlagString(cmd, "output")*/

	// templates are logically optional for specifying a format.
	// TODO once https://github.com/kubernetes/kubernetes/issues/12668 is fixed, this should fall back to GetFlagString
	/*templateFile, _ := cmd.Flags().GetString("template")*/
	if len(outputFormat) == 0 && len(templateFile) != 0 {
		outputFormat = "template"
	}

	templateFormat := []string{
		"go-template=", "go-template-file=", "jsonpath=", "jsonpath-file=", "custom-columns=", "custom-columns-file=",
	}
	for _, format := range templateFormat {
		if strings.HasPrefix(outputFormat, format) {
			templateFile = outputFormat[len(format):]
			outputFormat = format[:len(format)-1]
		}
	}

	printer, generic, err := kubectl.GetPrinter(outputFormat, templateFile)
	if err != nil {
		return nil, generic, err
	}

	return MaybeWrapSortingPrinter(sortby, printer), generic, nil
}

func MaybeWrapSortingPrinter(sortby string, printer kubectl.ResourcePrinter) kubectl.ResourcePrinter {
	/*sorting, err := cmd.Flags().GetString("sort-by")
	if err != nil {
		// error can happen on missing flag or bad flag type.  In either case, this command didn't intent to sort
		return printer
	}*/
	sorting := ""

	if len(sorting) != 0 {
		return &kubectl.SortingPrinter{
			Delegate:  printer,
			SortField: fmt.Sprintf("{%s}", sorting),
		}
	}
	return printer
}
