# ml2 ts=2 sts=2 sw=2
{buildGoModule, version ? "HEAD"}: buildGoModule {
  pname = "tt";
  inherit version;

  # In 'nix develop', we don't need a copy of the source tree
  # in the Nix store.
  src = ./..;

  # This hash locks the dependencies of this package. It is
  # necessary because of how Go requires network access to resolve
  # VCS.  See https://www.tweag.io/blog/2021-03-04-gomod2nix/ for
  # details. Normally one can build with a fake hash and rely on native Go
  # mechanisms to tell you what the hash should be or determine what
  # it should be "out-of-band" with other tooling (eg. gomod2nix).
  # To begin with it is recommended to set this, but one must
  # remember to bump this hash when your dependencies change.
  # vendorHash = pkgs.lib.fakeHash;

  vendorHash = "sha256-cpjMZr+kP+XyXRxK4YI/lGHl97IROFVVz8DrID0vi6E=";
}
