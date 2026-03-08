package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

//go:embed icon.svg
var iconDarkBytes []byte

//go:embed icon.light.svg
var iconLightBytes []byte

//go:embed icon.tray.svg
var trayIconBytes []byte

// themedIcon picks the correct static resource based on the
// current theme variant. Implements fyne.Resource.
type themedIcon struct{}

func (t *themedIcon) Name() string { return "hi-icon.svg" }

func (t *themedIcon) Content() []byte {
	if fyne.CurrentApp().Settings().ThemeVariant() == theme.VariantLight {
		return iconLightBytes
	}
	return iconDarkBytes
}

// Icon returns the correctly themed app icon.
var Icon fyne.Resource = &themedIcon{}

var Tray fyne.Resource = fyne.NewStaticResource("hi-tray.svg", trayIconBytes)
