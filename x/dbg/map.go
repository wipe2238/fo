package dbg

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

type Map map[string]any

const errPackage = "fo/x/dbg:"

//

func (dbg Map) AddOffset(name string, stream io.Seeker) (err error) {
	return dbg.AddOffsetSeek(name, stream, 0, io.SeekCurrent)
}

func (dbg Map) AddOffsetSeek(name string, stream io.Seeker, offset int64, whence int) (err error) {
	var pos int64
	if pos, err = stream.Seek(offset, whence); err != nil {
		return err
	}

	dbg[name] = pos

	return nil
}

func (dbg Map) AddSize(name string, keyBegin string, keyEnd string) (err error) {
	var (
		begin int64
		end   int64
		ok    bool
	)

	begin, ok = dbg[keyBegin].(int64)
	if !ok {
		return fmt.Errorf("%s keybegin(%s) must be int64", errPackage, keyBegin)
	}

	end, ok = dbg[keyEnd].(int64)
	if !ok {
		return fmt.Errorf("%s keyEnd(%s) must be int64", errPackage, keyEnd)
	}

	if begin > end {
		return fmt.Errorf("%s keyBegin(%s = %d) > keyEnd(%s = %d)", errPackage, keyBegin, begin, keyEnd, end)
	}

	dbg[name] = end - begin

	return nil
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

func (dbg Map) Dump(keysPrefix string, dumpPrefix string, callback func(string, any, string, string)) {
	var (
		keyN = dbg.KeysMaxLen(keysPrefix)
		valN = dbg.ValsTypeMaxLen(keysPrefix)
	)

	for _, key := range dbg.Keys(keysPrefix) {
		var (
			val = dbg[key]

			left  = dumpPrefix + fmt.Sprintf("%-*s", keyN, key)
			right = fmt.Sprintf("%-*T", valN, val) + " = " + Fmt(" = ", val)
		)

		if callback != nil {
			callback(key, val, left, right)
		} else {
			fmt.Println(left, ":", right)
		}
	}
}
