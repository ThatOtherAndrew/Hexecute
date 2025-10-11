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
          go

          # Build deps
          libGL
          libxxf86vm
          pkg-config
          SDL2
          sdl3

          # X11
          xorg.libXi
          xorg.libXcursor
          xorg.libXrandr
          xorg.libXinerama

          # Wayland
          wayland
          libxkbcommon
        ];
      };
    };
}
