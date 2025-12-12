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
        qvpnPackage = pkgs.buildGoModule {
          pname = "qvpn";
          version = self.shortRev or "dirty";

          src = ./.;

          vendorHash = pkgs.lib.fakeHash; 

          # Build flags (optional)
          ldflags = [
            "-s"
            "-w"
            "-X main.version=${self.shortRev or "dirty"}"
          ];

          subPackages = [ "." ];

          meta = with pkgs.lib; {
            description = "A VPN CLI tool written in Go";
            homepage = "https://github.com/PraveenPrabhuT/qvpn";
            license = licenses.mit;
            maintainers = [ ];
            platforms = platforms.unix;
          };
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
          program = "${self.packages.default}/bin/qvpn";
        };
      }
    );
}