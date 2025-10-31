package wayland

/*
#cgo pkg-config: wayland-client wayland-egl egl gl xkbcommon
#cgo LDFLAGS: -lwayland-client -lwayland-egl -lEGL -lGL -lxkbcommon
#cgo CFLAGS: -I.
#include "wayland.h"
#include "wayland-client.h"
*/
import "C"
import (
	"fmt"
)

type WaylandError struct {
	msg string
}

func (e *WaylandError) Error() string {
	return e.msg
}

type WaylandWindow struct {
	display       *C.struct_wl_display
	registry      *C.struct_wl_registry
	surface       *C.struct_wl_surface
	layerSurface  *C.struct_zwlr_layer_surface_v1
	eglWindow     *C.struct_wl_egl_window
	eglDisplay    C.EGLDisplay
	eglContext    C.EGLContext
	eglSurface    C.EGLSurface
	width, height int32
}

func NewWaylandWindow() (*WaylandWindow, error) {
	w := &WaylandWindow{}

	C.xkb_context = C.xkb_context_new(C.XKB_CONTEXT_NO_FLAGS)
	if C.xkb_context == nil {
		return nil, &WaylandError{"failed to create xkb context"}
	}

	w.display = C.wl_display_connect(nil)
	if w.display == nil {
		return nil, &WaylandError{"failed to connect to Wayland display"}
	}

	w.registry = C.get_registry(w.display)
	C.add_registry_listener(w.registry)
	C.wl_display_roundtrip(w.display)
	if C.compositor == nil {
		return nil, &WaylandError{"compositor not available"}
	}
	if C.layer_shell == nil {
		return nil, &WaylandError{"layer shell not available"}
	}

	w.surface = C.wl_compositor_create_surface(C.compositor)
	if w.surface == nil {
		return nil, &WaylandError{"failed to create surface"}
	}

	w.layerSurface = C.create_layer_surface(w.surface)

	C.wl_display_roundtrip(w.display)

	var width, height C.int32_t
	C.get_dimensions(&width, &height)
	w.width = int32(width)
	w.height = int32(height)

	if w.width == 0 || w.height == 0 {
		w.width = 1920
		w.height = 1080
	}

	C.wl_display_roundtrip(w.display)

	C.set_input_region(C.int32_t(w.width), C.int32_t(w.height))

	if err := w.initEGL(); err != nil {
		return nil, err
	}

	C.wl_surface_commit(w.surface)
	C.wl_display_flush(w.display)

	C.wl_display_roundtrip(w.display)
	C.wl_display_roundtrip(w.display)
	C.wl_display_flush(w.display)

	return w, nil
}

func (w *WaylandWindow) initEGL() error {
	w.eglWindow = C.wl_egl_window_create(w.surface, C.int(w.width), C.int(w.height))
	if w.eglWindow == nil {
		return fmt.Errorf("failed to create EGL window")
	}

	w.eglDisplay = C.eglGetDisplay(C.EGLNativeDisplayType(w.display))
	if w.eglDisplay == C.EGLDisplay(C.EGL_NO_DISPLAY) {
		return fmt.Errorf("failed to get EGL display")
	}

	var major, minor C.EGLint
	if C.eglInitialize(w.eglDisplay, &major, &minor) == C.EGL_FALSE {
		return fmt.Errorf("failed to initialize EGL")
	}

	configAttribs := []C.EGLint{
		C.EGL_SURFACE_TYPE, C.EGL_WINDOW_BIT,
		C.EGL_RED_SIZE, 8,
		C.EGL_GREEN_SIZE, 8,
		C.EGL_BLUE_SIZE, 8,
		C.EGL_ALPHA_SIZE, 8,
		C.EGL_RENDERABLE_TYPE, C.EGL_OPENGL_BIT,
		C.EGL_NONE,
	}

	var config C.EGLConfig
	var numConfigs C.EGLint
	if C.eglChooseConfig(w.eglDisplay, &configAttribs[0], &config, 1, &numConfigs) == C.EGL_FALSE {
		return fmt.Errorf("failed to choose EGL config")
	}

	C.eglBindAPI(C.EGL_OPENGL_API)
	contextAttribs := []C.EGLint{
		C.EGL_CONTEXT_MAJOR_VERSION, 4,
		C.EGL_CONTEXT_MINOR_VERSION, 1,
		C.EGL_CONTEXT_OPENGL_PROFILE_MASK, C.EGL_CONTEXT_OPENGL_CORE_PROFILE_BIT,
		C.EGL_NONE,
	}

	w.eglContext = C.eglCreateContext(w.eglDisplay, config, nil, &contextAttribs[0])
	if w.eglContext == nil {
		return fmt.Errorf("failed to create EGL context")
	}

	w.eglSurface = C.eglCreateWindowSurface(
		w.eglDisplay,
		config,
		C.native_window(w.eglWindow),
		nil,
	)
	if w.eglSurface == nil {
		return fmt.Errorf("failed to create EGL surface")
	}

	if C.eglMakeCurrent(w.eglDisplay, w.eglSurface, w.eglSurface, w.eglContext) == C.EGL_FALSE {
		return fmt.Errorf("failed to make EGL context current")
	}

	return nil
}

func (w *WaylandWindow) GetSize() (int, int) {
	var width, height C.int32_t
	C.get_dimensions(&width, &height)
	if width > 0 && height > 0 {
		w.width = int32(width)
		w.height = int32(height)
	}
	return int(w.width), int(w.height)
}

func (w *WaylandWindow) ShouldClose() bool {
	return false
}

func (w *WaylandWindow) SwapBuffers() {
	C.eglSwapBuffers(w.eglDisplay, w.eglSurface)
}

func (w *WaylandWindow) PollEvents() {
	C.wl_display_flush(w.display)
	C.wl_display_dispatch_pending(w.display)
}

func (w *WaylandWindow) GetCursorPos() (float64, float64) {
	var x, y C.double
	C.get_mouse_pos(&x, &y)
	return float64(x), float64(y)
}

func (w *WaylandWindow) GetMouseButton() bool {
	state := C.get_button_state()
	return state == 1
}

func (w *WaylandWindow) DisableInput() {
	C.disable_all_input()
}

func (w *WaylandWindow) GetLastKey() (uint32, uint32, bool) {
	key := uint32(C.get_last_key())
	state := uint32(C.get_last_key_state())
	return key, state, key != 0
}

func (w *WaylandWindow) ClearLastKey() {
	C.clear_last_key()
}

func (w *WaylandWindow) Destroy() {
	if w.eglContext != C.EGLContext(C.EGL_NO_CONTEXT) {
		C.eglDestroyContext(w.eglDisplay, w.eglContext)
	}
	if w.eglSurface != C.EGLSurface(C.EGL_NO_SURFACE) {
		C.eglDestroySurface(w.eglDisplay, w.eglSurface)
	}
	if w.eglWindow != nil {
		C.wl_egl_window_destroy(w.eglWindow)
	}
	if w.eglDisplay != C.EGLDisplay(C.EGL_NO_DISPLAY) {
		C.eglTerminate(w.eglDisplay)
	}
	if w.surface != nil {
		C.wl_surface_destroy(w.surface)
	}
	if w.display != nil {
		C.wl_display_disconnect(w.display)
	}
}
