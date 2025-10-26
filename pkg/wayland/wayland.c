#include "wayland.h"
#include "keyboard-shortcuts-inhibit-client.h"
#include "tablet-v2.h"
#include "wlr-layer-shell-client.h"
#include <stdbool.h>
#include <stdint.h>
#include <string.h>
#include <sys/mman.h>
#include <unistd.h>

#ifndef __has_attribute
#define __has_attribute(x) 0
#endif

#if (__has_attribute(visibility) || defined(__GNUC__) && __GNUC__ >= 4)
#define WL_PRIVATE __attribute__((visibility("hidden")))
#else
#define WL_PRIVATE
#endif

extern const struct wl_interface wl_output_interface;
extern const struct wl_interface wl_surface_interface;
extern const struct wl_interface zwlr_layer_surface_v1_interface;

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
    {"get_layer_surface", "no?ous", wlr_layer_shell_unstable_v1_types + 4},
    {"destroy", "3", wlr_layer_shell_unstable_v1_types + 0},
};

WL_PRIVATE const struct wl_interface zwlr_layer_shell_v1_interface = {
    "zwlr_layer_shell_v1", 4, 2, zwlr_layer_shell_v1_requests, 0, NULL,
};

static const struct wl_message zwlr_layer_surface_v1_requests[] = {
    {"set_size", "uu", wlr_layer_shell_unstable_v1_types + 0},
    {"set_anchor", "u", wlr_layer_shell_unstable_v1_types + 0},
    {"set_exclusive_zone", "i", wlr_layer_shell_unstable_v1_types + 0},
    {"set_margin", "iiii", wlr_layer_shell_unstable_v1_types + 0},
    {"set_keyboard_interactivity", "u", wlr_layer_shell_unstable_v1_types + 0},
    {"get_popup", "o", wlr_layer_shell_unstable_v1_types + 9},
    {"ack_configure", "u", wlr_layer_shell_unstable_v1_types + 0},
    {"destroy", "", wlr_layer_shell_unstable_v1_types + 0},
    {"set_layer", "2u", wlr_layer_shell_unstable_v1_types + 0},
};

static const struct wl_message zwlr_layer_surface_v1_events[] = {
    {"configure", "uuu", wlr_layer_shell_unstable_v1_types + 0},
    {"closed", "", wlr_layer_shell_unstable_v1_types + 0},
};

WL_PRIVATE const struct wl_interface zwlr_layer_surface_v1_interface = {
    "zwlr_layer_surface_v1",        4, 9,
    zwlr_layer_surface_v1_requests, 2, zwlr_layer_surface_v1_events,
};

static const struct wl_interface
    *keyboard_shortcuts_inhibit_unstable_v1_types[] = {
        &zwp_keyboard_shortcuts_inhibitor_v1_interface,
        &wl_surface_interface,
        &wl_seat_interface,
};

static const struct wl_message
    zwp_keyboard_shortcuts_inhibit_manager_v1_requests[] = {
        {"destroy", "", keyboard_shortcuts_inhibit_unstable_v1_types + 0},
        {"inhibit_shortcuts", "noo",
         keyboard_shortcuts_inhibit_unstable_v1_types + 0},
};

WL_PRIVATE const struct wl_interface
    zwp_keyboard_shortcuts_inhibit_manager_v1_interface = {
        "zwp_keyboard_shortcuts_inhibit_manager_v1",        1, 2,
        zwp_keyboard_shortcuts_inhibit_manager_v1_requests, 0, NULL,
};

static const struct wl_message zwp_keyboard_shortcuts_inhibitor_v1_requests[] =
    {
        {"destroy", "", keyboard_shortcuts_inhibit_unstable_v1_types + 0},
};

static const struct wl_message zwp_keyboard_shortcuts_inhibitor_v1_events[] = {
    {"active", "", keyboard_shortcuts_inhibit_unstable_v1_types + 0},
    {"inactive", "", keyboard_shortcuts_inhibit_unstable_v1_types + 0},
};

WL_PRIVATE const struct wl_interface
    zwp_keyboard_shortcuts_inhibitor_v1_interface = {
        "zwp_keyboard_shortcuts_inhibitor_v1",
        1,
        1,
        zwp_keyboard_shortcuts_inhibitor_v1_requests,
        2,
        zwp_keyboard_shortcuts_inhibitor_v1_events,
};

