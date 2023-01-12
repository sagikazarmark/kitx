{
  description = "Go kit extensions";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          buildDeps = with pkgs; [ git go gnumake ];
          devDeps = with pkgs; buildDeps ++ [ golangci-lint gotestsum ];

          ciShell = go:
            pkgs.mkShell {
              buildInputs = with pkgs; [
                git
                gnumake
                gotestsum
              ] ++ [ go ];
            };

          goVerToPkg = goVersion: builtins.replaceStrings [ "." ] [ "_" ] goVersion;

          genCiShells = goVersions:
            builtins.listToAttrs (map (goVersion: pkgs.lib.attrsets.nameValuePair "ci${goVerToPkg goVersion}" (ciShell pkgs."go_${goVerToPkg goVersion}")) goVersions);
        in
        {
          devShells = {
            default = pkgs.mkShell {
              buildInputs = with pkgs; [
                git
                go_1_19
                gnumake
                golangci-lint
                gotestsum
              ];
            };
          } // genCiShells [ "1.18" "1.19" ];
        });
}
