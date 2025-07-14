package themes

import (
	"github.com/ofri/mde/pkg/theme"
)

// DarkTheme implements a dark theme for the editor
type DarkTheme struct {
	colorScheme theme.ColorScheme
	styles      map[theme.ElementType]theme.Style
}

// NewDarkTheme creates a new dark theme
func NewDarkTheme() *DarkTheme {
	t := &DarkTheme{
		colorScheme: theme.ColorScheme{
			Primary:   "#61AFEF",
			Secondary: "#98C379",
			Accent:    "#E06C75",
			
			Background:    "#282C34",
			BackgroundAlt: "#21252B",
			
			Foreground:    "#ABB2BF",
			ForegroundAlt: "#5C6370",
			
			Success: "#98C379",
			Warning: "#E5C07B",
			Error:   "#E06C75",
			Info:    "#61AFEF",
			
			SyntaxKeyword:  "#C678DD",
			SyntaxString:   "#98C379",
			SyntaxComment:  "#5C6370",
			SyntaxNumber:   "#D19A66",
			SyntaxOperator: "#56B6C2",
			SyntaxFunction: "#61AFEF",
			SyntaxVariable: "#E06C75",
			SyntaxType:     "#E5C07B",
		},
		styles: make(map[theme.ElementType]theme.Style),
	}
	
	t.initializeStyles()
	return t
}

// Name returns the theme name
func (t *DarkTheme) Name() string {
	return "dark"
}

// GetStyle returns the style for a given element type
func (t *DarkTheme) GetStyle(elementType theme.ElementType) theme.Style {
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
func (t *DarkTheme) GetColorScheme() theme.ColorScheme {
	return t.colorScheme
}

// Configure configures the theme with options
func (t *DarkTheme) Configure(options map[string]interface{}) error {
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
func (t *DarkTheme) initializeStyles() {
	// Editor elements
	t.styles[theme.EditorBackground] = theme.Style{
		Background: t.colorScheme.Background,
		Foreground: t.colorScheme.Foreground,
	}
	
	t.styles[theme.EditorForeground] = theme.Style{
		Foreground: t.colorScheme.Foreground,
	}
	
	t.styles[theme.EditorCursor] = theme.Style{
		Background: t.colorScheme.Primary,
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
	}
	
	t.styles[theme.TextBold] = theme.Style{
		Foreground: t.colorScheme.Foreground,
		Bold:       true,
	}
	
	t.styles[theme.TextItalic] = theme.Style{
		Foreground: t.colorScheme.Foreground,
		Italic:     true,
	}
	
	t.styles[theme.TextStrikethrough] = theme.Style{
		Foreground:    t.colorScheme.ForegroundAlt,
		Strikethrough: true,
	}
	
	t.styles[theme.TextCode] = theme.Style{
		Foreground: t.colorScheme.Secondary,
		Background: t.colorScheme.BackgroundAlt,
	}
	
	t.styles[theme.TextLink] = theme.Style{
		Foreground: t.colorScheme.Primary,
		Underline:  true,
	}
	
	// Markdown elements
	t.styles[theme.MarkdownHeading] = theme.Style{
		Foreground: t.colorScheme.Accent,
		Bold:       true,
	}
	
	t.styles[theme.MarkdownHeading1] = theme.Style{
		Foreground: t.colorScheme.Accent,
		Bold:       true,
	}
	
	t.styles[theme.MarkdownHeading2] = theme.Style{
		Foreground: t.colorScheme.Accent,
		Bold:       true,
	}
	
	t.styles[theme.MarkdownHeading3] = theme.Style{
		Foreground: t.colorScheme.Accent,
	}
	
	t.styles[theme.MarkdownHeading4] = theme.Style{
		Foreground: t.colorScheme.Accent,
	}
	
	t.styles[theme.MarkdownHeading5] = theme.Style{
		Foreground: t.colorScheme.Accent,
	}
	
	t.styles[theme.MarkdownHeading6] = theme.Style{
		Foreground: t.colorScheme.Accent,
	}
	
	t.styles[theme.MarkdownBold] = theme.Style{
		Foreground: t.colorScheme.Foreground,
		Bold:       true,
	}
	
	t.styles[theme.MarkdownItalic] = theme.Style{
		Foreground: t.colorScheme.Foreground,
		Italic:     true,
	}
	
	t.styles[theme.MarkdownCode] = theme.Style{
		Foreground: t.colorScheme.Secondary,
		Background: t.colorScheme.BackgroundAlt,
	}
	
	t.styles[theme.MarkdownLink] = theme.Style{
		Foreground: t.colorScheme.Primary,
		Underline:  true,
	}
	
	t.styles[theme.MarkdownLinkText] = theme.Style{
		Foreground: t.colorScheme.Primary,
	}
	
	t.styles[theme.MarkdownLinkURL] = theme.Style{
		Foreground: t.colorScheme.ForegroundAlt,
	}
	
	t.styles[theme.MarkdownImage] = theme.Style{
		Foreground: t.colorScheme.Warning,
	}
	
	t.styles[theme.MarkdownDelimiter] = theme.Style{
		Foreground: t.colorScheme.ForegroundAlt,
	}
	
	t.styles[theme.MarkdownQuote] = theme.Style{
		Foreground: t.colorScheme.ForegroundAlt,
		Italic:     true,
	}
	
	t.styles[theme.MarkdownCodeBlock] = theme.Style{
		Foreground: t.colorScheme.Secondary,
		Background: t.colorScheme.BackgroundAlt,
	}
	
	t.styles[theme.MarkdownTable] = theme.Style{
		Foreground: t.colorScheme.Foreground,
	}
	
	t.styles[theme.MarkdownTableHeader] = theme.Style{
		Foreground: t.colorScheme.Foreground,
		Bold:       true,
	}
	
	t.styles[theme.MarkdownList] = theme.Style{
		Foreground: t.colorScheme.Foreground,
	}
	
	t.styles[theme.MarkdownListItem] = theme.Style{
		Foreground: t.colorScheme.Foreground,
	}
	
	// Syntax highlighting
	t.styles[theme.SyntaxKeyword] = theme.Style{
		Foreground: t.colorScheme.SyntaxKeyword,
		Bold:       true,
	}
	
	t.styles[theme.SyntaxString] = theme.Style{
		Foreground: t.colorScheme.SyntaxString,
	}
	
	t.styles[theme.SyntaxComment] = theme.Style{
		Foreground: t.colorScheme.SyntaxComment,
		Italic:     true,
	}
	
	t.styles[theme.SyntaxNumber] = theme.Style{
		Foreground: t.colorScheme.SyntaxNumber,
	}
	
	t.styles[theme.SyntaxOperator] = theme.Style{
		Foreground: t.colorScheme.SyntaxOperator,
	}
	
	t.styles[theme.SyntaxFunction] = theme.Style{
		Foreground: t.colorScheme.SyntaxFunction,
	}
	
	t.styles[theme.SyntaxVariable] = theme.Style{
		Foreground: t.colorScheme.SyntaxVariable,
	}
	
	t.styles[theme.SyntaxType] = theme.Style{
		Foreground: t.colorScheme.SyntaxType,
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