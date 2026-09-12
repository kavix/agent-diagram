package layout

// RenderMode defines how a diagram is rendered based on terminal geometry and user flags.
type RenderMode string

const (
	ModeAuto    RenderMode = "auto"
	ModeFull    RenderMode = "full"
	ModeCompact RenderMode = "compact"
	ModeNarrow  RenderMode = "narrow"
)
