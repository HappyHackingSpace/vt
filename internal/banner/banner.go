// Package banner provides functionality to display the application banner.
package banner

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strings"

	"golang.org/x/term"
)

const (
	// AppName is the name of the application.
	AppName = "vt"
	// AppVersion is the current version of the application.
	AppVersion = "v0.0.7"
)

// ANSI color codes
const (
// yellow  = "\033[38;2;255;216;43m" // HHS Yellow #ffd82b (Unused)
// cyan    = "\033[38;2;0;255;255m"  // Cyan for text (Unused)
// magenta = "\033[38;2;255;0;255m"  // Magenta for accents (Unused)
// gray    = "\033[90m"              // Gray for dim text (Unused)
// reset   = "\033[0m"               // (Unused)
)

var rainbowColors = []string{
	"\033[31m", // Red
	"\033[33m", // Yellow
	"\033[32m", // Green
	"\033[36m", // Cyan
	"\033[34m", // Blue
	"\033[35m", // Magenta
}

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

// RainbowText applies rainbow colors to the input text.
func RainbowText(text string) string {
	runes := []rune(text)
	var b strings.Builder
	for i, r := range runes {
		color := rainbowColors[i%len(rainbowColors)]
		fmt.Fprintf(&b, "%s%c", color, r)
	}
	b.WriteString("\033[0m") // reset
	return b.String()
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
	return fmt.Sprintf("%s — %s", q.Text, q.Author)
}

// Banner returns the ASCII art banner with application information.
func Banner() string {
	quote := randomQuote()
	return `
 HHS     HHS HHSHHSHHSHHS
 HHS     HHS     HHS
 HHS     HHS     HHS
 HHSx   xHHS     HHS     ` + " " + RainbowText("- Create vulnerable environment") + `
  xHHS xHHS      HHS     ` + " " + fmt.Sprintf("\033[1;3m- %s\033[0m", quote) + `
   HHSHHS        HHS
    HHHH         HHS
     HHS         HHS     			   ` + fmt.Sprintf("\033[1m%s\033[0m", AppVersion) + `
`
}

// Print displays the banner to stdout.
func Print() {
	if isTerminal() {
		fmt.Print(Banner())
	}
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
