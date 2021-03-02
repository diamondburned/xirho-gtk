{ pkgs ? import <nixpkgs> {} }:

pkgs.stdenv.mkDerivation rec {
	name = "xirho-gtk";
	version = "0.0.1";

	buildInputs = with pkgs; [
		gnome3.glib gnome3.gtk libhandy
	];

	nativeBuildInputs = with pkgs; [
		pkgconfig go
	];
}
