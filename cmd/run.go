package cmd

import (
	"log"
	"runtime"
	"time"

	"github.com/ThatOtherAndrew/Hexecute/internal/draw"
	"github.com/ThatOtherAndrew/Hexecute/internal/execute"
	gestures "github.com/ThatOtherAndrew/Hexecute/internal/gesture"
	"github.com/ThatOtherAndrew/Hexecute/internal/models"
	"github.com/ThatOtherAndrew/Hexecute/internal/opengl"
	"github.com/ThatOtherAndrew/Hexecute/internal/spawn"
	"github.com/ThatOtherAndrew/Hexecute/internal/update"
	"github.com/ThatOtherAndrew/Hexecute/pkg/wayland"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run the program",
	Run:   Run,
}

func init() {
	rootCmd.AddCommand(runCmd)
	runtime.LockOSThread()

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func Run(cmd *cobra.Command, args []string) {
	window, err := wayland.NewWaylandWindow()
	if err != nil {
		log.Fatal("Failed to create Wayland window:", err)
	}
	defer window.Destroy()

	app := &models.App{StartTime: time.Now()}

	gesture, err := gestures.LoadGestures()
	if err != nil {
		log.Fatal("Failed to load gestures:", err)
	}
	app.SavedGestures = gesture
	log.Printf("Loaded %d gesture(s)", len(gesture))

	opengl := opengl.New(app)
	if err := opengl.InitGL(); err != nil {
		log.Fatal("Failed to initialize OpenGL:", err)
	}

	gl.ClearColor(0, 0, 0, 0)

	for range 5 {
		window.PollEvents()
		gl.Clear(gl.COLOR_BUFFER_BIT)
		window.SwapBuffers()
	}

	x, y := window.GetCursorPos()
	app.LastCursorX = float32(x)
	app.LastCursorY = float32(y)

	lastTime := time.Now()
	var wasPressed bool

	for !window.ShouldClose() {
		now := time.Now()
		dt := float32(now.Sub(lastTime).Seconds())
		lastTime = now

		window.PollEvents()
		update := update.New(app)
		update.UpdateCursor(window)

		if key, state, hasKey := window.GetLastKey(); hasKey {
			if state == 1 && key == 0xff1b {
				if !app.IsExiting {
					app.IsExiting = true
					app.ExitStartTime = time.Now()
					window.DisableInput()
					x, y := window.GetCursorPos()
					spawn := spawn.New(app)
					spawn.SpawnExitWisps(float32(x), float32(y))
				}
			}
			window.ClearLastKey()
		}

		if app.IsExiting {
			if time.Since(app.ExitStartTime).Seconds() > 0.8 {
				break
			}
		}
		isPressed := window.GetMouseButton()
		if isPressed && !wasPressed {
			app.IsDrawing = true
		} else if !isPressed && wasPressed {
			app.IsDrawing = false

			x, y := window.GetCursorPos()
			exec := execute.New(app)
			exec.RecognizeAndExecute(window, float32(x), float32(y))
			app.Points = nil
		}
		wasPressed = isPressed

		if app.IsDrawing {
			x, y := window.GetCursorPos()
			gesture := gestures.New(app)
			gesture.AddPoint(float32(x), float32(y))

			spawn := spawn.New(app)
			spawn.SpawnCursorSparkles(float32(x), float32(y))
		}

		update.UpdateParticles(dt)
		drawer := draw.New(app)
		drawer.Draw(window)
		window.SwapBuffers()
	}

}
