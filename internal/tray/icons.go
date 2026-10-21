package tray

import _ "embed"

//go:embed icons/tray.png
var iconNormal []byte

//go:embed icons/tray-offline.png
var iconOffline []byte

// PNG returns the image of an Icon.
func (i Icon) PNG() []byte {
	if i == IconOffline {
		return iconOffline
	}
	return iconNormal
}
