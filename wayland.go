package main

/*
#cgo pkg-config: wayland-client wayland-egl egl gl
#cgo LDFLAGS: -lwayland-client -lwayland-egl -lEGL -lGL
#cgo CFLAGS: -I.
#include <wayland-client.h>
#include <wayland-egl.h>
#include <EGL/egl.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include "wlr-layer-shell-client.h"

// Include protocol implementation inline
#include <stdbool.h>
#include <stdint.h>

#ifndef __has_attribute
# define __has_attribute(x) 0
#endif

#if (__has_attribute(visibility) || defined(__GNUC__) && __GNUC__ >= 4)
#define WL_PRIVATE __attribute__ ((visibility("hidden")))
#else
#define WL_PRIVATE
#endif

extern const struct wl_interface wl_output_interface;
extern const struct wl_interface wl_surface_interface;
extern const struct wl_interface zwlr_layer_surface_v1_interface;

// Stub for xdg_popup_interface (not actually used but referenced in types array)
static const struct wl_interface xdg_popup_interface = {
	"xdg_popup", 0, 0, NULL, 0, NULL,
};

static const struct wl_interface *wlr_layer_shell_unstable_v1_types[] = {
	NULL,
	NULL,
	NULL,
	NULL,
	&zwlr_layer_surface_v1_interface,
	&wl_surface_interface,
	&wl_output_interface,
	NULL,
	NULL,
	&xdg_popup_interface,
};

static const struct wl_message zwlr_layer_shell_v1_requests[] = {
	{ "get_layer_surface", "no?ous", wlr_layer_shell_unstable_v1_types + 4 },
	{ "destroy", "3", wlr_layer_shell_unstable_v1_types + 0 },
};

WL_PRIVATE const struct wl_interface zwlr_layer_shell_v1_interface = {
	"zwlr_layer_shell_v1", 4,
	2, zwlr_layer_shell_v1_requests,
	0, NULL,
};

static const struct wl_message zwlr_layer_surface_v1_requests[] = {
	{ "set_size", "uu", wlr_layer_shell_unstable_v1_types + 0 },
	{ "set_anchor", "u", wlr_layer_shell_unstable_v1_types + 0 },
	{ "set_exclusive_zone", "i", wlr_layer_shell_unstable_v1_types + 0 },
	{ "set_margin", "iiii", wlr_layer_shell_unstable_v1_types + 0 },
	{ "set_keyboard_interactivity", "u", wlr_layer_shell_unstable_v1_types + 0 },
	{ "get_popup", "o", wlr_layer_shell_unstable_v1_types + 9 },
	{ "ack_configure", "u", wlr_layer_shell_unstable_v1_types + 0 },
	{ "destroy", "", wlr_layer_shell_unstable_v1_types + 0 },
	{ "set_layer", "2u", wlr_layer_shell_unstable_v1_types + 0 },
};

static const struct wl_message zwlr_layer_surface_v1_events[] = {
	{ "configure", "uuu", wlr_layer_shell_unstable_v1_types + 0 },
	{ "closed", "", wlr_layer_shell_unstable_v1_types + 0 },
};

WL_PRIVATE const struct wl_interface zwlr_layer_surface_v1_interface = {
	"zwlr_layer_surface_v1", 4,
	9, zwlr_layer_surface_v1_requests,
	2, zwlr_layer_surface_v1_events,
};

// Globals
struct wl_compositor *compositor = NULL;
struct zwlr_layer_shell_v1 *layer_shell = NULL;
struct wl_seat *seat = NULL;
struct wl_pointer *pointer = NULL;
int32_t width_global = 0;
int32_t height_global = 0;

// Callback for layer surface configure
void layer_surface_configure(void *data, struct zwlr_layer_surface_v1 *surface,
                             uint32_t serial, uint32_t width, uint32_t height) {
    width_global = width;
    height_global = height;
    zwlr_layer_surface_v1_ack_configure(surface, serial);
}

void layer_surface_closed(void *data, struct zwlr_layer_surface_v1 *surface) {
}

static struct zwlr_layer_surface_v1_listener layer_surface_listener = {
    .configure = layer_surface_configure,
    .closed = layer_surface_closed,
};

// Forward declarations for seat
void seat_capabilities(void *data, struct wl_seat *seat, uint32_t capabilities);
void seat_name(void *data, struct wl_seat *seat, const char *name);

static const struct wl_seat_listener seat_listener = {
    .capabilities = seat_capabilities,
    .name = seat_name,
};

// Registry listener
static void registry_global(void *data, struct wl_registry *registry,
                           uint32_t name, const char *interface,
                           uint32_t version) {
    printf("Registry global: %s (version %d)\n", interface, version);
    fflush(stdout);
    if (strcmp(interface, "wl_compositor") == 0) {
        compositor = wl_registry_bind(registry, name, &wl_compositor_interface, 4);
        printf("Compositor bound\n");
        fflush(stdout);
    } else if (strcmp(interface, "zwlr_layer_shell_v1") == 0) {
        layer_shell = (struct zwlr_layer_shell_v1 *)
            wl_registry_bind(registry, name, &zwlr_layer_shell_v1_interface, 1);
        printf("Layer shell bound\n");
        fflush(stdout);
    } else if (strcmp(interface, "wl_seat") == 0) {
        seat = wl_registry_bind(registry, name, &wl_seat_interface, 1);
        printf("Seat bound\n");
        fflush(stdout);
        // Add listener immediately to catch capabilities event
        wl_seat_add_listener(seat, &seat_listener, NULL);
        printf("Seat listener added in registry callback\n");
        fflush(stdout);
    }
}

static void registry_global_remove(void *data, struct wl_registry *registry,
                                   uint32_t name) {
}

static const struct wl_registry_listener registry_listener = {
    .global = registry_global,
    .global_remove = registry_global_remove,
};

// Helper functions
struct wl_registry *get_registry(struct wl_display *display) {
    return wl_display_get_registry(display);
}

void add_registry_listener(struct wl_registry *registry) {
    wl_registry_add_listener(registry, &registry_listener, NULL);
}

struct wl_surface *surface_global = NULL;

struct zwlr_layer_surface_v1 *create_layer_surface(struct wl_surface *surface) {
    surface_global = surface;

    struct zwlr_layer_surface_v1 *layer_surface =
        zwlr_layer_shell_v1_get_layer_surface(
            layer_shell, surface, NULL,
            ZWLR_LAYER_SHELL_V1_LAYER_OVERLAY, "hexecute");

    // Configure as fullscreen transparent overlay
    zwlr_layer_surface_v1_set_anchor(layer_surface,
        ZWLR_LAYER_SURFACE_V1_ANCHOR_TOP |
        ZWLR_LAYER_SURFACE_V1_ANCHOR_BOTTOM |
        ZWLR_LAYER_SURFACE_V1_ANCHOR_LEFT |
        ZWLR_LAYER_SURFACE_V1_ANCHOR_RIGHT);

    zwlr_layer_surface_v1_set_exclusive_zone(layer_surface, -1);

    // Don't capture keyboard - we only need mouse input
    zwlr_layer_surface_v1_set_keyboard_interactivity(layer_surface,
        ZWLR_LAYER_SURFACE_V1_KEYBOARD_INTERACTIVITY_NONE);

    zwlr_layer_surface_v1_add_listener(layer_surface, &layer_surface_listener, NULL);

    wl_surface_commit(surface);

    return layer_surface;
}

void set_input_region(int32_t width, int32_t height) {
    if (surface_global) {
        printf("Setting input region: %dx%d\n", width, height);
        fflush(stdout);
        // Create input region covering the full surface to capture all input
        struct wl_region *region = wl_compositor_create_region(compositor);
        wl_region_add(region, 0, 0, width, height);
        wl_surface_set_input_region(surface_global, region);
        wl_region_destroy(region);
        wl_surface_commit(surface_global);
    }
}

// Pointer listener
static int button_state = 0;
static double mouse_x = 0;
static double mouse_y = 0;

void pointer_enter(void *data, struct wl_pointer *pointer, uint32_t serial,
                  struct wl_surface *surface, wl_fixed_t x, wl_fixed_t y) {
    mouse_x = wl_fixed_to_double(x);
    mouse_y = wl_fixed_to_double(y);
    printf("Pointer enter: x=%f, y=%f\n", mouse_x, mouse_y);
    fflush(stdout);
}

void pointer_leave(void *data, struct wl_pointer *pointer, uint32_t serial,
                  struct wl_surface *surface) {
    printf("Pointer leave\n");
    fflush(stdout);
}

void pointer_motion(void *data, struct wl_pointer *pointer, uint32_t time,
                   wl_fixed_t x, wl_fixed_t y) {
    mouse_x = wl_fixed_to_double(x);
    mouse_y = wl_fixed_to_double(y);
}

void pointer_button(void *data, struct wl_pointer *pointer, uint32_t serial,
                   uint32_t time, uint32_t button, uint32_t state) {
    printf("Pointer button: button=%d, state=%d\n", button, state);
    fflush(stdout);
    if (button == 272) { // BTN_LEFT
        button_state = state;
    }
}

void pointer_axis(void *data, struct wl_pointer *pointer, uint32_t time,
                 uint32_t axis, wl_fixed_t value) {
}

void pointer_frame(void *data, struct wl_pointer *pointer) {
}

void pointer_axis_source(void *data, struct wl_pointer *pointer, uint32_t source) {
}

void pointer_axis_stop(void *data, struct wl_pointer *pointer, uint32_t time, uint32_t axis) {
}

void pointer_axis_discrete(void *data, struct wl_pointer *pointer, uint32_t axis, int32_t discrete) {
}

static const struct wl_pointer_listener pointer_listener = {
    .enter = pointer_enter,
    .leave = pointer_leave,
    .motion = pointer_motion,
    .button = pointer_button,
    .axis = pointer_axis,
    .frame = pointer_frame,
    .axis_source = pointer_axis_source,
    .axis_stop = pointer_axis_stop,
    .axis_discrete = pointer_axis_discrete,
};

// Seat listener
void seat_capabilities(void *data, struct wl_seat *seat, uint32_t capabilities) {
    printf("Seat capabilities: %d (pointer=%d)\n", capabilities, (capabilities & WL_SEAT_CAPABILITY_POINTER) != 0);
    fflush(stdout);
    if (capabilities & WL_SEAT_CAPABILITY_POINTER) {
        pointer = wl_seat_get_pointer(seat);
        wl_pointer_add_listener(pointer, &pointer_listener, NULL);
        printf("Pointer listener added\n");
        fflush(stdout);
    }
}

void seat_name(void *data, struct wl_seat *seat, const char *name) {
}

void setup_seat_listener() {
    printf("setup_seat_listener called, seat=%p\n", seat);
    fflush(stdout);
    if (seat) {
        wl_seat_add_listener(seat, &seat_listener, NULL);
        printf("Seat listener added\n");
        fflush(stdout);
    } else {
        printf("ERROR: No seat available!\n");
        fflush(stdout);
    }
}

int get_button_state() {
    return button_state;
}

void get_mouse_pos(double *x, double *y) {
    *x = mouse_x;
    *y = mouse_y;
}

void get_dimensions(int32_t *w, int32_t *h) {
    *w = width_global;
    *h = height_global;
}
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

	// Connect to Wayland display
	w.display = C.wl_display_connect(nil)
	if w.display == nil {
		return nil, &WaylandError{"failed to connect to Wayland display"}
	}

	// Get registry and add listener
	w.registry = C.get_registry(w.display)
	C.add_registry_listener(w.registry)

	// Roundtrip to get globals
	C.wl_display_roundtrip(w.display)

	// Check if we got compositor and layer shell
	if C.compositor == nil {
		return nil, &WaylandError{"compositor not available"}
	}
	if C.layer_shell == nil {
		return nil, &WaylandError{"layer shell not available"}
	}

	// Create surface
	w.surface = C.wl_compositor_create_surface(C.compositor)
	if w.surface == nil {
		return nil, &WaylandError{"failed to create surface"}
	}

	// Create layer surface
	w.layerSurface = C.create_layer_surface(w.surface)

	// Roundtrip to get configure event
	C.wl_display_roundtrip(w.display)

	// Get dimensions
	var width, height C.int32_t
	C.get_dimensions(&width, &height)
	w.width = int32(width)
	w.height = int32(height)

	if w.width == 0 || w.height == 0 {
		// Default to reasonable size if not set
		w.width = 1920
		w.height = 1080
	}

	// Do another roundtrip to receive seat capabilities
	C.wl_display_roundtrip(w.display)

	// Set input region now that we have dimensions
	C.set_input_region(C.int32_t(w.width), C.int32_t(w.height))

	// Initialize EGL
	if err := w.initEGL(); err != nil {
		return nil, err
	}

	// Commit surface after EGL setup to ensure it's ready to receive events
	C.wl_surface_commit(w.surface)
	C.wl_display_flush(w.display)

	return w, nil
}

func (w *WaylandWindow) initEGL() error {
	// Create EGL window
	w.eglWindow = C.wl_egl_window_create(w.surface, C.int(w.width), C.int(w.height))
	if w.eglWindow == nil {
		return fmt.Errorf("failed to create EGL window")
	}

	// Get EGL display
	w.eglDisplay = C.eglGetDisplay(C.EGLNativeDisplayType(w.display))
	if w.eglDisplay == C.EGLDisplay(C.EGL_NO_DISPLAY) {
		return fmt.Errorf("failed to get EGL display")
	}

	// Initialize EGL
	var major, minor C.EGLint
	if C.eglInitialize(w.eglDisplay, &major, &minor) == C.EGL_FALSE {
		return fmt.Errorf("failed to initialize EGL")
	}

	// Configure EGL
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

	// Bind OpenGL API
	C.eglBindAPI(C.EGL_OPENGL_API)

	// Create EGL context
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

	// Create EGL surface
	w.eglSurface = C.eglCreateWindowSurface(w.eglDisplay, config, C.EGLNativeWindowType(w.eglWindow), nil)
	if w.eglSurface == nil {
		return fmt.Errorf("failed to create EGL surface")
	}

	// Make context current
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
	return false // Add proper close handling if needed
}

func (w *WaylandWindow) SwapBuffers() {
	C.eglSwapBuffers(w.eglDisplay, w.eglSurface)
}

func (w *WaylandWindow) PollEvents() {
	// Flush outgoing requests
	C.wl_display_flush(w.display)
	// Dispatch any pending events
	C.wl_display_dispatch_pending(w.display)
}

func (w *WaylandWindow) GetCursorPos() (float64, float64) {
	var x, y C.double
	C.get_mouse_pos(&x, &y)
	return float64(x), float64(y)
}

func (w *WaylandWindow) GetMouseButton() bool {
	state := C.get_button_state()
	return state == 1 // WL_POINTER_BUTTON_STATE_PRESSED
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
