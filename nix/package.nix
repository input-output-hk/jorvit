{ buildGoPackage, nix-gitignore }:
buildGoPackage rec {
  name = "jorvit-${version}";
  version = "0.0.0";
  goPackagePath = "github.com/input-output-hk/jorvit";
  src = nix-gitignore.gitignoreSource [] ../.;
  goDeps = ./deps.nix;
}
