package dbg

import (
	"fmt"
	"reflect"
	"strings"
)

func Fmt(separator string, val any) string {
	return strings.Join(Fmts(val), separator)
}

func Fmts(val any) (result []string) {
	var valKind = reflect.ValueOf(val).Kind()

	result = make([]string, 0)

	if valKind >= reflect.Int && valKind <= reflect.Uint64 {
		result = append(result, fmt.Sprintf("0x%0X", val))
		result = append(result, fmt.Sprintf("%d", val))
	} else if valKind == reflect.Array || valKind == reflect.Slice {
		// FIXME assumes that slice/array can hold ints only
		result = append(result, strings.ReplaceAll(fmt.Sprintf("%#X", val), "X", "x"))
		result = append(result, fmt.Sprintf("%d", val))
	} else {
		result = append(result, fmt.Sprintf("%#v", val))
	}

	return result
}
