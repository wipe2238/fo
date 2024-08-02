package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/wipe2238/fo/cmd"
	"github.com/wipe2238/fo/dat"
)

const errUnpack = "unpack:"

var optionsUnpack = struct {
	// TODO: FilenameCase
	// TODO: Extension
	IgnoreMissing bool

	Single bool
}{}

var doUnpackMulti func(*os.File, dat.FalloutDat, map[string]dat.FalloutFile, string) error

func init() {
	var cmdUnpack = &cobra.Command{
		Use:   "unpack <dat file> <output directory> <file/directory name>...",
		Short: "Unpack files from DAT file",

		GroupID: app.GroupID,
		Args:    cobra.MinimumNArgs(3),
		RunE:    runUnpack,
	}

	cmdUnpack.Flags().BoolVar(&optionsUnpack.IgnoreMissing, "ignore-missing", false, "Requested files which are not present in DAT file will be ignored")

	// WIP
	if doUnpackMulti != nil {
		cmdUnpack.Flags().BoolVar(&optionsUnpack.Single, "single", false, "")
	} else {
		optionsUnpack.Single = true
	}

	app.AddCommand(cmdUnpack)
}

func runUnpack(cmdUnpack *cobra.Command, args []string) (err error) {
	if err = cmd.ResolveFilename(&args[0], "@"); err != nil {
		return err
	}

	var (
		osFile  *os.File
		datFile dat.FalloutDat
	)

	var names []string
	for idx := range args {
		if idx >= 2 {
			// filenames starting with `@` contains a list of files to unpack
			if strings.HasPrefix(args[idx], "@") {
				var filelist = filepath.Clean(args[idx][1:])
				if filelist == "" {
					return fmt.Errorf("%s empty fileslist name argument(%d)", errUnpack, idx)
				}

				var namesFilelist []string
				if namesFilelist, err = cmd.ReadFileLines(filelist); err != nil {
					return err
				}

				names = append(names, namesFilelist...)
			} else {
				names = append(names, args[idx])
			}

			continue
		}

		args[idx] = filepath.Clean(args[idx])
	}

	// TODO: "*"
	for idx := range names {
		names[idx] = strings.ReplaceAll(names[idx], `\`, "/")
		names[idx] = path.Clean(names[idx])
		names[idx] = strings.ToLower(names[idx])
	}

	names = slices.Compact(names)
	names = slices.Clip(names)

	if osFile, datFile, err = dat.Open(args[0]); err != nil {
		return err
	}

	return doUnpack(osFile, datFile, names, filepath.Clean(args[1]))
}

func doUnpack(osFile *os.File, datFile dat.FalloutDat, names []string, dirOutput string) (err error) {
	if len(names) < 1 {
		return fmt.Errorf("unpack: list of files to unpack is empty")
	}

	// removes duplicates and empty entries
	var cleanupStringSlice = func(slice []string) []string {
		return slices.DeleteFunc(slices.Compact(slice), func(name string) bool {
			return name == ""
		})
	}

	var durationStart = time.Now()
	var dirMap = make(map[string]dat.FalloutDir)
	var fileMap = make(map[string]dat.FalloutFile)
	for _, dir := range datFile.GetDirs() {
		dirMap[strings.ToLower(dir.GetPath())] = dir
		for _, file := range dir.GetFiles() {
			fileMap[strings.ToLower(file.GetPath())] = file
		}
	}
	var durationMaps = time.Since(durationStart)

	// List of files to extract, all keys must be lowercased
	var fileExtract = make(map[string]dat.FalloutFile)

	// List of files to extract, all keys must be lowercased;
	// later converted to files paths added to `fileExtract`
	var dirExtract = make(map[string]dat.FalloutDir)

	durationStart = time.Now()
	for idx, name := range names {
		var pathLower string

		if file, ok := fileMap[name]; ok {
			pathLower = strings.ToLower(file.GetPath())

			fileExtract[pathLower] = file
			names[idx] = ""
		}

		if dir, ok := dirMap[name]; ok {
			pathLower = strings.ToLower(dir.GetPath())

			dirExtract[pathLower] = dir
			names[idx] = ""

			continue
		}
	}

	names = cleanupStringSlice(names)

	// If `names` slice is still non-empty,
	// check if it contains paths to directories which do not have their own dir entry
	if len(names) > 0 {
		for idx, name := range names {
			for dirName, dir := range dirMap {
				if strings.HasPrefix(dirName, name+"/") {
					dirExtract[strings.ToLower(dir.GetPath())] = dir
					names[idx] = ""
				}
			}
		}

		names = cleanupStringSlice(names)
	}

	// It's time to give up or ignore any remains
	if len(names) > 0 && !optionsUnpack.IgnoreMissing {
		slices.Sort(names)
		return fmt.Errorf("%s cannot find following files/directories: '%s'", errUnpack, strings.Join(names, "', '"))
	}
	var durationFiles = time.Since(durationStart)

	// Append files in wanted directories to filenames list
	durationStart = time.Now()
	for _, dir := range dirExtract {
		for _, file := range dir.GetFiles() {
			fileExtract[strings.ToLower(file.GetPath())] = file
		}
	}
	var durationDirs = time.Since(durationStart)

	clear(fileMap)
	clear(dirMap)
	clear(dirExtract)
	runtime.GC()

	durationStart = time.Now()
	if optionsUnpack.Single {
		var sorted = make([]dat.FalloutFile, 0, len(fileExtract))
		for _, file := range fileExtract {
			sorted = append(sorted, file)
		}

		slices.SortFunc(sorted, func(a dat.FalloutFile, b dat.FalloutFile) int {
			if a.GetPath() < b.GetPath() {
				return -1
			} else if a.GetPath() > b.GetPath() {
				return 1
			}

			return 0
		})

		for _, file := range sorted {
			var filename = filepath.Clean(filepath.FromSlash(dirOutput + "/" + file.GetPath()))

			var bytes []byte
			if bytes, err = file.GetBytesReal(osFile); err != nil {
				return fmt.Errorf("unpack: %w", err)
			}

			if err = os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
				return fmt.Errorf("doUnpackData(%s%c): %w", filepath.Dir(filename), filepath.Separator, err)
			}

			if err = os.WriteFile(filename, bytes, 0644); err != nil {
				return fmt.Errorf("doUnpackData(%s): %w", filename, err)
			}

			fmt.Printf("%s → %s\n", file.GetPath(), filename)
		}
	} else {
		// WIP
		if err = doUnpackMulti(osFile, datFile, fileExtract, dirOutput); err != nil {
			return err
		}
	}
	var durationUnpack = time.Since(durationStart)

	fmt.Println("durationMaps   =", durationMaps)
	fmt.Println("durationFiles  =", durationFiles)
	fmt.Println("durationDirs   =", durationDirs)
	fmt.Println("durationUnpack =", durationUnpack)

	return nil
}
