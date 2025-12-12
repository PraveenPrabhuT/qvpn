{
  description = "QVPN - A VPN CLI tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix.url = "github:nix-community/gomod2nix";
    gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
    gomod2nix.inputs.flake-utils.follows = "flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, gomod2nix }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ gomod2nix.overlays.default ];
        };
      in
      {
      packages = rec {
        qvpn = pkgs.buildGoApplication {
          pname = "qvpn";
          version = self.shortRev or "dirty";

          src = ./.;

          modules = ./gomod2nix.toml;
          

          ldflags = [
            "-s"
            "-w"
            "-X github.com/PraveenPrabhuT/qvpn/cmd.Version=${self.shortRev or "dirty"}"
          ];


          meta = with pkgs.lib; {
            description = "A VPN CLI tool written in Go";
            homepage = "https://github.com/PraveenPrabhuT/qvpn";
            license = licenses.mit;
            maintainers = [ ];
            platforms = platforms.unix;
          };
          };

          # Set the default package (allows `nix build` without arguments)
          default = qvpn;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gotools
            gopls
            golangci-lint
            gomod2nix.packages.${system}.default
          ];
        };

        apps.default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/qvpn";
        };
      }
    );
}