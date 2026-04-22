flake:

{ config, lib, pkgs, ... }:

let
  cfg = config.programs.softcred;
in
{
  options.programs.softcred = {
    enable = lib.mkEnableOption "softcred - US Bank Triple Cash software credit MCC tracker";

    package = lib.mkOption {
      type = lib.types.package;
      default = flake.packages.${pkgs.system}.default;
      description = "The softcred package to use.";
    };

    dataDir = lib.mkOption {
      type = lib.types.str;
      default = "${config.xdg.dataHome}/softcred";
      description = "Directory to store the softcred database.";
    };
  };

  config = lib.mkIf cfg.enable {
    home.packages = [
      (pkgs.writeShellScriptBin "softcred" ''
        exec ${lib.getExe cfg.package} --db "${cfg.dataDir}/data.db" "$@"
      '')
    ];
  };
}
