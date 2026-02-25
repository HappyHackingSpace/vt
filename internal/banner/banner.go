// Package banner provides ASCII art banner and colorful text utilities for the vulnerable target application.
package banner

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	mathrand "math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/term"
)

const (
	// AppName is the application name.
	AppName = "vt"
	// AppVersion is the application version.
	AppVersion = "v0.0.1"
)

// DANGER (#FF3355) base RGB — matching JS DANGER_RGB.
const (
	dangerR = 255
	dangerG = 51
	dangerB = 85
)

const (
	white = "\033[38;2;255;255;255m"
	gray  = "\033[90m"
	reset = "\033[0m"
)

// Quote represents a motivational quote with its author.
type Quote struct {
	Text   string
	Author string
}

var quotesList = []Quote{
	{Text: "Pirêze Hayat, Doxrî Yașanmaz.", Author: "Pișo Meheme"},
	{Text: "Talk is cheap. Show me the code.", Author: "Linus Torvalds"},
	{Text: "Given enough eyeballs, all bugs are shallow.", Author: "Eric S. Raymond"},
	{Text: "The quieter you become, the more you are able to hear.", Author: "Anonymous"},
	{Text: "Hack the planet!", Author: "Hackers (1995)"},
	{Text: "Code is poetry.", Author: "WP Community"},
	{Text: "Think like a hacker, act like an engineer.", Author: "Security Community"},
	{Text: "Open source is power.", Author: "Open Source Advocates"},
	{Text: "Information wants to be free.", Author: "Stewart Brand"},
}

// Character set matching JS RETICLE_CHARS — 20 katakana + 16 hex digits.
var matrixChars = []rune("アイウエオカキクケコサシスセソタチツテト0123456789ABCDEF")

const (
	reticleWidth   = 80
	reticleHeight  = 26
	reticleCenterX = 58
	reticleCenterY = 13
	maxRadius      = 12.0
	textZoneEnd    = 36
)

// Ring band definitions — matching JS RING_BANDS exactly.
type ringBand struct {
	inner, outer, bright float64
}

var ringBands = []ringBand{
	{0.20, 0.26, 220}, // innermost ring  (center 0.23, width 0.06)
	{0.44, 0.50, 190}, // second ring     (center 0.47, width 0.06)
	{0.68, 0.74, 150}, // third ring      (center 0.71, width 0.06)
	{0.90, 1.00, 100}, // outermost ring  (center 0.95, width 0.10)
}

// Tick mark ring centers — matching JS TICK_RADII.
var tickRadii = [4]float64{0.23, 0.47, 0.71, 0.95}

type reticleChar struct {
	y, x int
	ch   rune
	ansi string
	dist float64
}

var reticleChars []reticleChar

func init() {
	reticleChars = generateReticle()
}

func isWide(r rune) bool {
	return r >= 0x30A0 && r <= 0x30FF
}

func visibleWidth(s string) int {
	inEsc := false
	w := 0
	for _, r := range s {
		if r == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		if isWide(r) {
			w += 2
		} else {
			w++
		}
	}
	return w
}

func rgbAnsi(r, g, b float64) string {
	ri := int(math.Max(0, math.Min(255, r)))
	gi := int(math.Max(0, math.Min(255, g)))
	bi := int(math.Max(0, math.Min(255, b)))
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", ri, gi, bi)
}

