package kubernetes

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/kubernetes"
)

// PrintObject prints an api object given command line flags to modify the output format
func PrintObject(f *util.Factory, mapper meta.RESTMapper, obj runtime.Object, out io.Writer,
	outputversion, template, sortby string, noheaders, wide, showall, showlabels, iswatch bool,
	labelcolumns []string) error {
	gvks, _, err := api.Scheme.ObjectKinds(obj)
	if err != nil {
		return err
	}

	mapping, err := mapper.RESTMapping(gvks[0].GroupKind())
	if err != nil {
		return err
	}

	printer, err := PrinterForMapping(f, mapping, false, outputversion, template, sortby, noheaders, wide, showall, showlabels, iswatch, labelcolumns)
	if err != nil {
		return err
	}
	return printer.PrintObj(obj, out)
}

// PrinterForMapping returns a printer suitable for displaying the provided resource type.
// Requires that printer flags have been added to cmd (see AddPrinterFlags).
func PrinterForMapping(f *util.Factory, mapping *meta.RESTMapping, withNamespace bool,
	outputversion, template, sortby string, noheaders, wide, showall, showlabels, iswatch bool,
	labelcolumns []string) (kubectl.ResourcePrinter, error) {
	printer, ok, err := kubernetes.PrinterForCommand(outputversion, template, sortby)
	if err != nil {
		return nil, err
	}
	if ok {
		clientConfig, err := f.ClientConfig()
		if err != nil {
			return nil, err
		}

		version, err := kubernetes.OutputVersion(outputversion, clientConfig.GroupVersion)
		if err != nil {
			return nil, err
		}
		if version.IsEmpty() && mapping != nil {
			version = mapping.GroupVersionKind.GroupVersion()
		}
		if version.IsEmpty() {
			return nil, fmt.Errorf("you must specify an output-version when using this output format")
		}

		if mapping != nil {
			printer = kubectl.NewVersionedPrinter(printer, mapping.ObjectConvertor, version, mapping.GroupVersionKind.GroupVersion())
		}

	} else {
		// Some callers do not have "label-columns" so we can't use the GetFlagStringSlice() helper
		/*columnLabel, err := cmd.Flags().GetStringSlice("label-columns")
		if err != nil {
			columnLabel = []string{}
		}*/

		printer, err = f.Printer(mapping, &kubectl.PrintOptions{
			NoHeaders:          noheaders, /*GetFlagBool(cmd, "no-headers")*/
			WithNamespace:      withNamespace,
			Wide:               wide,         /*GetWideFlag(cmd)*/
			ShowAll:            showall,      /*GetFlagBool(cmd, "show-all")*/
			ShowLabels:         showlabels,   /*GetFlagBool(cmd, "show-labels")*/
			AbsoluteTimestamps: iswatch,      /*isWatch(cmd)*/
			ColumnLabels:       labelcolumns, /*columnLabel*/
		})
		if err != nil {
			return nil, err
		}
		printer = kubernetes.MaybeWrapSortingPrinter(sortby, printer)
	}

	return printer, nil
}
