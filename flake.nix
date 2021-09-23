{
  description = "Go kit extensions";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-21.05";
    flake-utils.url = "github:numtide/flake-utils";
    flake-compat = {
      url = "github:edolstra/flake-compat";
      flake = false;
    };
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        buildDeps = with pkgs; [ git go gnumake ];
        devDeps = with pkgs; buildDeps ++ [ golangci-lint gotestsum ];

        generateGoEnv = go:
          pkgs.buildEnv {
            name = "go" + go.version;
            paths = (pkgs.lib.remove pkgs.go devDeps) ++ [ go ];
          };
      in {
        devShell = pkgs.mkShell {
          buildInputs = devDeps;

          shellHook = ''
            echo -e "Welcome to the developer console!\n"
            echo "Available make commands:"
            make
          '';
        };

        packages.go1_15 = generateGoEnv pkgs.go_1_15;
        packages.go1_16 = generateGoEnv pkgs.go_1_16;
      });
}
