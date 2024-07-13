package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wipe2238/fo/cmd"
	"github.com/wipe2238/fo/dat"
)

func init() {
	var cmd = &cobra.Command{
		Use:   "dump [dat file]",
		Short: "Dump DAT file data (only DAT1 supported right now)",

		GroupID: app.GroupID,
		Args:    cobra.ExactArgs(1),
		RunE:    func(cmd *cobra.Command, args []string) error { return argDump(cmd, args) },
	}

	app.AddCommand(cmd)
}

func argDump(arg *cobra.Command, args []string) (err error) {
	if err = cmd.ResolveFilename(&args[0], "@"); err != nil {
		return err
	}

	var osFile *os.File
	if osFile, err = os.Open(args[0]); err == nil {
		defer osFile.Close()
	} else {
		return err
	}

	var datFile dat.FalloutDat
	if datFile, err = dat.Fallout1(osFile); err != nil {
		return err
	}

	return doDump(arg, datFile)
}

func doDump(_ *cobra.Command, datFile dat.FalloutDat) (err error) {
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

	var printVal = func(key string, val any, left string, right string) {

		if strings.HasPrefix(key, "Size:") || (strings.HasPrefix(key, "DAT1:") && strings.Contains(key, ":Size")) {
			// just give me my goddamn number, i don't care how
			// in case of any errors, size simply wont't be humanized, not a big deal
			if val64, errP := strconv.ParseInt(fmt.Sprintf("%d", val), 10, 64); errP == nil && val64 > 1024 {
				right += " = " + sizeHuman(val64)
			}
		}

		fmt.Println(left, "=", right)
	}

	datFile.FillDbg()

	if datFile.GetDbg().KeysMaxLen("") > 0 {
		fmt.Printf("DAT%d\n", datFile.GetGame())
		datFile.GetDbg().Dump("", "", printVal)
	}

	for _, dir := range datFile.GetDirs() {
		if dir.GetDbg().KeysMaxLen("") > 0 {
			fmt.Printf(" DIR [%s]\n", dir.GetPath())
			dir.GetDbg().Dump("", "  ", printVal)
		}

		for _, file := range dir.GetFiles() {
			if file.GetDbg().KeysMaxLen("") > 0 {
				fmt.Printf("  FILE [%s]\n", file.GetPath())
				file.GetDbg().Dump("", "   ", printVal)
			}
		}
	}

	return nil
}
