{
  description = "Hexecute";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs =
    inputs:
    let
      system = "x86_64-linux";
      pkgs = inputs.nixpkgs.legacyPackages.${system};
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        name = "hexecute";

        packages = with pkgs; [
          at-spi2-atk
          atkmm
          bun
          cairo
          cargo
          gdk-pixbuf
          glib
          gobject-introspection
          gtk3
          harfbuzz
          librsvg
          libsoup_3
          openssl
          pango
          pkg-config
          rustc
          webkitgtk_4_1
          xdg-utils
        ];
      };
    };
}
