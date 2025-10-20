#ifndef WAYLAND_H
#define WAYLAND_H

#include "wlr-layer-shell-client.h"
#include <EGL/egl.h>
#include <stdlib.h>
#include <wayland-client.h>
#include <wayland-egl.h>
#include <xkbcommon/xkbcommon.h>

void layer_surface_configure(void *data, struct zwlr_layer_surface_v1 *surface,
                             uint32_t serial, uint32_t width, uint32_t height);
void layer_surface_closed(void *data, struct zwlr_layer_surface_v1 *surface);
void seat_capabilities(void *data, struct wl_seat *seat, uint32_t capabilities);
void seat_name(void *data, struct wl_seat *seat, const char *name);
void registry_global(void *data, struct wl_registry *registry, uint32_t name,
                     const char *interface, uint32_t version);
void registry_global_remove(void *data, struct wl_registry *registry,
                            uint32_t name);
struct wl_registry *get_registry(struct wl_display *display);
void add_registry_listener(struct wl_registry *registry);
struct zwlr_layer_surface_v1 *create_layer_surface(struct wl_surface *surface);
void set_input_region(int32_t width, int32_t height);
void disable_all_input();
void pointer_enter(void *data, struct wl_pointer *pointer, uint32_t serial,
                   struct wl_surface *surface, wl_fixed_t x, wl_fixed_t y);
void pointer_leave(void *data, struct wl_pointer *pointer, uint32_t serial,
                   struct wl_surface *surface);
void pointer_motion(void *data, struct wl_pointer *pointer, uint32_t time,
                    wl_fixed_t x, wl_fixed_t y);
void pointer_button(void *data, struct wl_pointer *pointer, uint32_t serial,
                    uint32_t time, uint32_t button, uint32_t state);
void pointer_axis(void *data, struct wl_pointer *pointer, uint32_t time,
                  uint32_t axis, wl_fixed_t value);
void pointer_frame(void *data, struct wl_pointer *pointer);
void pointer_axis_source(void *data, struct wl_pointer *pointer,
                         uint32_t source);
void pointer_axis_stop(void *data, struct wl_pointer *pointer, uint32_t time,
                       uint32_t axis);
void pointer_axis_discrete(void *data, struct wl_pointer *pointer,
                           uint32_t axis, int32_t discrete);
void keyboard_keymap(void *data, struct wl_keyboard *keyboard, uint32_t format,
                     int32_t fd, uint32_t size);
void keyboard_enter(void *data, struct wl_keyboard *keyboard, uint32_t serial,
                    struct wl_surface *surface, struct wl_array *keys);
void keyboard_leave(void *data, struct wl_keyboard *keyboard, uint32_t serial,
                    struct wl_surface *surface);
void keyboard_key(void *data, struct wl_keyboard *keyboard, uint32_t serial,
                  uint32_t time, uint32_t key, uint32_t state);
void keyboard_modifiers(void *data, struct wl_keyboard *keyboard,
                        uint32_t serial, uint32_t mods_depressed,
                        uint32_t mods_latched, uint32_t mods_locked,
                        uint32_t group);
void keyboard_repeat_info(void *data, struct wl_keyboard *keyboard,
                          int32_t rate, int32_t delay);
int get_button_state();
void get_mouse_pos(double *x, double *y);
void get_dimensions(int32_t *w, int32_t *h);
uint32_t get_last_key();
uint32_t get_last_key_state();
void clear_last_key();
EGLNativeWindowType native_window(struct wl_egl_window *egl_window);

extern struct wl_compositor *compositor;
extern struct zwlr_layer_shell_v1 *layer_shell;
extern struct wl_seat *seat;
extern struct wl_pointer *pointer;
extern struct wl_keyboard *keyboard;
extern struct zwp_keyboard_shortcuts_inhibit_manager_v1
    *shortcuts_inhibit_manager;
extern struct zwp_keyboard_shortcuts_inhibitor_v1 *shortcuts_inhibitor;
extern struct zwlr_layer_surface_v1 *layer_surface_global;
extern struct xkb_context *xkb_context;
extern struct xkb_keymap *xkb_keymap;
extern struct xkb_state *xkb_state;

#endif // WAYLAND_H
