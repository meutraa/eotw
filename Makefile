all:
	nix-shell -p wayland mesa_glu libxkbcommon alsaLib --run 'go build -tags wayland'
