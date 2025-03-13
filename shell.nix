{ pkgs ? import (
  fetchTarball {
    url = "https://github.com/NixOS/nixpkgs/archive/b62d2a95c72fb068aecd374a7262b37ed92df82b.tar.gz";
    sha256 = "sha256:0rxdf3jhh3ac3z361gz6cjqrd3vvnd9wbsljxi6rmcgq01snmm3h";
  }
) {} }:

pkgs.mkShellNoCC {
  buildInputs = [
    pkgs.go_1_24
    pkgs.gitMinimal
    pkgs.goreleaser
    pkgs.syft
    pkgs.cosign
    pkgs.golangci-lint
    pkgs.terraform
    pkgs.terraform-plugin-docs
  ];
}
