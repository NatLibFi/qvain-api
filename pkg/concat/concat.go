// Package concat is a utility package that concatenates multiple files into one with optional date dependency checking.
/*
examples:
	// check if file is older than files in a directory
	if IsOlderThanDirectory("parts.txt", ".") {
		fmt.Println("parts.txt needs update")
	} else {
		fmt.Println("parts.txt up to date")
	}

	// check if file is older than files matching a shell glob pattern
	if IsOlderThanGlob("parts.txt", "part[0-9].txt") {
		fmt.Println("parts.txt needs update")
	} else {
		fmt.Println("parts.txt up to date")
	}

	// concat files:
	err := Concat("concat.txt", "part[0-9].txt", nil)

	// concat files, wrapping them in a json object with the filename as key:
	if err := Concat("big.json", "json[0-9].json", JsonObjectWrapper); err != nil {
		log.Fatal(err)
	}

	// concat files, wrapping them in a json array:
	if err := Concat("big.json", "json[0-9].json", JsonArrayWrapper); err != nil {
		log.Fatal(err)
	}
*/
package concat

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	// Use system temp directory or current working directory for temp files
	UseTemp = true

	// Error or skip on non-fatal errors
	SkipOnError = false
)

// Copy (surprisingly) copies a file to another location, possibly overwriting existing files.
// It doesn't use Link/Rename so works across file system boundaries.
func Copy(in, out string) error {
	r, err := os.Open(in)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.OpenFile(out, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666) // or 0640
	if err != nil {
		return err
	}
	defer w.Close()

	if _, err := io.Copy(w, r); err != nil {
		return err
	}

	return nil
}

// Options struct specifies optional arguments to the Concat function for defining separating text.
type Options struct {
	atStart    string
	atEnd      string
	beforeEach string
	afterEach  string
	skipLast   bool

	beforeEachFunc func(fn string) string
}

// JsonObjectWrapper is an Options struct predefined for concatenating Json files into a new object with the filename as key.
var JsonObjectWrapper = &Options{
	atStart:   "{\n",
	atEnd:     "}\n",
	afterEach: ",\n",
	skipLast:  true,

	beforeEachFunc: func(fn string) string {
		fmt.Println("beforeEachFunc:", fn)
		return `"` + strings.TrimSuffix(filepath.Base(fn), filepath.Ext(fn)) + `":`
	},
}

// JsonArrayWrapper is an Options struct predefined for concatenating Json files into a new array.
var JsonArrayWrapper = &Options{
	atStart:   "[\n",
	atEnd:     "]\n",
	afterEach: ",\n",
	skipLast:  true,
}

// Concat concatenates files into a new, large file. It takes an Options struct to define separators between concatenated files.
func Concat(outname, glob string, options *Options) error {
	var written int64

	// get a default options struct if we weren't give one
	if options == nil {
		options = new(Options)
	}

	matches, err := filepath.Glob(glob)
	if err != nil {
		return err
	}

	if len(matches) <= 0 {
		return fmt.Errorf("no matches for glob pattern")
	}

	var outdir string
	if !UseTemp {
		// make temp file in same dir as source file
		outdir = filepath.Dir(outname)
	}

	tmpfile, err := ioutil.TempFile(outdir, outname)
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	fmt.Println("tmpfile:", tmpfile.Name())

	if options.atStart != "" {
		if _, err := tmpfile.Write([]byte(options.atStart)); err != nil {
			return err
		}
	}

	for i, match := range matches {
		r, err := os.Open(match)
		if err != nil {
			return err
		}
		// don't use defer in loop
		//defer r.Close()

		if options.beforeEach != "" {
			if _, err := tmpfile.Write([]byte(options.beforeEach)); err != nil {
				return err
			}
		}
		if options.beforeEachFunc != nil {
			if _, err := tmpfile.Write([]byte(options.beforeEachFunc(match))); err != nil {
				return err
			}
		}

		if n, err := io.Copy(tmpfile, r); err != nil {
			r.Close()
			return err
		} else {
			written += n
		}
		r.Close()

		if options.afterEach != "" && !(options.skipLast && i == len(matches)-1) {
			if _, err := tmpfile.Write([]byte(options.afterEach)); err != nil {
				return err
			}
		}
	}

	if options.atEnd != "" {
		if _, err := tmpfile.Write([]byte(options.atEnd)); err != nil {
			return err
		}
	}

	if err := tmpfile.Close(); err != nil {
		return err
	}

	fmt.Println("written:", written)

	// try link
	if err := os.Rename(tmpfile.Name(), outname); err == nil {
		return nil
	} else {
		//warn("move failed, trying copy")
	}

	// move failed (across partitions?), try copy
	return Copy(tmpfile.Name(), outname)
}

func readDir(dirname string) ([]os.FileInfo, error) {
	dir, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := dir.Readdir(-1)
	dir.Close()
	if err != nil {
		return nil, err
	}
	return list, nil
}

// IsOlderThanGlob checks if a given file is older than the files matching the glob pattern.
// If the file doesn't exist, this function returns true.
func IsOlderThanGlob(bundlename, glob string) (bool, error) {
	bundleInfo, err := os.Stat(bundlename)
	if err != nil {
		// doesn't exist, so return true
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}

	updated := bundleInfo.ModTime()

	matches, err := filepath.Glob(glob)
	if err != nil {
		return false, err
	}

	if len(matches) <= 0 && !SkipOnError {
		return false, fmt.Errorf("no matches for glob pattern")
	}

	for _, match := range matches {
		fi, err := os.Stat(match)
		if err != nil {
			if !SkipOnError {
				return false, fmt.Errorf("error for file %s: %s", match, err)
			}
			continue
		}
		if fi.ModTime().After(updated) {
			return true, nil
		}
	}
	return false, nil
}

// IsOlderThanDirectory checks if a given file is older than any of the files in a given directory.
// If the file doesn't exist, this function returns true.
func IsOlderThanDirectory(bundlename, dirname string) (bool, error) {
	bundleInfo, err := os.Stat(bundlename)
	if err != nil {
		// doesn't exist, so return true
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}

	updated := bundleInfo.ModTime()

	filesInfo, err := readDir(dirname)
	if err != nil {
		return false, err
	}

	for _, fi := range filesInfo {
		if fi.ModTime().After(updated) {
			return true, nil
		}
	}
	return false, nil
}
