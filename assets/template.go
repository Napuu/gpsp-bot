package assets

import (
	"embed"
)

//go:embed pate_template.jpg day_template.mp4 Roboto-Bold.ttf
var TemplateFS embed.FS
