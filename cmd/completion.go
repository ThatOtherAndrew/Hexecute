package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:                   "completion [bash|zsh|fish]",
	Short:                 "Get completions for your favorite shell",
	RunE:                  GetCompletion,
	ValidArgs:             []string{"bash", "zsh", "fish"},
	DisableFlagsInUseLine: true,
}

func init() {
	rootCmd.AddCommand(completionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// completionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// completionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func GetCompletion(cmd *cobra.Command, args []string) error {
	if len(args) <= 0 {
		return cmd.Help()
	}

	switch args[0] {
	case "bash":
		return rootCmd.GenBashCompletionV2(os.Stdout, true)
	case "zsh":
		return rootCmd.GenZshCompletion(os.Stdout)
	case "fish":
		return rootCmd.GenFishCompletion(os.Stdout, true)
	default:
		return fmt.Errorf("unsupported shell: %s", args[0])
	}
}
