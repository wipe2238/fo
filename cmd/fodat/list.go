package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wipe2238/fo/cmd"
	"github.com/wipe2238/fo/compress/lzss"
	"github.com/wipe2238/fo/dat"
)

func init() {
	var cmd = &cobra.Command{
		Use:   "list [dat file]",
		Short: "List files in DAT file (only DAT1 supported right now)",

		GroupID: app.GroupID,
		Args:    cobra.ExactArgs(1),
		RunE:    func(cmd *cobra.Command, args []string) error { return argList(cmd, args) },
	}

	app.AddCommand(cmd)
}

func argList(arg *cobra.Command, args []string) (err error) {
	// TODO move to `cmd`, make less verbose
	// FIXME expose @steam functionality to users
	// replace "@steam:<game>:<file>" with absolute path, if needed
	if strings.HasPrefix(args[0], "@") {
		var tmp string
		if tmp, err = cmd.SteamGameFile(args[0], "@steam"); err == nil {
			args[0] = tmp
		} else {
			return err
		}
	}

	var osFile *os.File
	if osFile, err = os.Open(args[0]); err == nil {
		defer osFile.Close()
	} else {
		return err
	}

	// TODO detect DAT1/DAT2 automagically

	var datFile dat.FalloutDat
	if datFile, err = dat.Fallout1(osFile, dat.DefaultOptions()); err != nil {
		return err
	}

	return doList(arg, osFile, datFile)
}

// XXX indexes
func doList(_ *cobra.Command, osFile *os.File, datFile dat.FalloutDat) (err error) {
	var (
		dirs  []dat.FalloutDir
		files []dat.FalloutFile
	)

	if dirs, err = datFile.GetDirs(); err != nil {
		return err
	}

	for idxDir, dir := range dirs {
		if files, err = dir.GetFiles(); err != nil {
			fmt.Println("ERROR: @dat.GetFiles("+dir.GetPath()+")", err)
			continue
		}

		fmt.Printf("%-55s [%4d]\n", dir.GetPath(), idxDir)

		for idxFile, file := range files {
			var (
				pack         string
				packCompress uint8
				save         int32  = 0
				perc         uint64 = 100
			)

			if packCompress, err = lzss.Fallout1.CompressMethod(osFile, file); err == nil {
				switch packCompress {
				case lzss.FalloutCompressZero:
					pack = "zero"
				case lzss.FalloutCompressPlain:
					pack = "plain"
				case lzss.FalloutCompressStore:
					pack = "store"
				case lzss.FalloutCompressLZSS:
					pack = "lzss"
				}
			} else {
				fmt.Println("ERROR: list("+dir.GetPath()+")", err)
			}

			if file.GetPacked() {
				save = file.GetSizeReal() - file.GetSizePacked()
				perc = (uint64(file.GetSizePacked()) * 100) / uint64(file.GetSizeReal())
			}
			fmt.Printf("  %-12s %5s %-10s %8d %8d %8d %3d%% [%4d][%4d]\n", file.GetName(), pack, fmt.Sprintf("0x%X", file.GetOffset()), file.GetSizeReal(), file.GetSizePacked(), save, perc, idxDir, idxFile)
		}
	}

	return nil
}
