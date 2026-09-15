# CLI & Terminal Standards for Representing UML Diagrams

A comprehensive survey of international standards, character encoding conventions, terminal protocols, and layout architectures for rendering Unified Modeling Language (UML) diagrams in CLI environments.

---

## 1. Executive Summary

Representing graphical models like UML diagrams within text-mode terminals presents a unique challenge: the terminal is a discrete, cell-based character grid with variable dimensions, fixed monospaced fonts, and non-proportional aspect ratios (typically ~1:2 height-to-width per character cell).

Historically, software engineering documentation solved this through three primary paradigms:
1. **Telecommunication standards** (ITU-T Z.120 Message Sequence Charts).
2. **Internet Engineering Task Force (IETF) RFC specifications** (72-column ASCII sequence charts).
3. **Unicode Box-Drawing Standards** (ISO/IEC 10646 / Unicode 1.0–16.0).

Modern AI coding agents require a 4th generation: **Width-Adaptive Semantic Layouts**, where the layout engine treats the terminal geometry as a constraint and dynamically alters the visualization's dimensionality.

---

## 2. Formal Governing Standards & Specifications

### 2.1. ITU-T Recommendation Z.120 (Message Sequence Charts)
- **Standard**: ITU-T Z.120 (first standardized in 1993, updated through Z.120 (02/99)).
- **Relevance**: Predecessor to UML Sequence Diagrams. Z.120 formally defined the textual and graphical interchange format for message sequences in distributed systems.
- **Key Principles**:
  - **Lifelines**: Vertical timelines representing process instances (`instance Client; ... endinstance`).
  - **Messages**: Directed horizontal vectors representing asynchronous or synchronous message passing.
  - **Ordering**: Strict partial ordering along the vertical axis (time flows top-to-bottom).
  - **Co-regions**: Regions where event ordering is unordered.

### 2.2. OMG UML Specification (Version 2.5.1)
- **Standard**: Object Management Group (OMG) Formal Specification `formal/2017-12-05`.
- **Key UML Diagram Types in Terminals**:
  - **Interaction / Sequence Diagrams (Section 17)**:
    - Lifeline headers: rectangular box containing `[name : type]`.
    - Lifeline stems: vertical dashed or solid line.
    - Synchronous Call: solid line with filled arrowhead (`►` / `▲`).
    - Asynchronous Signal: solid line with open stick arrowhead (`>` / `→`).
    - Reply / Return Message: dashed line (`┄┄►` / `-->`).
    - Execution Specification / Activation: vertical rectangular box overlaid on the lifeline.
  - **Activity Diagrams / Flowcharts (Section 12)**:
    - Action nodes: rounded rectangles.
    - Decision / Merge nodes: diamond symbols (`◇`).
    - Fork / Join nodes: synchronization bars (`━━━`).
    - Control flows: solid directed connectors.
  - **Class / Struct Diagrams (Section 10)**:
    - Three-compartment rectangular boxes: `[ Name | Attributes | Operations ]`.

### 2.3. IETF RFC Standards (RFC 7322 / RFC 2360 / RFC 2821 / RFC 3261)
- **Standard**: IETF RFC Publication Style Guidelines.
- **The 72-Column Constraint**: IETF RFCs mandate that all line lengths MUST NOT exceed 72 characters (to accommodate pagination and email quoting).
- **Canonical Sequence Format**:
  Used extensively in protocol RFCs (e.g. SIP in RFC 3261, SMTP in RFC 2821, ICE in RFC 5245):
  ```text
       Alice               Proxy                Bob
         |                   |                   |
         |  INVITE F1        |                   |
         |------------------>|                   |
         |                   |  INVITE F2        |
         |                   |------------------>|
         |                   |  200 OK F3        |
         |                   |<------------------|
         |  200 OK F4        |                   |
         |<------------------|                   |
  ```
- **Conventions established by IETF**:
  - Minimum column spacing: 4–6 characters.
  - Centered participant headers.
  - Message identifiers (`F1`, `F2`) or numbers to prevent wrapping within lifelines.

---

## 3. Terminal Character Encodings & Cell Metrics

To achieve clean visual representations, terminal renderers must adhere to Unicode cell measurement standards.

### 3.1. Unicode Standard Annex #11 (East Asian Width)
A common defect in terminal renderers is computing string length in bytes (`len(str)`) or runes (`len([]rune(str))`). 
- On terminal emulators, characters have visual widths:
  - Half-width / Narrow (1 cell): ASCII, Latin, Greek, Cyrillic, Box-drawing characters.
  - Full-width / Wide (2 cells): CJK characters, Emoji, certain mathematical symbols.
  - Zero-width: Combining accents, ANSI escape sequences, zero-width joiners.
- **Standard Requirement**: Renderers must measure cell width using `UAX #11` (implemented via `golang.org/x/text/width` or `github.com/mattn/go-runewidth`).

