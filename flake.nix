{
  inputs = {
    nix-go = {
      inputs.nixpkgs.follows = "nixpkgs";
      url = "github:matthewdargan/nix-go";
    };
    nixpkgs.url = "nixpkgs/nixos-unstable";
    parts.url = "github:hercules-ci/flake-parts";
    pre-commit-hooks = {
      inputs.nixpkgs.follows = "nixpkgs";
      url = "github:cachix/pre-commit-hooks.nix";
    };
  };
  outputs = inputs:
    inputs.parts.lib.mkFlake {inherit inputs;} {
      imports = [inputs.pre-commit-hooks.flakeModule];
      perSystem = {
        config,
        inputs',
        lib,
        pkgs,
        ...
      }: {
        devShells.default = pkgs.mkShell {
          packages = [inputs'.nix-go.packages.go];
          shellHook = "${config.pre-commit.installationScript}";
        };
        packages = {
          swippy = inputs'.nix-go.legacyPackages.buildGoModule {
            meta = with lib; {
              description = "Retrieve from the eBay Finding API and store results in a PostgreSQL database";
              homepage = "https://github.com/matthewdargan/swippy";
              license = licenses.asl20;
              maintainers = with maintainers; [matthewdargan];
            };
            pname = "swippy";
            src = ./.;
            vendorHash = "sha256-G4D/eZjveGEhCVZUtjuFm53rVRWHI5OWvg3iexmKXB4=";
            version = "0.1.0";
          };
        };
        pre-commit = {
          check.enable = false;
          settings = {
            hooks = {
              golangci-lint = {
                enable = true;
                package = inputs'.nix-go.packages.golangci-lint;
              };
              gotest = {
                enable = true;
                package = inputs'.nix-go.packages.go;
              };
            };
            src = ./.;
          };
        };
      };
      systems = ["aarch64-darwin" "aarch64-linux" "x86_64-darwin" "x86_64-linux"];
    };
}
