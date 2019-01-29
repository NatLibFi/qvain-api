package selfcheck

import (
	"fmt"
	"io"
)

type HtmlWriter struct {
	Out io.Writer
}

func (w *HtmlWriter) StartGroup(s string) {
	fmt.Fprintln(w.Out, `<p class="check-group">`)
}
func (w *HtmlWriter) Tag(s string) {
	fmt.Fprint(w.Out, `<span class="check-tag">`+s+` - </span>`)
}

func (w *HtmlWriter) Msg(s string) {
	fmt.Fprint(w.Out, `<span class="check-msg">`+s+`... </span>`)
}
func (w *HtmlWriter) Good(s string) {
	fmt.Fprint(w.Out, `<span class="check-good">✓</span>`)
	w.end()
}
func (w *HtmlWriter) Bad(s string) {
	fmt.Fprint(w.Out, `<span class="check-bad">✗</span> <span class="check-error">`+s+`</span>`)
	w.end()
}

func (w *HtmlWriter) Skipping() {
	fmt.Fprintln(w.Out, `<span class="check-skipping">skipping futher tests</span>`)
}

func (w *HtmlWriter) end() {
	fmt.Fprintln(w.Out, `<br/>`)
}

func (w *HtmlWriter) EndGroup() {
	fmt.Fprintln(w.Out, `</p>`)
}
