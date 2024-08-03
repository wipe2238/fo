package cmd

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"text/tabwriter"
)

func Version() string {
	var buff = strings.Builder{}
	var writer = new(tabwriter.Writer)
	writer.Init(&buff, 0, 0, 2, ' ', 0 /*tabwriter.Debug*/)

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		var getInfo = func(key string) string {
			for _, setting := range buildInfo.Settings {
				if setting.Key == key {
					return setting.Value
				}
			}

			return ""
		}

		fmt.Fprintf(writer, "Repository\thttps://%s/\n", buildInfo.Main.Path)

		fmt.Fprintln(writer, "\t")

		if buildInfo.Main.Version != "" {
			fmt.Fprintf(writer, "GitVersion\t%s\n", buildInfo.Main.Version)
		}

		if val := getInfo("vcs.revision"); val != "" {
			fmt.Fprintf(writer, "GitCommitSHA\t%s\n", val)
		}

		if val := getInfo("vcs.time"); val != "" {
			fmt.Fprintf(writer, "GitCommitDate\t%s\n", val)
		}

		if val := getInfo("vcs.modified"); val == "true" {
			fmt.Fprintln(writer, "GitState\tdirty")
		}

		fmt.Fprintln(writer, "\t")
	}

	fmt.Fprintf(writer, "GoVersion\t%s\n", runtime.Version())
	fmt.Fprintf(writer, "GoCompiler\t%s\n", runtime.Compiler)
	fmt.Fprintf(writer, "GoPlatform\t%s/%s\n", runtime.GOOS, runtime.GOARCH)

	writer.Flush()

	return buff.String()
}