// reticleMask returns brightness (0–255) at a grid position.
// Ported from vulnerable-target.js reticleMask().
func reticleMask(x, y int) float64 {
	dx := (float64(x) - float64(reticleCenterX)) / 2.0 // aspect-ratio correction
	dy := float64(y) - float64(reticleCenterY)
	dist := math.Sqrt(dx*dx + dy*dy)
	nd := dist / maxRadius

	// Ring bands
	for _, band := range ringBands {
		if nd >= band.inner && nd <= band.outer {
			return band.bright
		}
	}

	// Bullseye hot spot
	if nd < 0.08 {
		return 255
	}

	// Crosshair lines (with center gap at 0.10r)
	adx := math.Abs(dx) / maxRadius
	ady := math.Abs(dy) / maxRadius
	if ady < 0.02 && adx > 0.10 && adx < 1.05 {
		return 110
	}
	if adx < 0.02 && ady > 0.10 && ady < 1.05 {
		return 110
	}

	// Tick marks — every 10° around each ring, skip cardinal directions
	angle := math.Atan2(dy, dx)
	for _, tr := range tickRadii {
		if math.Abs(nd-tr) < 0.045 {
			for t := range 36 {
				if t%9 == 0 {
					continue
				}
				ta := float64(t) / 36.0 * math.Pi * 2
				ad := math.Abs(angle - ta)
				if ad > math.Pi {
					ad = math.Pi*2 - ad
				}
				if ad < 0.06 {
					return 80
				}
			}
		}
	}

	// Corner brackets at ±1.08r
	nx := dx / maxRadius
	ny := dy / maxRadius
	for _, sx := range [2]float64{-1, 1} {
		for _, sy := range [2]float64{-1, 1} {
			bx := sx * 1.08
			by := sy * 1.08
			if math.Abs(ny-by) < 0.025 && nx*sx <= bx*sx && nx*sx >= (bx-sx*0.12)*sx {
				return 90
			}
			if math.Abs(nx-bx) < 0.025 && ny*sy <= by*sy && ny*sy >= (by-sy*0.12)*sy {
				return 90
			}
		}
	}

	// Diagonal notches at 45° on outer ring
	if nd > 0.94 && nd < 1.06 {
		for _, da := range [4]float64{math.Pi / 4, 3 * math.Pi / 4, -3 * math.Pi / 4, -math.Pi / 4} {
			ad := math.Abs(angle - da)
			if ad > math.Pi {
				ad = math.Pi*2 - ad
			}
			if ad < 0.05 {
				return 100
			}
		}
	}

	return 0
}

// generateReticle builds the character grid.
// Ported from vulnerable-target.js ReticleCanvas rendering logic.
func generateReticle() []reticleChar {
	layoutRng := mathrand.New(mathrand.NewSource(42)) // #nosec G404 -- decorative pattern
	drawRng := mathrand.New(mathrand.NewSource(77))   // #nosec G404 -- decorative pattern
	var result []reticleChar
	occupied := make([][]bool, reticleHeight)
	for y := range reticleHeight {
		occupied[y] = make([]bool, reticleWidth+2)
	}

	for y := range reticleHeight {
		for x := range reticleWidth {
			if occupied[y][x] {
				continue
			}

			bv := reticleMask(x, y)
			dx := (float64(x) - float64(reticleCenterX)) / 2.0
			dy := float64(y) - float64(reticleCenterY)
			dist := math.Sqrt(dx*dx + dy*dy)

			ch := matrixChars[layoutRng.Intn(len(matrixChars))]

			if bv <= 0 {
				// Background scatter — approximates JS animated noise layer
				if layoutRng.Float64() < 0.05 {
					w := 1
					if isWide(ch) {
						w = 2
					}
					if x+w > reticleWidth {
						continue
					}
					if w == 2 && occupied[y][x+1] {
						continue
					}
					ansi := rgbAnsi(24, 24, 24)
					result = append(result, reticleChar{y, x, ch, ansi, dist})
					occupied[y][x] = true
					if w == 2 {
						occupied[y][x+1] = true
					}
				}
				continue
			}

			// Density gating + alpha — matching JS lines 193-203
			var alpha float64
			if bv > 180 {
				alpha = (200 + drawRng.Float64()*55) / 255
			} else if bv > 120 {
				if drawRng.Float64() < 0.8 {
					alpha = (120 + drawRng.Float64()*90) / 255
				}
			} else if bv > 60 {
				if drawRng.Float64() < 0.5 {
					alpha = (60 + drawRng.Float64()*80) / 255
				}
			} else if bv > 25 {
				if drawRng.Float64() < 0.2 {
					alpha = (25 + drawRng.Float64()*45) / 255
				}
			}

			if alpha <= 0 {
				continue
			}

			// Per-character RGB with slight variation — matching JS lines 206-208
			rv := math.Min(255, float64(dangerR)+(drawRng.Float64()-0.5)*30)
			gv := math.Min(255, float64(dangerG)+(drawRng.Float64()-0.5)*12)

			// Effective RGB on black = color * alpha
			fr := rv * alpha
			fg := gv * alpha
			fb := float64(dangerB) * alpha

			// Bullseye glow — approximates JS radial gradient overlay
			nd := dist / maxRadius
			if nd < 0.08 {
				fr += 40
				fg += 50
				fb += 40
			}

			w := 1
			if isWide(ch) {
				w = 2
			}
			if x+w > reticleWidth {
				continue
			}
			if w == 2 && occupied[y][x+1] {
				continue
			}

			ansi := rgbAnsi(fr, fg, fb)
			result = append(result, reticleChar{y, x, ch, ansi, dist})
			occupied[y][x] = true
			if w == 2 {
				occupied[y][x+1] = true
			}
		}
	}

	return result
}

