package dialog

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kirmad/superopencode/internal/tui/theme"
	"github.com/kirmad/superopencode/internal/tui/util"
)

// SlashSuggestionDialog represents a dialog for showing slash command suggestions
type SlashSuggestionDialog struct {
	width        int
	height       int
	suggestions  []Suggestion
	selectedIdx  int
	visible      bool
	processor    *SlashCommandProcessor
}

// SlashSuggestionKeyMap defines key bindings for suggestion navigation
type SlashSuggestionKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Escape key.Binding
	Tab    key.Binding
}

var slashSuggestionKeys = SlashSuggestionKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "ctrl+k"),
		key.WithHelp("↑", "previous suggestion"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "ctrl+j"),
		key.WithHelp("↓", "next suggestion"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select suggestion"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "hide suggestions"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "autofill"),
	),
}

// NewSlashSuggestionDialog creates a new suggestion dialog
func NewSlashSuggestionDialog(processor *SlashCommandProcessor) *SlashSuggestionDialog {
	return &SlashSuggestionDialog{
		processor:   processor,
		selectedIdx: 0,
		visible:     false,
	}
}

// Init initializes the dialog
func (d *SlashSuggestionDialog) Init() tea.Cmd {
	return nil
}

// Update handles messages and key presses
func (d *SlashSuggestionDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !d.visible {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, slashSuggestionKeys.Up):
			if d.selectedIdx > 0 {
				d.selectedIdx--
			} else {
				d.selectedIdx = len(d.suggestions) - 1
			}
		case key.Matches(msg, slashSuggestionKeys.Down):
			if d.selectedIdx < len(d.suggestions)-1 {
				d.selectedIdx++
			} else {
				d.selectedIdx = 0
			}
		case key.Matches(msg, slashSuggestionKeys.Select):
			if len(d.suggestions) > 0 && d.selectedIdx < len(d.suggestions) {
				selected := d.suggestions[d.selectedIdx]
				d.visible = false
				return d, tea.Batch(
					util.CmdHandler(SlashSuggestionSelectedMsg{
						Command: "/" + selected.Command,
					}),
				)
			}
		case key.Matches(msg, slashSuggestionKeys.Escape):
			d.visible = false
		case key.Matches(msg, slashSuggestionKeys.Tab):
			// Tab performs autofill with common prefix
			return d, util.CmdHandler(SlashSuggestionAutofillMsg{})
		}
	}

	return d, nil
}

// View renders the suggestion dialog
func (d *SlashSuggestionDialog) View() string {
	if !d.visible || len(d.suggestions) == 0 {
		return ""
	}

	t := theme.CurrentTheme()
	
	// Base style for the container
	containerStyle := lipgloss.NewStyle().
		Background(t.Background()).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary()).
		Padding(0, 1).
		MaxWidth(d.width).
		MaxHeight(8) // Limit to 8 suggestions max

	var items []string
	maxItems := 8
	if len(d.suggestions) < maxItems {
		maxItems = len(d.suggestions)
	}

	for i := 0; i < maxItems; i++ {
		suggestion := d.suggestions[i]
		
		// Create the item content
		content := "/" + suggestion.Command
		if suggestion.Title != "" {
			content = "/" + suggestion.Command + " (" + suggestion.Title + ")"
		}
		
		// Style based on selection
		if i == d.selectedIdx {
			content = lipgloss.NewStyle().
				Background(t.Primary()).
				Foreground(t.Background()).
				Render(content)
		} else {
			content = lipgloss.NewStyle().
				Foreground(t.Text()).
				Render(content)
		}
		
		items = append(items, content)
	}

	// Add indicator if there are more items
	if len(d.suggestions) > maxItems {
		moreText := lipgloss.NewStyle().
			Foreground(t.TextMuted()).
			Render("... and " + string(rune(len(d.suggestions)-maxItems)) + " more")
		items = append(items, moreText)
	}

	content := strings.Join(items, "\n")
	return containerStyle.Render(content)
}

// SetSize sets the dialog dimensions
func (d *SlashSuggestionDialog) SetSize(width, height int) {
	d.width = width
	d.height = height
}

// Show displays the suggestions for given input
func (d *SlashSuggestionDialog) Show(input string) {
	if d.processor == nil {
		return
	}
	
	d.suggestions = d.processor.GetSuggestions(input, 20)
	d.selectedIdx = 0
	d.visible = len(d.suggestions) > 0
}

// Hide hides the suggestions dialog
func (d *SlashSuggestionDialog) Hide() {
	d.visible = false
	d.suggestions = nil
}

// IsVisible returns whether the dialog is currently visible
func (d *SlashSuggestionDialog) IsVisible() bool {
	return d.visible && len(d.suggestions) > 0
}

// GetSelectedSuggestion returns the currently selected suggestion
func (d *SlashSuggestionDialog) GetSelectedSuggestion() *Suggestion {
	if !d.visible || len(d.suggestions) == 0 || d.selectedIdx >= len(d.suggestions) {
		return nil
	}
	return &d.suggestions[d.selectedIdx]
}

// Messages for slash suggestion events
type SlashSuggestionSelectedMsg struct {
	Command string
}

type SlashSuggestionAutofillMsg struct{}