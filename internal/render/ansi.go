package render

func ansi(code, s string) string {
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func Bold(s string) string       { return ansi("1", s) }
func Italic(s string) string     { return ansi("3", s) }
func Red(s string) string        { return ansi("31", s) }
func Yellow(s string) string     { return ansi("33", s) }
func Cyan(s string) string       { return ansi("36", s) }
func Blue(s string) string       { return ansi("34", s) }
func Gray(s string) string       { return ansi("90", s) }
func White(s string) string      { return ansi("97", s) }
func BoldCyan(s string) string   { return ansi("1;36", s) }
func BoldRed(s string) string    { return ansi("1;31", s) }
func BoldBlue(s string) string   { return ansi("1;34", s) }
func BoldYellow(s string) string { return ansi("1;33", s) }
