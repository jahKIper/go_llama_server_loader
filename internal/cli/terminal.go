package cli

import (
	"fmt"
	"os"
)

// ANSI escape codes for terminal color manipulation
const (
	// Set background color to RGB (24-bit truecolor)
	ansiSetBgRGB = "\033[48;2;%d;%d;%dm"
	// Reset all attributes and colors
	ansiReset    = "\033[0m"
)

// HexToRGB converts a hex color string (#RRGGBB or RRGGBB) to RGB values
func HexToRGB(hex string) (uint8, uint8, uint8) {
	// Skip # prefix if present
	if len(hex) >= 1 && hex[0] == '#' {
		hex = hex[1:]
	}

	var r, g, b uint8

	// Parse exactly 6 hex characters
	for i := 0; i < 6 && i < len(hex); i++ {
		var digit uint8
		c := hex[i]
		switch {
		case c >= '0' && c <= '9':
			digit = c - '0'
		case c >= 'a' && c <= 'f':
			digit = c - 'a' + 10
		case c >= 'A' && c <= 'F':
			digit = c - 'A' + 10
		default:
			return 0, 0, 0 // Invalid hex character
		}

		switch i {
		case 0: // R high nibble
			r += digit * 16
		case 1: // R low nibble
			r += digit
		case 2: // G high nibble
			g += digit * 16
		case 3: // G low nibble
			g += digit
		case 4: // B high nibble
			b += digit * 16
		case 5: // B low nibble
			b += digit
		}
	}

	return r, g, b
}

// SetTerminalBackground sends ANSI escape code to set terminal background color
func SetTerminalBackground(hexColor string) {
	r, g, b := HexToRGB(hexColor)
	fmt.Fprintf(os.Stdout, ansiSetBgRGB, r, g, b)
	fmt.Fprintf(os.Stderr, ansiSetBgRGB, r, g, b)
}

// ResetTerminalBackground sends ANSI escape code to reset terminal background
func ResetTerminalBackground() {
	fmt.Fprint(os.Stdout, ansiReset)
	fmt.Fprint(os.Stderr, ansiReset)
}
