package handlers

import (
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/pkg/utils"
)

type happenerPage struct {
	name string
	id   string
}

var happenerPages = []happenerPage{
	{name: "radiation_levels", id: "867_0002"},
	{name: "price_of_electricity", id: "189_0001"},
	{name: "exchange", id: "180_0001"},
	{name: "foreign_currencies", id: "173_0002"},
	{name: "marine_weather", id: "403_0001"},
	{name: "emergency_warnings", id: "112_0001"},
}

func pickHappenerPage() happenerPage {
	return happenerPages[rand.Intn(len(happenerPages))]
}

type HappenerHandler struct {
	next ContextHandler
}

func (h *HappenerHandler) Execute(m *Context) {
	slog.Debug("Entering HappenerHandler")

	if m.action == Happener {
		selected := pickHappenerPage()
		page, err := utils.FetchTeletextPage(selected.id)
		if err != nil {
			slog.Error("Failed to fetch teletext page", "error", err, "page", selected.id, "name", selected.name)
		} else {
			slog.Info("Sending teletext page", "page", selected.id, "name", selected.name)
			path := filepath.Join(config.FromEnv().TELETEXT_IMAGE_DIR, uuid.New().String()+page.FileExtension())
			if err := os.WriteFile(path, page.ImageData, 0644); err != nil {
				slog.Error("Failed to write teletext image", "error", err, "path", path)
			} else {
				m.finalImagePath = path
			}
		}
	}

	h.next.Execute(m)
}

func (h *HappenerHandler) SetNext(next ContextHandler) {
	h.next = next
}
