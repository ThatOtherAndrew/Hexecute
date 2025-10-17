package execute

import (
	"log"
	"os/exec"
	"syscall"
	"time"

	"github.com/ThatOtherAndrew/Hexecute/internal/models"
	"github.com/ThatOtherAndrew/Hexecute/internal/spawn"
	"github.com/ThatOtherAndrew/Hexecute/internal/stroke"
	"github.com/ThatOtherAndrew/Hexecute/pkg/wayland"
)

type App struct {
	app *models.App
}

func New(app *models.App) *App {
	return &App{app: app}
}

func Command(command string) error {
	if command == "" {
		return nil
	}

	cmd := exec.Command("sh", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	return cmd.Start()
}

func (a *App) RecognizeAndExecute(window *wayland.WaylandWindow, x, y float32) {
	if len(a.app.Points) < 5 {
		log.Println("Gesture too short, ignoring")
		return
	}

	processed := stroke.ProcessStroke(a.app.Points)

	bestMatch := -1
	bestScore := 0.0

	for i, gesture := range a.app.SavedGestures {
		match, score := stroke.UnistrokeRecognise(processed, gesture.Templates)
		log.Printf("Gesture %d (%s): template %d, score %.3f", i, gesture.Command, match, score)

		if score > bestScore {
			bestScore = score
			bestMatch = i
		}
	}

	if bestMatch >= 0 && bestScore > 0.6 {
		command := a.app.SavedGestures[bestMatch].Command
		log.Printf("Matched gesture: %s (score: %.3f)", command, bestScore)

		if err := Command(command); err != nil {
			log.Printf("Failed to execute command: %v", err)
		} else {
			log.Printf("Executed: %s", command)
		}

		a.app.IsExiting = true
		a.app.ExitStartTime = time.Now()
		window.DisableInput()
		spawn := spawn.New(a.app)
		spawn.SpawnExitWisps(x, y)
	} else {
		log.Printf("No confident match (best score: %.3f)", bestScore)
	}
}
