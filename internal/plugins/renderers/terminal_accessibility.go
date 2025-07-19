package renderers

import (
	"os"
)

// shouldUseColor checks if colors should be used based on environment
func shouldUseColor() bool {
	// Respect NO_COLOR environment variable
	// See https://no-color.org/
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	
	// Could also check TERM variable for "dumb" terminal
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	
	return true
}

// getAccessibleColor returns a color with better contrast
// This maps potentially problematic colors to more accessible alternatives
func getAccessibleColor(color string) string {
	if !shouldUseColor() {
		return "" // No color
	}
	
	// Map low-contrast colors to higher-contrast alternatives
	switch color {
	case ColorGray: // Bright black (8) - often too dark
		return ColorWhite // Use white (7) for better contrast
	case ColorBlue: // Blue (4) - often too dark for links
		return ColorBrightBlue // Use bright blue (12)
	case ColorYellow: // Yellow (3) - poor on light backgrounds
		return ColorBrightYellow // Use bright yellow (11)
	default:
		return color
	}
}

// getHighContrastColor returns high contrast colors for accessibility mode
func getHighContrastColor(color string) string {
	if !shouldUseColor() {
		return ""
	}
	
	// In high contrast mode, use only the brightest colors
	switch color {
	case ColorBlack, ColorRed, ColorGreen, ColorYellow, 
	     ColorBlue, ColorMagenta, ColorCyan, ColorWhite:
		// Convert all normal colors to their bright variants
		return ColorBrightWhite
	default:
		return color
	}
}