struct wl_compositor *compositor = NULL;
struct zwlr_layer_shell_v1 *layer_shell = NULL;
struct wl_seat *seat = NULL;
struct wl_pointer *pointer = NULL;
struct wl_touch *touch = NULL;
struct wl_keyboard *keyboard = NULL;
struct zwp_tablet_manager_v2 *tablet_manager = NULL;
struct zwp_tablet_tool_v2 *tablet_tool = NULL;
struct zwp_tablet_seat_v2 *tablet_seat = NULL;
struct zwp_keyboard_shortcuts_inhibit_manager_v1 *shortcuts_inhibit_manager =
    NULL;
struct zwp_keyboard_shortcuts_inhibitor_v1 *shortcuts_inhibitor = NULL;
struct zwlr_layer_surface_v1 *layer_surface_global = NULL;
struct xkb_context *xkb_context;
struct xkb_keymap *xkb_keymap;
struct xkb_state *xkb_state;
int32_t width_global = 0;
int32_t height_global = 0;

void layer_surface_configure(void *data, struct zwlr_layer_surface_v1 *surface,
                             uint32_t serial, uint32_t width, uint32_t height) {
  width_global = width;
  height_global = height;
  zwlr_layer_surface_v1_ack_configure(surface, serial);
}

void layer_surface_closed(void *data, struct zwlr_layer_surface_v1 *surface) {}

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

void registry_global(void *data, struct wl_registry *registry, uint32_t name,
                     const char *interface, uint32_t version) {
  if (strcmp(interface, "wl_compositor") == 0) {
    compositor = wl_registry_bind(registry, name, &wl_compositor_interface, 4);
  } else if (strcmp(interface, "zwlr_layer_shell_v1") == 0) {
    layer_shell = (struct zwlr_layer_shell_v1 *)wl_registry_bind(
        registry, name, &zwlr_layer_shell_v1_interface, 1);
  } else if (strcmp(interface, "wl_seat") == 0) {
    seat = wl_registry_bind(registry, name, &wl_seat_interface, 1);
    wl_seat_add_listener(seat, &seat_listener, NULL);
  } else if (strcmp(interface, "zwp_keyboard_shortcuts_inhibit_manager_v1") ==
             0) {
    shortcuts_inhibit_manager =
        (struct zwp_keyboard_shortcuts_inhibit_manager_v1 *)wl_registry_bind(
            registry, name,
            &zwp_keyboard_shortcuts_inhibit_manager_v1_interface, 1);
  } else if (strcmp(interface, zwp_tablet_manager_v2_interface.name) == 0) {
    tablet_manager = (struct zwp_tablet_manager_v2 *)wl_registry_bind(
        registry, name, &zwp_tablet_manager_v2_interface, 1);
  }
}

void registry_global_remove(void *data, struct wl_registry *registry,
                            uint32_t name) {}

static const struct wl_registry_listener registry_listener = {
    .global = registry_global,
    .global_remove = registry_global_remove,
};

struct wl_registry *get_registry(struct wl_display *display) {
  return wl_display_get_registry(display);
}

void add_registry_listener(struct wl_registry *registry) {
  wl_registry_add_listener(registry, &registry_listener, NULL);
}

struct wl_surface *surface_global = NULL;

struct zwlr_layer_surface_v1 *create_layer_surface(struct wl_surface *surface) {
  surface_global = surface;

  layer_surface_global = zwlr_layer_shell_v1_get_layer_surface(
      layer_shell, surface, NULL, ZWLR_LAYER_SHELL_V1_LAYER_OVERLAY,
      "hexecute");

  zwlr_layer_surface_v1_set_anchor(layer_surface_global,
                                   ZWLR_LAYER_SURFACE_V1_ANCHOR_TOP |
                                       ZWLR_LAYER_SURFACE_V1_ANCHOR_BOTTOM |
                                       ZWLR_LAYER_SURFACE_V1_ANCHOR_LEFT |
                                       ZWLR_LAYER_SURFACE_V1_ANCHOR_RIGHT);

  zwlr_layer_surface_v1_set_exclusive_zone(layer_surface_global, -1);
  zwlr_layer_surface_v1_set_keyboard_interactivity(
      layer_surface_global,
      ZWLR_LAYER_SURFACE_V1_KEYBOARD_INTERACTIVITY_EXCLUSIVE);

