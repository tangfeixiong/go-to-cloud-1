package origin

import (
	"fmt"
	"io"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apimachinery/registered"
	"k8s.io/kubernetes/pkg/runtime"
)

// ErrExit is a marker interface for cli commands indicating that the response has been processed
var ErrExit = fmt.Errorf("exit directly")

// convertItemsForDisplay returns a new list that contains parallel elements that have been converted to the most preferred external version
func convertItemsForDisplay(objs []runtime.Object, preferredVersions ...unversioned.GroupVersion) ([]runtime.Object, error) {
	ret := []runtime.Object{}

	for i := range objs {
		obj := objs[i]
		kind, _, err := kapi.Scheme.ObjectKind(obj)
		if err != nil {
			return nil, err
		}
		groupMeta, err := registered.Group(kind.Group)
		if err != nil {
			return nil, err
		}

		requestedVersion := unversioned.GroupVersion{}
		for _, preferredVersion := range preferredVersions {
			if preferredVersion.Group == kind.Group {
				requestedVersion = preferredVersion
				break
			}
		}

		actualOutputVersion := unversioned.GroupVersion{}
		for _, externalVersion := range groupMeta.GroupVersions {
			if externalVersion == requestedVersion {
				actualOutputVersion = externalVersion
				break
			}
			if actualOutputVersion.IsEmpty() {
				actualOutputVersion = externalVersion
			}
		}

		convertedObject, err := kapi.Scheme.ConvertToVersion(obj, actualOutputVersion)
		if err != nil {
			return nil, err
		}

		ret = append(ret, convertedObject)
	}

	return ret, nil
}

// convertItemsForDisplayFromDefaultCommand returns a new list that contains parallel elements that have been converted to the most preferred external version
// TODO: move this function into the core factory PrintObjects method
// TODO: print-objects should have preferred output versions
func convertItemsForDisplayFromDefaultCommand(outputversion string, objs []runtime.Object) ([]runtime.Object, error) {
	requested := outputversion /*kcmdutil.GetFlagString(cmd, "output-version")*/
	version, err := unversioned.ParseGroupVersion(requested)
	if err != nil {
		return nil, err
	}
	return convertItemsForDisplay(objs, version)
}

// VersionedPrintObject handles printing an object in the appropriate version by looking at 'output-version'
// on the command
func VersionedPrintObject(fn func(string, meta.RESTMapper, runtime.Object, io.Writer,
	string, string, string, bool, bool, bool, bool, bool, []string) error,
	outputversion string, mapper meta.RESTMapper, out io.Writer) func(runtime.Object) error {
	return func(obj runtime.Object) error {
		// TODO: fold into the core printer functionality (preferred output version)
		if list, ok := obj.(*kapi.List); ok {
			var err error
			if list.Items, err = convertItemsForDisplayFromDefaultCommand(outputversion, list.Items); err != nil {
				return err
			}
		} else {
			result, err := convertItemsForDisplayFromDefaultCommand(outputversion, []runtime.Object{obj})
			if err != nil {
				return err
			}
			obj = result[0]
		}
		return fn(outputversion, mapper, obj, out, _output_version, _template, _sort_by, _no_headers, _wide, _show_all, _show_labels, _is_watch, _label_columns)
	}
}
