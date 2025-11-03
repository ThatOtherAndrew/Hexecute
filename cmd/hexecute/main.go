package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/ThatOtherAndrew/Hexecute/internal/config"
	"github.com/ThatOtherAndrew/Hexecute/internal/draw"
	"github.com/ThatOtherAndrew/Hexecute/internal/execute"
	gestures "github.com/ThatOtherAndrew/Hexecute/internal/gesture"
	"github.com/ThatOtherAndrew/Hexecute/internal/models"
	"github.com/ThatOtherAndrew/Hexecute/internal/opengl"
	"github.com/ThatOtherAndrew/Hexecute/internal/spawn"
	"github.com/ThatOtherAndrew/Hexecute/internal/stroke"
	"github.com/ThatOtherAndrew/Hexecute/internal/update"
	"github.com/ThatOtherAndrew/Hexecute/pkg/wayland"
	"github.com/go-gl/gl/v4.1-core/gl"
)

func init() {
	runtime.LockOSThread()
}

type App struct {
	*models.App
}

func main() {
	learnCommand := flag.String("learn", "", "Learn a new gesture for the specified command")
	listGestures := flag.Bool("list", false, "List all registered gestures")
	removeGesture := flag.String("remove", "", "Remove a gesture by command name")
	flag.Parse()

	if flag.NArg() > 0 {
		log.Fatalf("Unknown arguments: %v", flag.Args())
	}

	if *listGestures {
		gestures, err := gestures.LoadGestures()
		if err != nil {
			log.Fatal("Failed to load gestures:", err)
		}
		if len(gestures) == 0 {
			println("No gestures registered")
		} else {
			println("Registered gestures:")
			for _, g := range gestures {
				println("  ", g.Command)
			}
		}
		return
	}

	if *removeGesture != "" {
		gestures, err := gestures.LoadGestures()
		if err != nil {
			log.Fatal("Failed to load gestures:", err)
		}

		found := false
		for i, g := range gestures {
			if g.Command == *removeGesture {
				gestures = append(gestures[:i], gestures[i+1:]...)
				found = true
				break
			}
		}

		if !found {
			log.Fatalf("Gesture not found: %s", *removeGesture)
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

		println("Removed gesture:", *removeGesture)
		return
	}

	window, err := wayland.NewWaylandWindow()
	if err != nil {
		log.Fatal("Failed to create Wayland window:", err)
	}
	defer window.Destroy()

	settings, err := config.LoadSettings()
	if err != nil {
		log.Fatal("Failed to load settings:", err)
	}

	app := &models.App{
		StartTime: time.Now(),
		Settings:  settings,
	}

	if *learnCommand != "" {
		app.LearnMode = true
		app.LearnCommand = *learnCommand
		log.Printf("Learn mode: Draw the gesture 3 times for command '%s'", *learnCommand)
	} else {
		gestures, err := gestures.LoadGestures()
		if err != nil {
			log.Fatal("Failed to load gestures:", err)
		}
		app.SavedGestures = gestures
		log.Printf("Loaded %d gesture(s)", len(gestures))
	}

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

			if app.LearnMode && len(app.Points) > 0 {
				processed := stroke.ProcessStroke(app.Points)
				app.LearnGestures = append(app.LearnGestures, processed)
				app.LearnCount++
				log.Printf("Captured gesture %d/3", app.LearnCount)

				app.Points = nil

				if app.LearnCount >= 3 {
					if err := gestures.SaveGesture(app.LearnCommand, app.LearnGestures); err != nil {
						log.Fatal("Failed to save gesture:", err)
					}
					log.Printf("Gesture saved for command: %s", app.LearnCommand)

					app.IsExiting = true
					app.ExitStartTime = time.Now()
					window.DisableInput()
					x, y := window.GetCursorPos()
					spawn := spawn.New(app)
					spawn.SpawnExitWisps(float32(x), float32(y))
				}
			} else if !app.LearnMode && len(app.Points) > 0 {
				x, y := window.GetCursorPos()
				exec := execute.New(app)
				exec.RecognizeAndExecute(window, float32(x), float32(y))
				app.Points = nil
			}
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
