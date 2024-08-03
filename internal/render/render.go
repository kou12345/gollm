package render

import (
	"github.com/charmbracelet/glamour"
)

func RenderMarkdown(md string) string {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	out, err := r.Render(md)
	if err != nil {
		return md // Return the original text if an error occurs
	}
	return out
}
