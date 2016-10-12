package utility

import (
	"reflect"
	"text/template"

	"github.com/Masterminds/sprig"
)

// http://stackoverflow.com/questions/22367337/last-item-in-a-golang-template-range
//
var TplFns = template.FuncMap{
	"last": func(x int, a interface{}) bool {
		return x == reflect.ValueOf(a).Len()-1
	},
	"plus1": func(x int) int {
		return x + 1
	},
}

var SprigTxtTplFns template.FuncMap = sprig.TxtFuncMap()
