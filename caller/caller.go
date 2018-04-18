// Package caller contains convenience functions to get function caller information.
package caller

import (
	"fmt"
	"runtime"
	"strconv"
)

// CreateStackInfoFunc returns a closure function that for a given depth returns stack caller information (file and line).
func CreateStackInfoFunc(depth int, shorten bool) func() string {
	return func() string {
		_, file, line, ok := runtime.Caller(depth)
		if !ok {
			return "???"
		}
		if shorten {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		//return fmt.Sprintf("%s:%d", file, line)
		return file + ":" + strconv.FormatInt(int64(line), 10)
	}
}

// CreateStackInfoWithFuncnameFunc returns a closure function that for a given depth returns stack caller information (file, line and function name *if* found).
func CreateStackInfoWithFuncnameFunc(depth int, shorten bool, withFn bool) func() string {
	return func() string {
		pc, file, line, ok := runtime.Caller(depth)
		if !ok {
			return "???"
		}

		var fn *runtime.Func
		if withFn {
			// fn also contains a FileLine() method returning (file string, line int)
			fn = runtime.FuncForPC(pc)
		}

		if shorten {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}

		if fn != nil {
			return fmt.Sprintf("%s:%d(%s)", file, line, fn.Name())
		}
		return fmt.Sprintf("%s:%d", file, line)
	}
}

// GetStackInfo returns caller stack information (file and line) for a given depth.
func GetStackInfo(depth int) string {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		return fmt.Sprintf("???:???")
	}
	return fmt.Sprintf("%s:%d", file, line)
}
