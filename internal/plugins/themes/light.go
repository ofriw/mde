package themes

import (
	"github.com/ofri/mde/pkg/theme"
)

// LightTheme implements a light theme for the editor
type LightTheme struct {
	colorScheme theme.ColorScheme
	styles      map[theme.ElementType]theme.Style
}

// NewLightTheme creates a new light theme
func NewLightTheme() *LightTheme {
	t := &LightTheme{
		colorScheme: theme.ColorScheme{
			Primary:   "#0366D6",
			Secondary: "#28A745",
			Accent:    "#D73A49",
			
			Background:    "#FFFFFF",
			BackgroundAlt: "#F6F8FA",
			
			Foreground:    "#24292E",
			ForegroundAlt: "#6A737D",
			
			Success: "#28A745",
			Warning: "#FFC107",
			Error:   "#DC3545",
			Info:    "#17A2B8",
			
			SyntaxKeyword:  "#D73A49",
			SyntaxString:   "#032F62",
			SyntaxComment:  "#6A737D",
			SyntaxNumber:   "#005CC5",
			SyntaxOperator: "#D73A49",
			SyntaxFunction: "#6F42C1",
			SyntaxVariable: "#E36209",
			SyntaxType:     "#005CC5",
		},
		styles: make(map[theme.ElementType]theme.Style),
	}
	
	t.initializeStyles()
	return t
}

// Name returns the theme name
func (t *LightTheme) Name() string {
	return "light"
}

// GetStyle returns the style for a given element type
func (t *LightTheme) GetStyle(elementType theme.ElementType) theme.Style {
	if style, exists := t.styles[elementType]; exists {
		return style
	}
	
	// Return default style if not found
	return theme.Style{
		Foreground: t.colorScheme.Foreground,
		Background: t.colorScheme.Background,
	}
}

// GetColorScheme returns the color scheme
func (t *LightTheme) GetColorScheme() theme.ColorScheme {
	return t.colorScheme
}

// Configure configures the theme with options
func (t *LightTheme) Configure(options map[string]interface{}) error {
	// Allow customization of colors
	if primary, ok := options["primary"].(string); ok {
		t.colorScheme.Primary = primary
	}
	
	if secondary, ok := options["secondary"].(string); ok {
		t.colorScheme.Secondary = secondary
	}
	
	if accent, ok := options["accent"].(string); ok {
		t.colorScheme.Accent = accent
	}
	
	// Re-initialize styles with new colors
	t.initializeStyles()
	
	return nil
}

