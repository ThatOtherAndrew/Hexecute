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

          # Build libs
          xorg.libX11
          xorg.libXrandr
          xorg.libXxf86vm
          xorg.libXi
          xorg.libXcursor
          xorg.libXinerama
        ];
      };
    };
}
