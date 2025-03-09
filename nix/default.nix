# ml2 ts=2 sts=2 sw=2

with import <nixpkgs> {};
{version ? "HEAD"}: callPackage ./package.nix {inherit version;}
