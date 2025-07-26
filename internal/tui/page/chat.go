package page

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kirmad/superopencode/internal/app"
	"github.com/kirmad/superopencode/internal/completions"
	"github.com/kirmad/superopencode/internal/logging"
	"github.com/kirmad/superopencode/internal/message"
	"github.com/kirmad/superopencode/internal/session"
	"github.com/kirmad/superopencode/internal/tui/components/chat"
	"github.com/kirmad/superopencode/internal/tui/components/dialog"
	"github.com/kirmad/superopencode/internal/tui/layout"
	"github.com/kirmad/superopencode/internal/tui/util"
)

var ChatPage PageID = "chat"

// slashNamedArgPattern is used to find named arguments in command content
var slashNamedArgPattern = regexp.MustCompile(`\$([A-Z][A-Z0-9_]*)`)

// CommandSetter interface for models that can accept commands
type CommandSetter interface {
	SetCommands(commands []dialog.Command)
}

type chatPage struct {
	app                        *app.App
	editor                     layout.Container
	messages                   layout.Container
	layout                     layout.SplitPaneLayout
	session                    session.Session
	completionDialog           dialog.CompletionDialog
	showCompletionDialog       bool
	commands                   []dialog.Command // Commands for slash command processing
	slashProcessor             *dialog.SlashCommandProcessor
	slashSuggestionDialog      *dialog.SlashSuggestionDialog
	showSlashSuggestions       bool
	dangerouslySkipPermissions bool
}

type ChatKeyMap struct {
	ShowCompletionDialog key.Binding
	NewSession           key.Binding
	Cancel               key.Binding
}

var keyMap = ChatKeyMap{
	ShowCompletionDialog: key.NewBinding(
		key.WithKeys("@"),
		key.WithHelp("@", "Complete"),
	),
	NewSession: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("ctrl+n", "new session"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
}

func (p *chatPage) Init() tea.Cmd {
	cmds := []tea.Cmd{
		p.layout.Init(),
		p.completionDialog.Init(),
	}
	if p.slashSuggestionDialog != nil {
		cmds = append(cmds, p.slashSuggestionDialog.Init())
	}
	return tea.Batch(cmds...)
}

func (p *chatPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmd := p.layout.SetSize(msg.Width, msg.Height)
		cmds = append(cmds, cmd)
	case dialog.CompletionDialogCloseMsg:
		p.showCompletionDialog = false
	case dialog.SlashSuggestionSelectedMsg:
		// Handle slash suggestion selection
		p.showSlashSuggestions = false
		return p, util.CmdHandler(chat.ReplaceInputMsg{Text: msg.Command})
	case dialog.SlashSuggestionAutofillMsg:
		// Handle autofill request
		if p.slashProcessor != nil {
			return p, util.CmdHandler(chat.GetCurrentInputMsg{})
		}
	case chat.CurrentInputMsg:
		// Handle autofill with common prefix
		if p.slashProcessor != nil {
			autofilled := p.slashProcessor.GetCommonPrefix(msg.Text)
			if autofilled != msg.Text {
				return p, util.CmdHandler(chat.ReplaceInputMsg{Text: autofilled})
			}
		}
	case chat.InputChangedMsg:
		// Handle input changes to show/hide suggestions
		if p.slashProcessor != nil && strings.HasPrefix(strings.TrimSpace(msg.Text), "/") {
			if p.slashSuggestionDialog != nil {
				p.slashSuggestionDialog.Show(msg.Text)
				p.showSlashSuggestions = p.slashSuggestionDialog.IsVisible()
			}
		} else {
			p.showSlashSuggestions = false
			if p.slashSuggestionDialog != nil {
				p.slashSuggestionDialog.Hide()
			}
		}
	case chat.SendMsg:
		cmd := p.sendMessage(msg.Text, msg.Attachments)
		if cmd != nil {
			return p, cmd
		}
	case dialog.CommandRunCustomMsg:
		// Check if the agent is busy before executing custom commands
		if p.app.CoderAgent.IsBusy() {
			return p, util.ReportWarn("Agent is busy, please wait before executing a command...")
		}
		
		// Process the command content with arguments if any
		content := msg.Content
		if msg.Args != nil {
			// Replace all named arguments with their values
			for name, value := range msg.Args {
				placeholder := "$" + name
				content = strings.ReplaceAll(content, placeholder, value)
			}
		}
		
		// Handle custom command execution
		cmd := p.sendMessage(content, nil)
		if cmd != nil {
			return p, cmd
		}
	case dialog.ClearSessionMsg:
		// Handle /clear command - clear messages from database and UI
		return p, p.clearSessionAndMessages()
	case chat.SessionSelectedMsg:
		if p.session.ID == "" {
			cmd := p.setSidebar()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		p.session = msg
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keyMap.ShowCompletionDialog):
			p.showCompletionDialog = true
			// Continue sending keys to layout->chat
		case key.Matches(msg, keyMap.NewSession):
			return p, p.clearSessionAndMessages()
		case key.Matches(msg, keyMap.Cancel):
			if p.session.ID != "" {
				// Cancel the current session's generation process
				// This allows users to interrupt long-running operations
				p.app.CoderAgent.Cancel(p.session.ID)
				return p, nil
			}
		}
	}
	if p.showCompletionDialog {
		context, contextCmd := p.completionDialog.Update(msg)
		p.completionDialog = context.(dialog.CompletionDialog)
		cmds = append(cmds, contextCmd)

		// Doesn't forward event if enter key is pressed
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "enter" {
				return p, tea.Batch(cmds...)
			}
		}
	}

	// Handle slash suggestions dialog
	if p.showSlashSuggestions && p.slashSuggestionDialog != nil {
		model, cmd := p.slashSuggestionDialog.Update(msg)
		p.slashSuggestionDialog = model.(*dialog.SlashSuggestionDialog)
		cmds = append(cmds, cmd)

		// Intercept navigation keys for suggestions
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "up", "down", "ctrl+k", "ctrl+j", "enter", "tab", "esc":
				return p, tea.Batch(cmds...)
			}
		}
	}

	u, cmd := p.layout.Update(msg)
	cmds = append(cmds, cmd)
	p.layout = u.(layout.SplitPaneLayout)

	return p, tea.Batch(cmds...)
}

