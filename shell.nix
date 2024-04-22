{ pkgs ? import (
  fetchTarball {
    url = "https://github.com/NixOS/nixpkgs/archive/f2d7a289c5a5ece8521dd082b81ac7e4a57c2c5c.tar.gz";
    sha256 = "sha256:1r62iaxd4hzimcpisssy3j0394iw9icc9kggdajl8rv9w2w0mpv1";
  }
) {} }:

pkgs.mkShellNoCC {
  buildInputs = [
    pkgs.go_1_22
    pkgs.gitMinimal
    pkgs.goreleaser
    pkgs.syft
    pkgs.cosign
    pkgs.golangci-lint
    pkgs.terraform
    pkgs.terraform-plugin-docs
  ];
}