  zwlr_layer_surface_v1_add_listener(layer_surface_global,
                                     &layer_surface_listener, NULL);

  wl_surface_commit(surface);

  return layer_surface_global;
}

void set_input_region(int32_t width, int32_t height) {
  if (surface_global) {
    struct wl_region *region = wl_compositor_create_region(compositor);
    wl_region_add(region, 0, 0, width, height);
    wl_surface_set_input_region(surface_global, region);
    wl_region_destroy(region);
    wl_surface_commit(surface_global);
  }
}

void disable_all_input() {
  if (shortcuts_inhibitor) {
    zwp_keyboard_shortcuts_inhibitor_v1_destroy(shortcuts_inhibitor);
    shortcuts_inhibitor = NULL;
  }

  if (layer_surface_global) {
    zwlr_layer_surface_v1_set_keyboard_interactivity(
        layer_surface_global,
        ZWLR_LAYER_SURFACE_V1_KEYBOARD_INTERACTIVITY_NONE);
  }

  if (surface_global) {
    struct wl_region *region = wl_compositor_create_region(compositor);
    wl_surface_set_input_region(surface_global, region);
    wl_region_destroy(region);
    wl_surface_commit(surface_global);
  }
}

static int button_state = 0;
static double mouse_x = 0;
static double mouse_y = 0;
static int32_t touch_id = 255555;

void pointer_enter(void *data, struct wl_pointer *pointer, uint32_t serial,
                   struct wl_surface *surface, wl_fixed_t x, wl_fixed_t y) {
  mouse_x = wl_fixed_to_double(x);
  mouse_y = wl_fixed_to_double(y);
  wl_pointer_set_cursor(pointer, serial, NULL, 0, 0);
}

void pointer_leave(void *data, struct wl_pointer *pointer, uint32_t serial,
                   struct wl_surface *surface) {}

void pointer_motion(void *data, struct wl_pointer *pointer, uint32_t time,
                    wl_fixed_t x, wl_fixed_t y) {
  mouse_x = wl_fixed_to_double(x);
  mouse_y = wl_fixed_to_double(y);
}

void pointer_button(void *data, struct wl_pointer *pointer, uint32_t serial,
                    uint32_t time, uint32_t button, uint32_t state) {
  if (button == 272) {
    button_state = state;
  }
}

void pointer_axis(void *data, struct wl_pointer *pointer, uint32_t time,
                  uint32_t axis, wl_fixed_t value) {}

void pointer_frame(void *data, struct wl_pointer *pointer) {}

void pointer_axis_source(void *data, struct wl_pointer *pointer,
                         uint32_t source) {}

void pointer_axis_stop(void *data, struct wl_pointer *pointer, uint32_t time,
                       uint32_t axis) {}

void pointer_axis_discrete(void *data, struct wl_pointer *pointer,
                           uint32_t axis, int32_t discrete) {}

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

void tablet_tool_removed(void *data, struct zwp_tablet_tool_v2 *id) {
  button_state = 0;
}

void tablet_tool_down(void *data, struct zwp_tablet_tool_v2 *id,
                      unsigned int serial) {
  button_state = 1;
}

void tablet_tool_up(void *data, struct zwp_tablet_tool_v2 *id) {
  button_state = 0;
}

void tablet_tool_motion(void *data, struct zwp_tablet_tool_v2 *id, wl_fixed_t x,
                        wl_fixed_t y) {
  mouse_x = wl_fixed_to_double(x);
  mouse_y = wl_fixed_to_double(y);
}

void tablet_tool_type(void *data, struct zwp_tablet_tool_v2 *id,
                      uint32_t tool_type) {}

void tablet_tool_serial(void *data, struct zwp_tablet_tool_v2 *id,
                        uint32_t high, uint32_t low) {}

void tablet_tool_id_wacom(void *data, struct zwp_tablet_tool_v2 *id,
                          uint32_t high, uint32_t low) {}

void tablet_tool_capability(void *data, struct zwp_tablet_tool_v2 *id,
                            uint32_t capability) {}

void tablet_tool_proximity_in(void *data, struct zwp_tablet_tool_v2 *id,
                              uint32_t serial, struct zwp_tablet_v2 *tablet_id,
                              struct wl_surface *surface) {}

void tablet_tool_proximity_out(void *data, struct zwp_tablet_tool_v2 *id) {}

void tablet_tool_pressure(void *data, struct zwp_tablet_tool_v2 *id,
                          uint32_t pressure) {}

