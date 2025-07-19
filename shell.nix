{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    cobra-cli
    gopls
    gotools
    go-tools
  ];

  shellHook = ''
    echo "Go development environment loaded"
    echo "Go version: $(go version)"
    echo "Cobra CLI Loaded"
  '';
}