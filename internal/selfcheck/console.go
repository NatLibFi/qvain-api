package selfcheck

import (
	"fmt"
	"io"
)

const (
	cReset    = 0
	cBold     = 1
	cRed      = 31
	cGreen    = 32
	cYellow   = 33
	cBlue     = 34
	cMagenta  = 35
	cCyan     = 36
	cGray     = 37
	cDarkGray = 90
)

type ConsoleWriter struct {
	Out io.Writer
}

func (w *ConsoleWriter) StartGroup(s string) {
}
func (w *ConsoleWriter) Tag(s string) {
	fmt.Fprint(w.Out, inColour(cDarkGray, s+" - "))
}

func (w *ConsoleWriter) Msg(s string) {
	fmt.Fprint(w.Out, inColour(cDarkGray, s+"... "))
}

func (w *ConsoleWriter) Good(s string) {
	fmt.Fprint(w.Out, inColour(cGreen, "✓"))
	w.end()
}

func (w *ConsoleWriter) Bad(s string) {
	fmt.Fprint(w.Out, inColour(cRed, "✗")+inColour(cBold, " "+s))
	w.end()
}

func (w *ConsoleWriter) Skipping() {
	fmt.Fprintln(w.Out, "... skipping futher tests")
}

func (w *ConsoleWriter) end() {
	//fmt.Fprintln(w.Out, inColour(cReset, "\n"))
	fmt.Fprintln(w.Out)
}

func (w *ConsoleWriter) EndGroup() {
	fmt.Fprintln(w.Out)
}

func inColour(c int, s interface{}) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}