void tablet_tool_distance(void *data, struct zwp_tablet_tool_v2 *id,
                          uint32_t distance) {}

void tablet_tool_tilt(void *data, struct zwp_tablet_tool_v2 *id, wl_fixed_t x,
                      wl_fixed_t y) {}

void tablet_tool_rotation(void *data, struct zwp_tablet_tool_v2 *id,
                          wl_fixed_t rotation) {}

void tablet_tool_slider(void *data, struct zwp_tablet_tool_v2 *id, int slider) {
}

void tablet_tool_wheel(void *data, struct zwp_tablet_tool_v2 *id,
                       wl_fixed_t degree, int clicks) {}

void tablet_tool_button(void *data, struct zwp_tablet_tool_v2 *id,
                        uint32_t serial, uint32_t button, uint32_t state) {}

void tablet_tool_frame(void *data, struct zwp_tablet_tool_v2 *id,
                       uint32_t time) {}

void tablet_tool_done(void *data, struct zwp_tablet_tool_v2 *id) { /* empty */ }

static const struct zwp_tablet_tool_v2_listener tablet_tool_listener = {
    .removed = tablet_tool_removed,
    .down = tablet_tool_down,
    .up = tablet_tool_up,
    .motion = tablet_tool_motion,

    .type = tablet_tool_type,
    .hardware_serial = tablet_tool_serial,
    .hardware_id_wacom = tablet_tool_id_wacom,
    .capability = tablet_tool_capability,
    .done = tablet_tool_done,

    .proximity_in = tablet_tool_proximity_in,
    .proximity_out = tablet_tool_proximity_out,
    .pressure = tablet_tool_pressure,
    .distance = tablet_tool_distance,
    .tilt = tablet_tool_tilt,
    .rotation = tablet_tool_rotation,
    .slider = tablet_tool_slider,
    .wheel = tablet_tool_wheel,
    .button = tablet_tool_button,
    .frame = tablet_tool_frame,
};

void tablet_added(void *data, struct zwp_tablet_seat_v2 *zwp_tablet_seat_v2,
                  struct zwp_tablet_v2 *zwp_tablet_v2) {}

void tool_added(void *data, struct zwp_tablet_seat_v2 *zwp_tablet_seat_v2,
                struct zwp_tablet_tool_v2 *zwp_tablet_tool_v2) {
  tablet_tool = zwp_tablet_tool_v2;
  zwp_tablet_tool_v2_add_listener(tablet_tool, &tablet_tool_listener, NULL);
}

static const struct zwp_tablet_seat_v2_listener tablet_seat_listener = {
    .tool_added = tool_added,
    .tablet_added = tablet_added,
};

void touch_down(void *data, struct wl_touch *wl_touch, uint serial, uint time,
                struct wl_surface *surface, int id, wl_fixed_t x,
                wl_fixed_t y) {
  if (touch_id == 255555) {
    mouse_x = wl_fixed_to_double(x);
    mouse_y = wl_fixed_to_double(y);
    touch_id = id;
    button_state = 1;
  }
}

void touch_up(void *data, struct wl_touch *wl_touch, uint serial, uint time,
              int id) {
  if (touch_id == id) {
    touch_id = 255555;
    button_state = 0;
  }
}

void touch_motion(void *data, struct wl_touch *wl_touch, uint time, int id,
                  wl_fixed_t x, wl_fixed_t y) {
  mouse_x = wl_fixed_to_double(x);
  mouse_y = wl_fixed_to_double(y);
}

void touch_frame(void *data, struct wl_touch *wl_touch) {}

void touch_cancel(void *data, struct wl_touch *wl_touch) {}

void touch_shape(void *data, struct wl_touch *wl_touch, int32_t id,
                 wl_fixed_t major, wl_fixed_t minor) {}

void touch_orientation(void *data, struct wl_touch *wl_touch, int32_t id,
                       wl_fixed_t orientation) {}

static const struct wl_touch_listener touch_listener = {
    .down = touch_down,
    .up = touch_up,
    .motion = touch_motion,
    .frame = touch_frame,
    .cancel = touch_cancel,
    .shape = touch_shape,
    .orientation = touch_orientation,
};

static uint32_t last_key = 0;
static uint32_t last_key_state = 0;

