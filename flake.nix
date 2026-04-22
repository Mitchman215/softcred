{
  description = "softcred - US Bank Triple Cash software credit MCC tracker";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    let
      overlay = final: prev: {
        softcred = final.callPackage ./nix/package.nix { };
      };

      hmModule = import ./nix/hm-module.nix self;
    in
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ overlay ];
        };
      in
      {
        packages = {
          default = pkgs.softcred;
          softcred = pkgs.softcred;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
          ];
        };
      }
    ) // {
      overlays.default = overlay;
      homeManagerModules.default = hmModule;
    };
}
