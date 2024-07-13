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
		Short: "List files in DAT file (only DAT1 supported right now)",

		GroupID: app.GroupID,
		Args:    cobra.ExactArgs(1),
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

	var osFile *os.File
	if osFile, err = os.Open(args[0]); err == nil {
		defer osFile.Close()
	} else {
		return err
	}

	// TODO: detect DAT1/DAT2 automagically

	var datFile dat.FalloutDat
	if datFile, err = dat.Fallout1(osFile); err != nil {
		return
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

			if file.GetSizeReal() > 0 {
				save = file.GetSizeReal() - file.GetSizePacked()
				perc = (uint64(file.GetSizePacked()) * 100) / uint64(file.GetSizeReal())
			}
			fmt.Printf("  %-12s %5s %-10s %8d %8d %8d %3d%%\n", file.GetName(), pack, fmt.Sprintf("0x%X", file.GetOffset()), file.GetSizeReal(), file.GetSizePacked(), save, perc)
		}
	}

	return nil
}
