package cmd

import (
	"fmt"

	"github.com/palagend/slowmade/internal/version"
	"github.com/spf13/cobra"
)

// versionCmd 代表 version 命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information of this application",
	Long:  `The version command prints detailed information about the build of this application, including the version number, Git commit, and build environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		v := version.Get()
		fmt.Println(v.String())
	},
}

func init() {
	// 将 versionCmd 添加到根命令 (rootCmd) 下
	rootCmd.AddCommand(versionCmd)
}
