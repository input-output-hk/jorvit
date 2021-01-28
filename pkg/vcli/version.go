package vcli

// Version - get vit-servicing-station-cli version.
//
//  vit-servicing-station-cli --version | STDOUT
func Version() ([]byte, error) {
	return vcli(nil, "--version")
}
