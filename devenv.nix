{ pkgs, lib, config, inputs, ... }:

{
  packages = [
    pkgs.go-tools
    pkgs.golangci-lint
  ];

  languages.go = {
    enable = true;
  };

  languages.javascript = {
    directory = "./ui/";
    enable = true;
    npm = {
      enable = true;
      install.enable = true;
    };
  };
}
