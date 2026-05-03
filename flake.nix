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
      hexecute = pkgs.buildGoModule {
        pname = "hexecute";
        version = "0.1.0";

        src = ./.;

        vendorHash = "sha256-CIlYhcX7F08Xwrr3/0tkgrfuP68UU0CeQ+HV63b6Ddg=";

        nativeBuildInputs = with pkgs; [
          pkg-config
          makeWrapper
        ];

        buildInputs = with pkgs; [
          wayland
          wayland-protocols
          libxkbcommon
          libGL
          libGLU
          mesa
          xorg.libX11
        ];

        postFixup = ''
          wrapProgram $out/bin/hexecute \
            --prefix __EGL_VENDOR_LIBRARY_DIRS : "/run/opengl-driver/share/glvnd/egl_vendor.d" \
            --prefix __EGL_VENDOR_LIBRARY_DIRS : "${pkgs.mesa}/share/glvnd/egl_vendor.d" \
            --prefix LIBGL_DRIVERS_PATH : "/run/opengl-driver/lib/dri" \
            --prefix LIBGL_DRIVERS_PATH : "${pkgs.mesa}/lib/dri"
        '';

        meta = {
          description = "Launch apps by casting spells! 🪄";
          homepage = "https://github.com/ThatOtherAndrew/Hexecute";
          license = pkgs.lib.licenses.gpl3;
          platforms = pkgs.lib.platforms.linux;
        };
      };
    in
    {
      packages.${system} = {
        inherit hexecute;
        default = hexecute;
      };

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
