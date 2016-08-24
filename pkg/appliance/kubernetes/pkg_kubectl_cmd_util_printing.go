package kubernetes

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/kubectl"
)

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

	if len(sorting) != 0 {
		return &kubectl.SortingPrinter{
			Delegate:  printer,
			SortField: fmt.Sprintf("{%s}", sorting),
		}
	}
	return printer
}