void keyboard_keymap(void *data, struct wl_keyboard *keyboard, uint32_t format,
                     int32_t fd, uint32_t size) {
  if (format != WL_KEYBOARD_KEYMAP_FORMAT_XKB_V1) {
    close(fd);
    return;
  }

  char *map_shm = mmap(NULL, size, PROT_READ, MAP_PRIVATE, fd, 0);
  if (map_shm == MAP_FAILED) {
    close(fd);
    return;
  }

  xkb_keymap = xkb_keymap_new_from_string(xkb_context, map_shm,
                                          XKB_KEYMAP_FORMAT_TEXT_V1,
                                          XKB_KEYMAP_COMPILE_NO_FLAGS);
  munmap(map_shm, size);
  close(fd);

  if (!xkb_keymap) {
    return;
  }

  xkb_state = xkb_state_new(xkb_keymap);
  if (!xkb_state) {
    return;
  }
}

void keyboard_enter(void *data, struct wl_keyboard *keyboard, uint32_t serial,
                    struct wl_surface *surface, struct wl_array *keys) {}

void keyboard_leave(void *data, struct wl_keyboard *keyboard, uint32_t serial,
                    struct wl_surface *surface) {}

void keyboard_key(void *data, struct wl_keyboard *keyboard, uint32_t serial,
                  uint32_t time, uint32_t key, uint32_t state) {
  if (xkb_state) {
    xkb_keysym_t sym = xkb_state_key_get_one_sym(xkb_state, key + 8);
    if (state == WL_KEYBOARD_KEY_STATE_PRESSED) {
      last_key = sym;
      last_key_state = 1;
    } else {
      last_key = 0;
      last_key_state = 0;
    }
  }
}

void keyboard_modifiers(void *data, struct wl_keyboard *keyboard,
                        uint32_t serial, uint32_t mods_depressed,
                        uint32_t mods_latched, uint32_t mods_locked,
                        uint32_t group) {
  if (xkb_state) {
    xkb_state_update_mask(xkb_state, mods_depressed, mods_latched, mods_locked,
                          0, 0, group);
  }
}

void keyboard_repeat_info(void *data, struct wl_keyboard *keyboard,
                          int32_t rate, int32_t delay) {}

static const struct wl_keyboard_listener keyboard_listener = {
    .keymap = keyboard_keymap,
    .enter = keyboard_enter,
    .leave = keyboard_leave,
    .key = keyboard_key,
    .modifiers = keyboard_modifiers,
    .repeat_info = keyboard_repeat_info,
};

// Seat listener
void seat_capabilities(void *data, struct wl_seat *seat,
                       uint32_t capabilities) {
  if (capabilities & WL_SEAT_CAPABILITY_POINTER) {
    pointer = wl_seat_get_pointer(seat);
    wl_pointer_add_listener(pointer, &pointer_listener, NULL);
  }

  if (capabilities & WL_SEAT_CAPABILITY_KEYBOARD) {
    keyboard = wl_seat_get_keyboard(seat);
    wl_keyboard_add_listener(keyboard, &keyboard_listener, NULL);

    if (shortcuts_inhibit_manager && surface_global && !shortcuts_inhibitor) {
      shortcuts_inhibitor =
          zwp_keyboard_shortcuts_inhibit_manager_v1_inhibit_shortcuts(
              shortcuts_inhibit_manager, surface_global, seat);
    }
  }
  if (capabilities & WL_SEAT_CAPABILITY_TOUCH) {
    touch = wl_seat_get_touch(seat);
    wl_touch_add_listener(touch, &touch_listener, NULL);
  }

  tablet_seat = zwp_tablet_manager_v2_get_tablet_seat(tablet_manager, seat);
  zwp_tablet_seat_v2_add_listener(tablet_seat, &tablet_seat_listener, seat);
}

void seat_name(void *data, struct wl_seat *seat, const char *name) {}

int get_button_state() { return button_state; }

void get_mouse_pos(double *x, double *y) {
  *x = mouse_x;
  *y = mouse_y;
}

void get_dimensions(int32_t *w, int32_t *h) {
  *w = width_global;
  *h = height_global;
}

uint32_t get_last_key() { return last_key; }

uint32_t get_last_key_state() { return last_key_state; }

void clear_last_key() {
  last_key = 0;
  last_key_state = 0;
}

EGLNativeWindowType native_window(struct wl_egl_window *egl_window) {
  return (EGLNativeWindowType)egl_window;
}
