{ nixpkgs ? import <nixpkgs> {} }:
with nixpkgs; mkShellNoCC {
  nativeBuildInputs = [
    gopls
    go
  ];
}