### 3.2. Unicode Box-Drawing Block (`U+2500` – `U+257F`)
The ISO/IEC 10646 standard defines 128 box-drawing characters:
- **Light Box Elements**: `─ │ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼` (Universal baseline).
- **Heavy / Thick Elements**: `━ ┃ ┏ ┓ ┗ ┛ ┣ ┫ ┳ ┻ ╋` (Used for active states or thick flows).
- **Double-Line Elements**: `═ ║ ╔ ╗ ╚ ╝ ╠ ╣ ╦ ╩ ╬` (Used for system boundaries or active processes).
- **Rounded Corners**: `╭ ╮ ╯ ╰` (`U+256D`–`U+2570`) (Standard for rounded activity / action nodes).
- **Dotted / Dashed Lines**: `┄ ┅ ┆ ┇ ┈ ┉ ┊ ┋` (Standard for reply / return lifelines).

### 3.3. Geometric Shapes & Arrow Blocks
- **Arrows (`U+2190` – `U+21FF`)**: `← ↑ → ↓ ↔ ↕ ➔ ➜ ◄ ► ▲ ▼`.
  - Preferred terminal sequence arrows: `►` (`U+25BA`) and `◄` (`U+25C4`) from Geometric Shapes, because their glyph aspect ratios match single character cells better than thin arrows on most font engines.
- **Decision Diamonds (`U+25C6`, `U+25C7`, `U+25C8`)**:
  - `◇` (`U+25C7` White Diamond) or `◈` (`U+25C8` White Diamond Containing Small Black Diamond) for branching gates.

### 3.4. 7-bit US-ASCII Fallback (ANSI X3.4)
For systems lacking UTF-8 locales (e.g. `LANG=C` in bare-metal rescue disks or stripped CI runners):
- Intersections: `+`
- Horizontal: `-`
- Vertical: `|`
- Arrows: `>`, `<`, `^`, `v`

---

## 4. Prior Art & Existing CLI Tools

| Project | Approach | Strengths | Limitations for AI Agents |
| :--- | :--- | :--- | :--- |
| **PlantUML (`-ttxt`)** | Java engine generating ASCII art | Supports broad UML syntax (Class, Sequence, State, Component) | Requires heavy Java Runtime (JRE); slow startup (~1.5s); not width-adaptive. |
| **Graph-Easy** | Perl-based ASCII graph generator | Excellent layout for complex DAGs and flowcharts | Slow startup; heavy Perl CPAN dependency; hard to bundle into AI agents. |
| **Svgbob** | Rust-based ASCII-to-SVG / SVG-to-ASCII | Generates elegant diagrams | Primarily focused on Markdown to SVG conversion rather than terminal adaptive rendering. |
| **Mermaid CLI (`mmdc`)** | Node.js + Puppeteer | Industry standard for web Mermaid | Requires Chromium headless browser (~300MB); cannot render natively to terminal text. |

---

## 5. High-Resolution Terminal Graphics Protocols

When pure Unicode text is insufficient, modern terminal emulators support inline pixel/vector graphics:

### 5.1. Kitty Graphics Protocol
- **Specification**: Designed by Kovid Goyal (Kitty terminal). Supported by Kitty, Ghostty, WezTerm.
- **Transport**: Escape sequence `\x1b_G...;data\x1b\`.
- **Capabilities**: Direct RGB/RGBA 32-bit pixel buffers, PNG transmission, z-index layering.

### 5.2. iTerm2 Inline Images Protocol
- **Specification**: Introduced by George Nachman (iTerm2). Supported by iTerm2, WezTerm, VS Code terminal.
- **Transport**: `\x1b]1337;File=inline=1;width=<w>px:...<base64>...\x07`.
- **Capabilities**: Direct inline display of PNG/JPEG/GIF/PDF files.

### 5.3. DEC Sixel Graphics
- **Specification**: DEC VT330 / VT340 terminal protocol (1980s).
- **Transport**: Escape sequence `\x1bPq...` transmitting 6-pixel vertical slices.
- **Capabilities**: 16–256 color indexed bitmap rendering supported on XTerm, Foot, Mintty, WezTerm.

---

## 6. The `agent-diagram` Standard: Adaptive Terminal UML

`agent-diagram` introduces a formal standard for AI coding assistants:

1. **Three-Tier Adaptive Degradation**:
   - **Tier 1 (Full 2D Spatial Grid)**: When `Width >= RequiredWidth`, render complete 2D boxes, lifelines, and cross-junctions.
   - **Tier 2 (Compact 1D Timeline / Tree)**: When `Width in [50, RequiredWidth)`, preserve all semantic events and arrows in a numbered chronological stream or indented hierarchical tree.
   - **Tier 3 (Narrow Flow)**: When `Width < 50`, collapse into abbreviated token chains `[1] A ─► B: Msg`.

2. **Decoupled Architecture**:
   - Parse input (Mermaid, PlantUML, or JSON) into a pure internal AST.
   - Separate AST from Layout and Renderer so alternative renderers (Unicode, ASCII, Sixel, Kitty) can be added without modifying the parser.

3. **Sub-millisecond Cold Start (< 5ms)**:
   - Single statically linked binary without JVM or Node.js runtime, ensuring AI agents can invoke it synchronously through MCP without stalling the conversational turn.
