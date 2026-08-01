package render

import "strings"

// margin describes blank lines/spaces to add around a box.
type margin struct {
	Top, Bottom, Left, Right int
}

// box renders content inside a rounded ASCII border with an optional
// centered title, akin to the "boxen" package used by the original CLI.
func box(content, title string, color func(string) string, padding int, m margin) string {
	lines := strings.Split(content, "\n")

	maxWidth := visibleLen(title)
	for _, l := range lines {
		if w := visibleLen(l); w > maxWidth {
			maxWidth = w
		}
	}

	innerWidth := maxWidth + padding*2
	leftIndent := strings.Repeat(" ", m.Left)

	dashSpace := innerWidth - visibleLen(title)
	if dashSpace < 0 {
		dashSpace = 0
	}
	leftDash := dashSpace / 2
	rightDash := dashSpace - leftDash

	blankRow := leftIndent + color("│") + strings.Repeat(" ", innerWidth) + color("│")

	var b strings.Builder

	for i := 0; i < m.Top; i++ {
		b.WriteByte('\n')
	}

	b.WriteString(leftIndent)
	b.WriteString(color("╭" + strings.Repeat("─", leftDash)))
	b.WriteString(title)
	b.WriteString(color(strings.Repeat("─", rightDash) + "╮"))
	b.WriteByte('\n')

	for i := 0; i < padding; i++ {
		b.WriteString(blankRow)
		b.WriteByte('\n')
	}

	for _, l := range lines {
		pad := maxWidth - visibleLen(l)
		b.WriteString(leftIndent)
		b.WriteString(color("│"))
		b.WriteString(strings.Repeat(" ", padding))
		b.WriteString(l)
		b.WriteString(strings.Repeat(" ", pad+padding))
		b.WriteString(color("│"))
		b.WriteByte('\n')
	}

	for i := 0; i < padding; i++ {
		b.WriteString(blankRow)
		b.WriteByte('\n')
	}

	b.WriteString(leftIndent)
	b.WriteString(color("╰" + strings.Repeat("─", innerWidth) + "╯"))

	for i := 0; i < m.Bottom; i++ {
		b.WriteByte('\n')
	}

	return b.String()
}
