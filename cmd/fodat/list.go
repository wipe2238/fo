package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/wipe2238/fo/cmd"
	"github.com/wipe2238/fo/compress/lzss"
	"github.com/wipe2238/fo/dat"
)

func init() {
	var cmd = &cobra.Command{
		Use:   "list [dat file]",
		Short: "List files in DAT file",

		GroupID: app.GroupID,
		Args:    cobra.ExactArgs(1), // TODO: allow multiple files
		RunE:    func(cmd *cobra.Command, args []string) error { return argList(cmd, args) },
	}

	app.AddCommand(cmd)
}

func argList(arg *cobra.Command, args []string) (err error) {
	// all todo/fixme/etc in here applies to every other arg as well

	// FIXME: expose functionality to users
	if err = cmd.ResolveFilename(&args[0], "@"); err != nil {
		return err
	}

	var (
		osFile  *os.File
		datFile dat.FalloutDat
	)

	if osFile, datFile, err = dat.Open(args[0]); err == nil {
		// `list` doesn't need stream open to work
		osFile.Close()
	} else {
		return err
	}

	return doList(arg, datFile)
}

func doList(_ *cobra.Command, datFile dat.FalloutDat) (err error) {
	for _, dir := range datFile.GetDirs() {
		fmt.Printf("%s\n", dir.GetPath())

		for _, file := range dir.GetFiles() {
			var (
				pack string
				save int64
				perc uint64 = 100
			)

			if datFile.GetGame() == 1 {
				switch file.GetCompressMode() {
				case lzss.FalloutCompressStore:
					pack = "store"
				case lzss.FalloutCompressNone:
					pack = "none"
				case lzss.FalloutCompressLZSS:
					pack = "lzss"
				default:
					pack = fmt.Sprintf("(%d)", file.GetCompressMode())
				}
			} else if datFile.GetGame() == 2 {
				switch file.GetCompressMode() {
				case 0:
					pack = "none"
				case 1:
					pack = "pack"
				default:
					pack = fmt.Sprintf("(%d)", file.GetCompressMode())
				}
			}

			if file.GetSizeReal() > 0 {
				save = file.GetSizeReal() - file.GetSizePacked()
				perc = (uint64(file.GetSizePacked()) * 100) / uint64(file.GetSizeReal())
			}
			fmt.Printf("  %-12s %5s %-10s %8d %8d %8d %3d%%\n", file.GetName(), pack, fmt.Sprintf("0x%X", file.GetOffset()), file.GetSizeReal(), file.GetSizePacked(), save, perc)
		}
	}

	return nil
}
