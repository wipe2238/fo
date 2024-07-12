package dbg

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

type Map map[string]any

//

func (dbg Map) AddOffset(name string, stream io.Reader) {
	// FIXME: return error, if any
	dbg[name], _ = stream.(io.Seeker).Seek(0, io.SeekCurrent)
}

func (dbg Map) AddSize(name string, keyBegin string, keyEnd string) {
	var (
		begin int64
		end   int64
		ok    bool
	)

	// FIXME: return error instead of panic

	begin, ok = dbg[keyBegin].(int64)
	if !ok {
		panic(fmt.Sprintf("Cannot AddSize(%s): keybegin(%s) must be int64", name, keyBegin))
	}

	end, ok = dbg[keyEnd].(int64)
	if !ok {
		panic(fmt.Sprintf("Cannot AddSize(%s): keyEnd(%s) must be int64", name, keyEnd))
	}

	if begin > end {
		panic(fmt.Sprintf("AddSize(%s) keyBegin(%s=%d) > keyEnd(%s=%d)", name, keyBegin, begin, keyEnd, end))
	}

	dbg[name] = end - begin
}

//

func (dbg Map) Keys(keysPrefix string) (out []string) {
	out = make([]string, 0)

	for key := range dbg {
		if keysPrefix != "" && !strings.HasPrefix(key, keysPrefix) {
			continue
		}

		out = append(out, key)
	}

	slices.Sort(out)

	return out
}

// KeysMaxLen return longest key with given prefix
func (dbg Map) KeysMaxLen(keysPrefix string) (length int) {
	for key := range dbg {
		if keysPrefix != "" && !strings.HasPrefix(key, keysPrefix) {
			continue
		}
		length = max(length, len(key))
	}

	return length
}

func (dbg Map) ValsTypeMaxLen(keysPrefix string) (length int) {
	for key, val := range dbg {
		if keysPrefix != "" && !strings.HasPrefix(key, keysPrefix) {
			continue
		}

		length = max(length, len(fmt.Sprintf("%T", val)))
	}

	return length
}

// KeysMaxLenStr returns result of KeysMaxLen() as string.
// Used when creating format string
//
// Example: `fmt.Sprintf("%-" + myap.KeysMaxLenStr("prefix") + "s", key)“
func (dbg Map) KeysMaxLenStr(keysPrefix string) string {
	return fmt.Sprintf("%d", dbg.KeysMaxLen(keysPrefix))
}

// ValsTypeMaxLenStr returns result ValsTypeMaxLen() as string
func (dbg Map) ValsTypeMaxLenStr(keysPrefix string) string {
	return fmt.Sprintf("%d", dbg.ValsTypeMaxLen(keysPrefix))
}

func (dbg Map) Dump(keysPrefix string, dumpPrefix string, callback func(string, any, string, string)) {
	var (
		keyN = dbg.KeysMaxLenStr(keysPrefix)
		valN = dbg.ValsTypeMaxLenStr(keysPrefix)
	)

	for _, key := range dbg.Keys(keysPrefix) {
		var (
			val   = dbg[key]
			left  = dumpPrefix + fmt.Sprintf("%-"+keyN+"s", key)
			right = fmt.Sprintf("%-"+valN+"T", val) + " = " + Fmt(" = ", val)
		)

		if callback != nil {
			callback(key, val, left, right)
		} else {
			fmt.Println(left, ":", right)
		}
	}
}
