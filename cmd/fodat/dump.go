package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wipe2238/fo/cmd"
	"github.com/wipe2238/fo/dat"
)

func init() {
	var cmdDump = &cobra.Command{
		Use:   "dump <dat file>",
		Short: "Dump DAT file data",

		GroupID: app.GroupID,
		Args:    cobra.ExactArgs(1), // TODO: allow multiple files
		RunE:    runDump,
	}

	app.AddCommand(cmdDump)
}

func runDump(cmdDump *cobra.Command, args []string) (err error) {
	if err = cmd.ResolveFilename(&args[0], "@"); err != nil {
		return err
	}

	var (
		osFile  *os.File
		datFile dat.FalloutDat
	)

	if osFile, datFile, err = dat.Open(args[0]); err == nil {
		// after calling `SetDbg()`, stream is not used
		osFile.Seek(0, io.SeekStart)
		datFile.SetDbg(osFile)
		osFile.Close()

		datFile.FillDbg() // TODO: finish transition to SetDbg()
	} else {
		return err
	}

	return doDump(cmdDump, datFile, args[0])
}

func doDump(_ *cobra.Command, datFile dat.FalloutDat, datName string) (err error) {
	var sizeHuman = func(total int64) string {
		const unit int64 = 1024
		if total < unit {
			return fmt.Sprintf("%d B", total)
		}

		var div, exp = unit, 0
		for n := total / unit; n >= unit; n /= unit {
			div *= unit
			exp++
		}

		return fmt.Sprintf("%.1f %cB", float64(total)/float64(div), "KMGTPE"[exp])
	}

	var showOLD = false
	var printVal = func(key string, val any, left string, right string) {
		var addSize bool

		// TODO: deleteme after transitioning from FillDbg()
		if !showOLD && strings.Contains(key, "OLD:") {
			return
		}

		if strings.HasPrefix(key, "Size:") {
			addSize = true
		} else if strings.HasPrefix(key, "DAT1:") || strings.HasPrefix(key, "DAT2:") {
			addSize = strings.Contains(key, ":Size")
		}

		if addSize {
			// just give me my goddamn number, i don't care how
			// in case of any errors, size simply wont't be humanized, not a big deal
			if val64, errP := strconv.ParseInt(fmt.Sprintf("%d", val), 10, 64); errP == nil && val64 > 1024 {
				right += " = " + sizeHuman(val64)
			}
		}

		fmt.Println(left, "=", right)
	}

	fmt.Printf("DAT%d [%s]\n", datFile.GetGame(), datName)
	datFile.GetDbg().Dump("", "", printVal)

	for _, dir := range datFile.GetDirs() {
		fmt.Printf(" DIR [%s]\n", dir.GetPath())
		dir.GetDbg().Dump("", "  ", printVal)

		for _, file := range dir.GetFiles() {
			fmt.Printf("  FILE [%s]\n", file.GetPath())
			file.GetDbg().Dump("", "   ", printVal)
		}
	}

	return nil
}
