{ pkgs ? import (
  fetchTarball {
    url = "https://github.com/NixOS/nixpkgs/archive/09ec6a0881e1a36c29d67497693a67a16f4da573.tar.gz";
    sha256 = "sha256:0cr9czryigfilys530dpp0q3hjny808i4dg7xckzjn8m55gjn6gc";
  }
) {} }:

pkgs.mkShellNoCC {
  buildInputs = [
    pkgs.go_1_21
    pkgs.gitMinimal
    pkgs.goreleaser
    pkgs.syft
    pkgs.cosign
    pkgs.golangci-lint
    pkgs.terraform
    pkgs.terraform-plugin-docs
  ];
}
