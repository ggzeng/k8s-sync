package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	BuildTS   = "None"
	GitHash   = ""
	GitBranch = "None"
	Version   = "None"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version",
	Long:  `print version`,
	Run: func(cmd *cobra.Command, args []string) {
		PrintFullVersionInfo()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// GetVersion prints build version.
func GetVersion() string {
	if GitHash != "" {
		h := GitHash
		if len(h) > 7 {
			h = h[:7]
		}
		return fmt.Sprintf("%s-%s", GitBranch, h)
	}
	return Version
}

func PrintFullVersionInfo() {
	fmt.Println("Version:          ", GetVersion())
	fmt.Println("Git Branch:       ", GitBranch)
	fmt.Println("Git Commit:       ", GitHash)
	fmt.Println("Build Time (UTC): ", BuildTS)
	fmt.Println("")
}
