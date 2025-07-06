package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
	GoVersion = runtime.Version()
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `Display version information for easy-test including version number,
git commit, build date, and Go version.

Use --verbose for detailed build information.`,
	Run: showVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func showVersion(cmd *cobra.Command, args []string) {
	if GetVerbose() {
		fmt.Printf("easy-test version information:\n")
		fmt.Printf("  Version:    %s\n", Version)
		fmt.Printf("  Git Commit: %s\n", GitCommit)
		fmt.Printf("  Build Date: %s\n", BuildDate)
		fmt.Printf("  Go Version: %s\n", GoVersion)
		fmt.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	} else {
		fmt.Printf("easy-test %s\n", Version)
	}
}