func randomQuote() string {
	if len(quotesList) == 0 {
		return ""
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(quotesList))))
	if err != nil {
		return ""
	}
	q := quotesList[n.Int64()]
	return fmt.Sprintf("\033[3m%s\033[0m — %s", q.Text, q.Author)
}

func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd())) // #nosec G115 -- stdout fd will not overflow int
}

func flush() {
	if err := os.Stdout.Sync(); err != nil {
		return
	}
}

func moveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

func hideCursor() {
	fmt.Print("\033[?25l")
}

func showCursor() {
	fmt.Print("\033[?25h")
}

func printColored(text, color string) {
	fmt.Print(color + text + reset)
}

// Banner returns the ASCII art banner (non-animated).
func Banner() string {
	var b strings.Builder
	quote := randomQuote()

	textRows := map[int]string{
		3: white + "  VT" + reset,
		5: gray + "  vulnerable target " + AppVersion + reset,
		7: gray + "  // spin up vulnerable targets" + reset,
		8: gray + "  from your terminal //" + reset,
	}

	byLine := make(map[int][]reticleChar)
	for _, c := range reticleChars {
		byLine[c.y] = append(byLine[c.y], c)
	}
	for y := range byLine {
		sort.Slice(byLine[y], func(i, j int) bool {
			return byLine[y][i].x < byLine[y][j].x
		})
	}

	for y := range reticleHeight {
		text, hasText := textRows[y]
		col := 0
		if hasText {
			b.WriteString(text)
			col = visibleWidth(text)
		}

		for _, c := range byLine[y] {
			if c.x < textZoneEnd {
				continue
			}
			for col < c.x {
				b.WriteRune(' ')
				col++
			}
			b.WriteString(c.ansi)
			b.WriteRune(c.ch)
			b.WriteString(reset)
			if isWide(c.ch) {
				col += 2
			} else {
				col++
			}
		}
		b.WriteRune('\n')
	}

	b.WriteString("\n")
	b.WriteString("  " + quote + "\n")
	b.WriteString(gray + strings.Repeat("─", 80) + reset + "\n")

	return b.String()
}

// PrintAnimated displays the banner with targeting lock-on animation.
func PrintAnimated() {
	if !isTerminal() {
		fmt.Print(Banner())
		return
	}

	hideCursor()
	defer showCursor()

	quote := randomQuote()
	startRow := 1

	var sorted []reticleChar
	for _, c := range reticleChars {
		if c.x >= textZoneEnd {
			sorted = append(sorted, c)
		}
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].dist > sorted[j].dist
	})

	for _, c := range sorted {
		moveCursor(startRow+c.y, c.x+1)
		printColored(string(c.ch), c.ansi)
		flush()
		time.Sleep(1 * time.Millisecond)
	}

	// VT title
	time.Sleep(200 * time.Millisecond)
	moveCursor(startRow+3, 3)
	for _, ch := range "VT" {
		printColored(string(ch), white)
		flush()
		time.Sleep(40 * time.Millisecond)
	}

	// Subtitle + version
	time.Sleep(80 * time.Millisecond)
	moveCursor(startRow+5, 3)
	for _, ch := range "vulnerable target " + AppVersion {
		printColored(string(ch), gray)
		flush()
		time.Sleep(15 * time.Millisecond)
	}

	// Tagline
	time.Sleep(60 * time.Millisecond)
	moveCursor(startRow+7, 3)
	for _, ch := range "// spin up vulnerable targets" {
		printColored(string(ch), gray)
		flush()
		time.Sleep(10 * time.Millisecond)
	}
	moveCursor(startRow+8, 3)
	for _, ch := range "from your terminal //" {
		printColored(string(ch), gray)
		flush()
		time.Sleep(10 * time.Millisecond)
	}

	// Quote
	time.Sleep(100 * time.Millisecond)
	moveCursor(startRow+reticleHeight+1, 3)
	for _, ch := range quote {
		fmt.Print(string(ch))
		flush()
		time.Sleep(6 * time.Millisecond)
	}

	// Separator
	time.Sleep(80 * time.Millisecond)
	moveCursor(startRow+reticleHeight+2, 1)
	for range 80 {
		printColored("─", gray)
		flush()
		time.Sleep(1 * time.Millisecond)
	}

	moveCursor(startRow+reticleHeight+4, 1)
	fmt.Println()
}

// Print displays the banner to stdout.
func Print() {
	fmt.Print(Banner())
}
