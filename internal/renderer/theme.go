package renderer

// BoxChars defines characters used for box borders, connectors, and arrows.
type BoxChars struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Horizontal  string
	Vertical    string
	TeeDown     string
	TeeUp       string
	TeeRight    string
	TeeLeft     string
	Cross       string
	ArrowRight  string
	ArrowLeft   string
	ArrowDown   string
	ArrowUp     string
	DottedHoriz string
}

var (
	UnicodeBox = BoxChars{
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
		Horizontal:  "─",
		Vertical:    "│",
		TeeDown:     "┬",
		TeeUp:       "┴",
		TeeRight:    "├",
		TeeLeft:     "┤",
		Cross:       "┼",
		ArrowRight:  "►",
		ArrowLeft:   "◄",
		ArrowDown:   "▼",
		ArrowUp:     "▲",
		DottedHoriz: "┄",
	}

	ASCIIBox = BoxChars{
		TopLeft:     "+",
		TopRight:    "+",
		BottomLeft:  "+",
		BottomRight: "+",
		Horizontal:  "-",
		Vertical:    "|",
		TeeDown:     "+",
		TeeUp:       "+",
		TeeRight:    "+",
		TeeLeft:     "+",
		Cross:       "+",
		ArrowRight:  ">",
		ArrowLeft:   "<",
		ArrowDown:   "v",
		ArrowUp:     "^",
		DottedHoriz: "-",
	}
)

// ANSI color escapes
const (
	ColorReset   = "\033[0m"
	ColorBold    = "\033[1m"
	ColorDim     = "\033[2m"
	ColorCyan    = "\033[36m"
	ColorBlue    = "\033[34m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorMagenta = "\033[35m"
	ColorGray    = "\033[90m"
)

// Style applies ANSI styling if enabled.
func Style(text string, styleCode string, enabled bool) string {
	if !enabled || text == "" {
		return text
	}
	return styleCode + text + ColorReset
}
