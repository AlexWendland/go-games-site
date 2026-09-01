{ pkgs, lib, config, inputs, ... }:

{
  packages = [
    pkgs.go-tools
    pkgs.golangci-lint
    pkgs.fd # For import sorting
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

  git-hooks.hooks = {
    # Go hooks
    govet.enable = true;
    golangci-lint.enable = true;
    gotest.enable = true;
    goimports = {
      enable = true;
      name = "goimports";
      entry = "bash -c 'fd -e go --exec goimports -l | grep -q . && { echo \"Run make fmt to fix import ordering\"; exit 1; } || exit 0'";
      pass_filenames = false;
      always_run = true;
    };

    # Frontend hooks
  };
}
