// Package renderは、MarkdownをANSIカラーコードを使用してレンダリングするためのユーティリティ関数を提供します。
package render

import (
	"sync"

	"github.com/charmbracelet/glamour"
)

var (
	renderer *glamour.TermRenderer
	once     sync.Once
)

// getRendererは、シングルトンパターンを使用してTermRendererを返します。
// 初回呼び出し時にのみ新しいTermRendererを作成し、以降の呼び出しでは同じインスタンスを返します。
func getRenderer() *glamour.TermRenderer {
	once.Do(func() {
		var err error
		renderer, err = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(100),
		)
		if err != nil {
			panic(err)
		}
	})
	return renderer
}

// RenderMarkdownは、指定されたMarkdown文字列をANSIカラーコードを使用してレンダリングします。
// レンダリングに成功した場合は装飾されたテキストを、エラーが発生した場合は元のMarkdown文字列をそのまま返します。
func RenderMarkdown(md string) string {
	r := getRenderer()
	out, err := r.Render(md)
	if err != nil {
		return md
	}
	return out
}
