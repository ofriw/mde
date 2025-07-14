package theme

import "github.com/charmbracelet/lipgloss"

// Theme defines the interface for editor themes
type Theme interface {
	// Name returns the theme name
	Name() string
	
	// GetStyle returns the style for a given element type
	GetStyle(elementType ElementType) Style
	
	// GetColorScheme returns the color scheme
	GetColorScheme() ColorScheme
	
	// Configure configures the theme with options
	Configure(options map[string]interface{}) error
}

// ElementType represents different types of elements that can be styled
type ElementType int

const (
	// Editor elements
	EditorBackground ElementType = iota
	EditorForeground
	EditorCursor
	EditorSelection
	EditorLineNumber
	EditorCurrentLine
	EditorScrollbar
	
	// Text elements
	TextNormal
	TextBold
	TextItalic
	TextStrikethrough
	TextCode
	TextLink
	
	// Markdown elements
	MarkdownHeading
	MarkdownHeading1
	MarkdownHeading2
	MarkdownHeading3
	MarkdownHeading4
	MarkdownHeading5
	MarkdownHeading6
	MarkdownBold
	MarkdownItalic
	MarkdownCode
	MarkdownCodeBlock
	MarkdownLink
	MarkdownLinkText
	MarkdownLinkURL
	MarkdownImage
	MarkdownQuote
	MarkdownTable
	MarkdownTableHeader
	MarkdownList
	MarkdownListItem
	MarkdownDelimiter
	
	// Syntax highlighting
	SyntaxKeyword
	SyntaxString
	SyntaxComment
	SyntaxNumber
	SyntaxOperator
	SyntaxFunction
	SyntaxVariable
	SyntaxType
	
	// UI elements
	UIStatusBar
	UIHelpBar
	UIScrollbar
	UIBorder
	UIError
	UIWarning
	UIInfo
)

// Style represents styling information for an element
type Style struct {
	// Foreground color
	Foreground string
	
	// Background color
	Background string
	
	// Bold text
	Bold bool
	
	// Italic text
	Italic bool
	
	// Underline text
	Underline bool
	
	// Strikethrough text
	Strikethrough bool
	
	// Additional lipgloss style
	LipglossStyle lipgloss.Style
}

// ColorScheme represents a color scheme
type ColorScheme struct {
	// Primary colors
	Primary   string
	Secondary string
	Accent    string
	
	// Background colors
	Background     string
	BackgroundAlt  string
	
	// Foreground colors
	Foreground    string
	ForegroundAlt string
	
	// State colors
	Success string
	Warning string
	Error   string
	Info    string
	
	// Syntax colors
	SyntaxKeyword  string
	SyntaxString   string
	SyntaxComment  string
	SyntaxNumber   string
	SyntaxOperator string
	SyntaxFunction string
	SyntaxVariable string
	SyntaxType     string
}

// ToLipgloss converts a Style to a lipgloss.Style
func (s Style) ToLipgloss() lipgloss.Style {
	style := lipgloss.NewStyle()
	
	if s.Foreground != "" {
		style = style.Foreground(lipgloss.Color(s.Foreground))
	}
	
	if s.Background != "" {
		style = style.Background(lipgloss.Color(s.Background))
	}
	
	if s.Bold {
		style = style.Bold(true)
	}
	
	if s.Italic {
		style = style.Italic(true)
	}
	
	if s.Underline {
		style = style.Underline(true)
	}
	
	if s.Strikethrough {
		style = style.Strikethrough(true)
	}
	
	// Merge with any additional lipgloss style if it exists
	// (we'll assume it has some styling if it was provided)
	style = style.Inherit(s.LipglossStyle)
	
	return style
}