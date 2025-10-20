# Hexecute

A gesture-based launcher for Wayland. Launch apps by casting spells! 🪄

![Demo GIF](.github/assets/demo.gif)

## Installation

### Nix / NixOS

If you're a lucky [Nix](https://nixos.org/) user, you can:

**Try it out without installing:**
```bash
nix run github:ThatOtherAndrew/Hexecute
```

**Install to your profile:**
```bash
nix profile install github:ThatOtherAndrew/Hexecute
```

**Add to your NixOS configuration:**
```nix
# flake.nix
{
  inputs.hexecute.url = "github:ThatOtherAndrew/Hexecute";
}
```
```nix
# configuration.nix
{
  environment.systemPackages = with pkgs; [
    inputs.hexecute.packages.${pkgs.system}.default
  ];
}
```

### Executable download

Download the latest version from the [release page](https://github.com/ThatOtherAndrew/Hexecute/releases/latest), and place it somewhere in your `$PATH`.

**Don't forget to rename the downloaded binary to `hexecute` and make it executable:**
```bash
mv hexecute-1.2.3-blah hexecute
chmod +x hexecute
```

### Build from Source

Clone the repository:
```bash
git clone https://github.com/ThatOtherAndrew/Hexecute
cd Hexecute
```

If you have [Nix](https://nixos.org/) installed, simply run `nix build`.

Otherwise, make sure you have Go (and all dependent Wayland (and X11!?) libs) installed, then run:
```bash
mkdir -p bin
go build -o bin ./...
./bin/hexecute
```

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