func (p *chatPage) setSidebar() tea.Cmd {
	sidebarContainer := layout.NewContainer(
		chat.NewSidebarCmp(p.session, p.app.History),
		layout.WithPadding(1, 1, 1, 1),
	)
	return tea.Batch(p.layout.SetRightPanel(sidebarContainer), sidebarContainer.Init())
}

func (p *chatPage) clearSidebar() tea.Cmd {
	return p.layout.ClearRightPanel()
}

// clearSessionAndMessages clears both the UI session and the database messages for true context clearing
func (p *chatPage) clearSessionAndMessages() tea.Cmd {
	sessionID := p.session.ID
	if sessionID != "" {
		// Clear messages from database to remove LLM context
		go func() {
			ctx := context.Background()
			err := p.app.Messages.DeleteSessionMessages(ctx, sessionID)
			if err != nil {
				// Log error but don't block the UI clearing
				fmt.Printf("Warning: failed to clear session messages: %v\n", err)
			}
		}()
	}
	// Clear the UI session state
	p.session = session.Session{}
	return tea.Batch(
		p.clearSidebar(),
		util.CmdHandler(chat.SessionClearedMsg{}),
	)
}

func (p *chatPage) sendMessage(text string, attachments []message.Attachment) tea.Cmd {
	var cmds []tea.Cmd
	
	// Check for slash command before processing
	if p.slashProcessor != nil && p.slashProcessor.IsSlashCommand(text) {
		return p.handleSlashCommand(text, attachments)
	}
	
	if p.session.ID == "" {
		session, err := p.app.Sessions.Create(context.Background(), "New Session")
		if err != nil {
			return util.ReportError(err)
		}

		// Auto-approve permissions if dangerous flag is set
		if p.dangerouslySkipPermissions {
			logging.Warn("⚠️ DANGEROUS: --dangerously-skip-permissions active. All tool permissions bypassed for interactive session %s", session.ID)
			p.app.Permissions.AutoApproveSession(session.ID)
		}

		p.session = session
		cmd := p.setSidebar()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, util.CmdHandler(chat.SessionSelectedMsg(session)))
	}

	_, err := p.app.CoderAgent.Run(context.Background(), p.session.ID, text, attachments...)
	if err != nil {
		return util.ReportError(err)
	}
	return tea.Batch(cmds...)
}