// initializeStyles sets up all the styles for different element types
func (t *LightTheme) initializeStyles() {
	// Editor elements
	t.styles[theme.EditorBackground] = theme.Style{
		Background: t.colorScheme.Background,
		Foreground: t.colorScheme.Foreground,
	}
	
	t.styles[theme.EditorForeground] = theme.Style{
		Foreground: t.colorScheme.Foreground,
	}
	
	t.styles[theme.EditorCursor] = theme.Style{
		Background: t.colorScheme.Foreground,
		Foreground: t.colorScheme.Background,
	}
	
	t.styles[theme.EditorSelection] = theme.Style{
		Background: t.colorScheme.Primary,
		Foreground: t.colorScheme.Background,
	}
	
	t.styles[theme.EditorLineNumber] = theme.Style{
		Foreground: t.colorScheme.ForegroundAlt,
	}
	
	t.styles[theme.EditorCurrentLine] = theme.Style{
		Background: t.colorScheme.BackgroundAlt,
		Foreground: t.colorScheme.Foreground,
	}
	
	// Text elements
	t.styles[theme.TextNormal] = theme.Style{
		Foreground: t.colorScheme.Foreground,
		Background: t.colorScheme.Background,
	}
	
	t.styles[theme.TextBold] = theme.Style{
		Foreground: t.colorScheme.Foreground,
		Background: t.colorScheme.Background,
		Bold:       true,
	}
	
	t.styles[theme.TextItalic] = theme.Style{
		Foreground: t.colorScheme.Foreground,
		Background: t.colorScheme.Background,
		Italic:     true,
	}
	
	t.styles[theme.TextStrikethrough] = theme.Style{
		Foreground:    t.colorScheme.ForegroundAlt,
		Background:    t.colorScheme.Background,
		Strikethrough: true,
	}
	
	t.styles[theme.TextCode] = theme.Style{
		Foreground: t.colorScheme.Secondary,
		Background: t.colorScheme.BackgroundAlt,
	}
	
	t.styles[theme.TextLink] = theme.Style{
		Foreground: t.colorScheme.Primary,
		Background: t.colorScheme.Background,
		Underline:  true,
	}
	
	// Markdown elements
	t.styles[theme.MarkdownHeading] = theme.Style{
		Foreground: t.colorScheme.Accent,
		Background: t.colorScheme.Background,
		Bold:       true,
	}
	
	t.styles[theme.MarkdownHeading1] = theme.Style{
		Foreground: t.colorScheme.Accent,
		Background: t.colorScheme.Background,
		Bold:       true,
	}
	
	t.styles[theme.MarkdownHeading2] = theme.Style{
		Foreground: t.colorScheme.Accent,
		Background: t.colorScheme.Background,
		Bold:       true,
	}
	
	t.styles[theme.MarkdownHeading3] = theme.Style{
		Foreground: t.colorScheme.Accent,
		Background: t.colorScheme.Background,
	}
	
	t.styles[theme.MarkdownHeading4] = theme.Style{
		Foreground: t.colorScheme.Accent,
		Background: t.colorScheme.Background,
	}
	
	t.styles[theme.MarkdownHeading5] = theme.Style{
		Foreground: t.colorScheme.Accent,
		Background: t.colorScheme.Background,
	}
	
	t.styles[theme.MarkdownHeading6] = theme.Style{
		Foreground: t.colorScheme.Accent,
		Background: t.colorScheme.Background,
	}
	
	t.styles[theme.MarkdownBold] = theme.Style{
		Foreground: t.colorScheme.Foreground,
		Background: t.colorScheme.Background,
		Bold:       true,
	}
	
	t.styles[theme.MarkdownItalic] = theme.Style{
		Foreground: t.colorScheme.Foreground,
		Background: t.colorScheme.Background,
		Italic:     true,
	}
	
	t.styles[theme.MarkdownCode] = theme.Style{
		Foreground: t.colorScheme.Secondary,
		Background: t.colorScheme.BackgroundAlt,
	}
	
	t.styles[theme.MarkdownLink] = theme.Style{
		Foreground: t.colorScheme.Primary,
		Background: t.colorScheme.Background,
		Underline:  true,
	}
	
	t.styles[theme.MarkdownLinkText] = theme.Style{
		Foreground: t.colorScheme.Primary,
		Background: t.colorScheme.Background,
	}
	
	t.styles[theme.MarkdownLinkURL] = theme.Style{
		Foreground: t.colorScheme.ForegroundAlt,
		Background: t.colorScheme.Background,
	}
	
	t.styles[theme.MarkdownImage] = theme.Style{
		Foreground: t.colorScheme.Warning,
		Background: t.colorScheme.Background,
	}
	
	t.styles[theme.MarkdownDelimiter] = theme.Style{
		Foreground: t.colorScheme.ForegroundAlt,
		Background: t.colorScheme.Background,
	}
	
	t.styles[theme.MarkdownQuote] = theme.Style{
		Foreground: t.colorScheme.ForegroundAlt,
		Background: t.colorScheme.Background,
		Italic:     true,
	}
	
	t.styles[theme.MarkdownCodeBlock] = theme.Style{
		Foreground: t.colorScheme.Secondary,
		Background: t.colorScheme.BackgroundAlt,
	}
	
	t.styles[theme.MarkdownTable] = theme.Style{
		Foreground: t.colorScheme.Foreground,
		Background: t.colorScheme.Background,
	}
	
	t.styles[theme.MarkdownTableHeader] = theme.Style{
		Foreground: t.colorScheme.Foreground,
		Background: t.colorScheme.Background,
		Bold:       true,
	}
	
	t.styles[theme.MarkdownList] = theme.Style{
		Foreground: t.colorScheme.Foreground,
		Background: t.colorScheme.Background,
	}
	
	t.styles[theme.MarkdownListItem] = theme.Style{
		Foreground: t.colorScheme.Foreground,
		Background: t.colorScheme.Background,
	}
	
	// Syntax highlighting
	t.styles[theme.SyntaxKeyword] = theme.Style{
		Foreground: t.colorScheme.SyntaxKeyword,
		Background: t.colorScheme.Background,
		Bold:       true,
	}
	
	t.styles[theme.SyntaxString] = theme.Style{
		Foreground: t.colorScheme.SyntaxString,
		Background: t.colorScheme.Background,
	}
	
	t.styles[theme.SyntaxComment] = theme.Style{
		Foreground: t.colorScheme.SyntaxComment,
		Background: t.colorScheme.Background,
		Italic:     true,
	}
	
	t.styles[theme.SyntaxNumber] = theme.Style{
		Foreground: t.colorScheme.SyntaxNumber,
		Background: t.colorScheme.Background,
	}
	
	t.styles[theme.SyntaxOperator] = theme.Style{
		Foreground: t.colorScheme.SyntaxOperator,
		Background: t.colorScheme.Background,
	}
	
	t.styles[theme.SyntaxFunction] = theme.Style{
		Foreground: t.colorScheme.SyntaxFunction,
		Background: t.colorScheme.Background,
	}
	
	t.styles[theme.SyntaxVariable] = theme.Style{
		Foreground: t.colorScheme.SyntaxVariable,
		Background: t.colorScheme.Background,
	}
	
	t.styles[theme.SyntaxType] = theme.Style{
		Foreground: t.colorScheme.SyntaxType,
		Background: t.colorScheme.Background,
	}
	
	// UI elements
	t.styles[theme.UIStatusBar] = theme.Style{
		Background: t.colorScheme.BackgroundAlt,
		Foreground: t.colorScheme.Foreground,
	}
	
	t.styles[theme.UIHelpBar] = theme.Style{
		Background: t.colorScheme.BackgroundAlt,
		Foreground: t.colorScheme.ForegroundAlt,
	}
	
	t.styles[theme.UIBorder] = theme.Style{
		Foreground: t.colorScheme.ForegroundAlt,
	}
	
	t.styles[theme.UIError] = theme.Style{
		Foreground: t.colorScheme.Error,
		Bold:       true,
	}
	
	t.styles[theme.UIWarning] = theme.Style{
		Foreground: t.colorScheme.Warning,
		Bold:       true,
	}
	
	t.styles[theme.UIInfo] = theme.Style{
		Foreground: t.colorScheme.Info,
	}
}