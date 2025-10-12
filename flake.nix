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
          pkg-config

          # Wayland libraries
          wayland
          wayland-protocols
          wayland-scanner
          libxkbcommon

          # EGL and OpenGL
          libGL
          libGLU
          mesa

          # Build tools
          gcc
        ];
      };
    };
}
