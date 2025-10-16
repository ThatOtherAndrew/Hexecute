package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ThatOtherAndrew/Hexecute/internal/config"
	gestures "github.com/ThatOtherAndrew/Hexecute/internal/gesture"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [gesture]",
	Short: "Remove a gesture by command name",
	Run:   removeGesture,
}

func init() {
	rootCmd.AddCommand(removeCmd)
	log.SetFlags(0)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func removeGesture(cmd *cobra.Command, args []string) {

	if len(args) <= 0 {
		log.Fatalf("Please specify a gesture")
	}
	gestures, err := gestures.LoadGestures()
	if err != nil {
		log.Fatal("Failed to load gestures:", err)
	}

	found := false
	for i, g := range gestures {
		if g.Command == args[0] {
			gestures = append(gestures[:i], gestures[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		log.Fatalf("Gesture not found: %s", args[0])
	}

	configFile, err := config.GetPath()
	if err != nil {
		log.Fatal("Failed to get config path:", err)
	}

	data, err := json.Marshal(gestures)
	if err != nil {
		log.Fatal("Failed to marshal gestures:", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		log.Fatal("Failed to save gestures:", err)
	}

	fmt.Println("Removed gesture:", args[0])
}
