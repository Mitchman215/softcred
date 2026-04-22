{ lib, buildGoModule }:

buildGoModule {
  pname = "softcred";
  version = "0.1.0";

  src = lib.cleanSource ./..;

  vendorHash = "sha256-Iebe5Kd67jR7W/VLn86PzKCQFj+0UC59u7LILktWYj8=";

  subPackages = [ "cmd/softcred" ];

  meta = {
    description = "US Bank Triple Cash software credit MCC tracker";
    mainProgram = "softcred";
  };
}
