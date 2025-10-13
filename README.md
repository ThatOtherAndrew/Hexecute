# Hexecute

Launch apps by casting spells! 🪄

![Demo GIF](assets/demo.gif)

## Installation

TODO

## Usage

### Setting a Keybind
The recommended way to use Hexecute is to bind it to a keyboard shortcut in your compositor.

Listed below are some examples for popular compositors using the `SUPER` + `SPACE` keybind.

#### Hyprland

If you're using Hyprland, add the following line to your `~/.config/hypr/hyprland.conf`:

```
bind = SUPER, SPACE, exec, hexecute
```

#### Sway

If you're using Sway, add the following line to your `~/.config/sway/config`:

```
bindsym $mod+space exec hexecute
```

### Learning a Gesture

To configure a gesture to launch an application, run `hexecute --learn [command]` in a terminal. Hexecute should launch - simply draw your chosen gesture **3 times** and it will be mapped to the command.

![Gesture learning demo](assets/hexecute-learn.gif)

### Managing Gestures

To view all your configured gestures, run `hexecute --list` in a terminal.

To delete a previously assigned gesture, use the `hexecute --delete [gesture]` command.

All gestures are saved in the `~/.config/hexecute/gestures.json` file. This file can be manually shared, edited, backed up, or swapped.