// handleSlashCommand processes slash commands
func (p *chatPage) handleSlashCommand(text string, attachments []message.Attachment) tea.Cmd {
	// Check if agent is busy before executing slash commands
	if p.app.CoderAgent.IsBusy() {
		return util.ReportWarn("Agent is busy, please wait before executing a command...")
	}

	// Validate slash command syntax
	if err := dialog.ValidateSlashCommand(text); err != nil {
		return util.ReportError(err)
	}

	result := p.slashProcessor.ProcessSlashCommand(text)
	if result.Error != nil {
		// Extract command name for better error message
		commandName := strings.TrimSpace(text)
		if strings.HasPrefix(commandName, "/") {
			parts := strings.SplitN(commandName[1:], " ", 2)
			if len(parts) > 0 {
				commandName = parts[0]
			}
		}
		errorMsg := dialog.FormatSlashCommandError(result.Error, commandName)
		return util.ReportError(fmt.Errorf(errorMsg))
	}

	// If the command needs arguments dialog, show it
	if result.NeedsArgDialog {
		// Extract argument names from the command content
		matches := slashNamedArgPattern.FindAllStringSubmatch(result.Processed.Content, -1)
		argNames := make([]string, 0)
		argMap := make(map[string]bool)

		for _, match := range matches {
			argName := match[1] // Group 1 is the name without $
			if !argMap[argName] {
				argMap[argName] = true
				argNames = append(argNames, argName)
			}
		}

		// Show multi-arguments dialog
		return util.CmdHandler(dialog.ShowMultiArgumentsDialogMsg{
			CommandID: result.Processed.Command.ID,
			Content:   result.Processed.Content,
			ArgNames:  argNames,
		})
	}

	// Execute the command directly with combined content
	return p.sendMessage(result.Processed.Content, attachments)
}

func (p *chatPage) SetSize(width, height int) tea.Cmd {
	return p.layout.SetSize(width, height)
}

func (p *chatPage) GetSize() (int, int) {
	return p.layout.GetSize()
}

func (p *chatPage) View() string {
	layoutView := p.layout.View()

	if p.showCompletionDialog {
		_, layoutHeight := p.layout.GetSize()
		editorWidth, editorHeight := p.editor.GetSize()

		p.completionDialog.SetWidth(editorWidth)
		overlay := p.completionDialog.View()

		layoutView = layout.PlaceOverlay(
			0,
			layoutHeight-editorHeight-lipgloss.Height(overlay),
			overlay,
			layoutView,
			false,
		)
	}

	// Show slash suggestions dialog
	if p.showSlashSuggestions && p.slashSuggestionDialog != nil {
		_, layoutHeight := p.layout.GetSize()
		editorWidth, editorHeight := p.editor.GetSize()

		p.slashSuggestionDialog.SetSize(editorWidth, 10)
		overlay := p.slashSuggestionDialog.View()

		if overlay != "" {
			layoutView = layout.PlaceOverlay(
				0,
				layoutHeight-editorHeight-lipgloss.Height(overlay),
				overlay,
				layoutView,
				false,
			)
		}
	}

	return layoutView
}

func (p *chatPage) BindingKeys() []key.Binding {
	bindings := layout.KeyMapToSlice(keyMap)
	bindings = append(bindings, p.messages.BindingKeys()...)
	bindings = append(bindings, p.editor.BindingKeys()...)
	return bindings
}

func NewChatPage(app *app.App, dangerouslySkipPermissions bool) tea.Model {
	cg := completions.NewFileAndFolderContextGroup()
	completionDialog := dialog.NewCompletionDialogCmp(cg)

	messagesContainer := layout.NewContainer(
		chat.NewMessagesCmp(app),
		layout.WithPadding(1, 1, 0, 1),
	)
	editorContainer := layout.NewContainer(
		chat.NewEditorCmp(app),
		layout.WithBorder(true, false, false, false),
	)
	return &chatPage{
		app:                        app,
		editor:                     editorContainer,
		messages:                   messagesContainer,
		completionDialog:           completionDialog,
		commands:                   nil, // Will be set later via SetCommands
		slashProcessor:             nil, // Will be created when commands are set
		dangerouslySkipPermissions: dangerouslySkipPermissions,
		layout: layout.NewSplitPane(
			layout.WithLeftPanel(messagesContainer),
			layout.WithBottomPanel(editorContainer),
		),
	}
}

// SetCommands sets the commands for slash command processing
func (p *chatPage) SetCommands(commands []dialog.Command) {
	p.commands = commands
	if len(commands) > 0 {
		p.slashProcessor = dialog.NewSlashCommandProcessor(commands)
		p.slashSuggestionDialog = dialog.NewSlashSuggestionDialog(p.slashProcessor)
	}
}
