package cmd

import (
	"fmt"
	"log"

	gestures "github.com/ThatOtherAndrew/Hexecute/internal/gesture"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered gestures",
	Run:   listGestures,
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func listGestures(cmd *cobra.Command, args []string) {

	gestures, err := gestures.LoadGestures()
	if err != nil {
		log.Fatal("Failed to load gestures:", err)
	}
	if len(gestures) == 0 {
		fmt.Println("No gestures registered")
	} else {
		fmt.Println("Registered gestures:")
		for _, g := range gestures {
			fmt.Println("  ", g.Command)
		}
	}
}
