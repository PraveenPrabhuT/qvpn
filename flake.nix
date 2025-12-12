{
  description = "QVPN - A VPN CLI tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
      packages = rec {
        qvpn = pkgs.buildGoModule {
          pname = "qvpn";
          version = self.shortRev or "dirty";

          src = ./.;

          vendorHash = pkgs.lib.fakeHash; 

          ldflags = [
            "-s"
            "-w"
            "-X main.version=${self.shortRev or "dirty"}"
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
          ];
        };

        apps.default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/qvpn";
        };
      }
    );
}