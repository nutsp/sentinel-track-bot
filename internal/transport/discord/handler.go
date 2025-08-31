package discord

import (
	"context"
	"fmt"
	"strings"

	"fix-track-bot/internal/domain"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler handles Discord interactions
type Handler struct {
	session        *discordgo.Session
	issueService   domain.IssueService
	channelService domain.ChannelService
	logger         *zap.Logger
}

// NewHandler creates a new Discord handler
func NewHandler(session *discordgo.Session, issueService domain.IssueService, channelService domain.ChannelService, logger *zap.Logger) *Handler {
	return &Handler{
		session:        session,
		issueService:   issueService,
		channelService: channelService,
		logger:         logger,
	}
}

// RegisterHandlers registers all Discord event handlers
func (h *Handler) RegisterHandlers() {
	h.session.AddHandler(h.handleMessageCreate)
	h.session.AddHandler(h.handleInteractionCreate)
}

// handleMessageCreate handles regular Discord messages
func (h *Handler) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	ctx := context.Background()

	// Handle simple ping/pong commands
	switch strings.ToLower(m.Content) {
	case "ping":
		h.sendMessage(ctx, m.ChannelID, "Pong! 🏓")
	case "pong":
		h.sendMessage(ctx, m.ChannelID, "Ping! 🏓")
	}
}

// handleInteractionCreate handles Discord interactions (slash commands, buttons, modals, etc.)
func (h *Handler) handleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		h.handleSlashCommand(ctx, i)
	case discordgo.InteractionModalSubmit:
		h.handleModalSubmit(ctx, i)
	case discordgo.InteractionMessageComponent:
		h.handleMessageComponent(ctx, i)
	}
}

// handleSlashCommand handles slash command interactions
func (h *Handler) handleSlashCommand(ctx context.Context, i *discordgo.InteractionCreate) {
	commandName := i.ApplicationCommandData().Name

	ch, err := h.session.State.Channel(i.ChannelID)
	if err != nil {
		// ถ้า state cache ไม่มี ลอง fetch จาก API ตรงๆ
		ch, err = h.session.Channel(i.ChannelID)
		if err != nil {
			h.logger.Error("Cannot fetch channel", zap.Error(err))
			return
		}
	}

	h.logger.Info("Handling slash command",
		zap.String("command", commandName),
		zap.String("user_id", i.Member.User.ID),
		zap.String("channel_id", i.ChannelID),
		zap.String("channel_name", ch.Name),
		zap.Any("data", ch),
	)

	switch commandName {
	case "issue":
		h.handleIssueCommand(ctx, i)

	case "issues":
		h.handleIssuesCommand(ctx, i)
	case "issue-status":
		h.handleIssueStatusCommand(ctx, i)
	case "register":
		h.handleRegisterCommand(ctx, i)
	case "help":
		h.handleHelpCommand(ctx, i)
	default:
		h.logger.Warn("Unknown slash command", zap.String("command", commandName))
		h.respondToInteraction(ctx, i, "Unknown command", true)
	}
}

// handleIssueCommand handles the /issue slash command
func (h *Handler) handleIssueCommand(ctx context.Context, i *discordgo.InteractionCreate) {
	modal := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "issue_modal",
			Title:    "Create New Issue",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "title",
							Label:       "Issue Title",
							Style:       discordgo.TextInputShort,
							Placeholder: "e.g. Cannot login",
							Required:    true,
							MaxLength:   255,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "description",
							Label:       "Description",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "Describe the issue in detail...",
							Required:    true,
							MaxLength:   2000,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "image_url",
							Label:       "Image URL (Optional)",
							Style:       discordgo.TextInputShort,
							Placeholder: "https://example.com/image.png",
							Required:    false,
							MaxLength:   500,
						},
					},
				},
			},
		},
	}

	if err := h.session.InteractionRespond(i.Interaction, modal); err != nil {
		h.logger.Error("Failed to respond with modal", zap.Error(err))
	}
}

// handleIssuesCommand handles the /issues slash command
func (h *Handler) handleIssuesCommand(ctx context.Context, i *discordgo.InteractionCreate) {
	h.logger.Info("Handling issues command",
		zap.String("user_id", i.Member.User.ID),
		zap.String("channel_id", i.ChannelID),
	)

	// Get all issues for this channel
	issues, err := h.issueService.ListIssuesByChannel(ctx, i.ChannelID)
	if err != nil {
		h.logger.Error("Failed to get issues for channel",
			zap.Error(err),
			zap.String("channel_id", i.ChannelID),
		)
		h.respondToInteraction(ctx, i, "❌ Failed to retrieve issues. Please try again.", true)
		return
	}

	if len(issues) == 0 {
		h.respondToInteraction(ctx, i, "📋 No issues found in this channel.", false)
		return
	}

	// Build the response message
	var content strings.Builder
	content.WriteString(fmt.Sprintf("📋 **Issues in this channel (%d total):**\n\n", len(issues)))

	openCount := 0
	closedCount := 0

	for i, issue := range issues {
		if i >= 10 { // Limit to first 10 issues to avoid message length limits
			content.WriteString(fmt.Sprintf("*... and %d more issues*\n", len(issues)-10))
			break
		}

		// Status emoji
		statusEmoji := "🟢"
		if issue.Status == domain.StatusClosed {
			statusEmoji = "🔴"
			closedCount++
		} else {
			openCount++
		}

		// Priority emoji
		priorityEmoji := "🟡"
		switch issue.Priority {
		case domain.PriorityLow:
			priorityEmoji = "🟢"
		case domain.PriorityHigh:
			priorityEmoji = "🔴"
		}

		// Format creation time
		createdTime := issue.CreatedAt.Format("Jan 2, 2006")

		content.WriteString(fmt.Sprintf("%s %s **%s**\n", statusEmoji, priorityEmoji, issue.Title))
		content.WriteString(fmt.Sprintf("   🆔 `%s` | 📅 %s | 👤 <@%s>\n",
			issue.ID.String()[:8], createdTime, issue.ReporterID))

		if issue.ThreadID != "" {
			content.WriteString(fmt.Sprintf("   💬 <#%s>\n", issue.ThreadID))
		}
		content.WriteString("\n")
	}

	// Add summary
	content.WriteString(fmt.Sprintf("📊 **Summary:** %d Open, %d Closed", openCount, closedCount))

	h.respondToInteraction(ctx, i, content.String(), false)
}

// handleIssueStatusCommand handles the /issue-status slash command
func (h *Handler) handleIssueStatusCommand(ctx context.Context, i *discordgo.InteractionCreate) {
	h.logger.Info("Handling issue-status command",
		zap.String("user_id", i.Member.User.ID),
		zap.String("channel_id", i.ChannelID),
	)

	// Get the issue ID from the command options
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		h.respondToInteraction(ctx, i, "❌ Please provide an issue ID.", true)
		return
	}

	issueIDStr := options[0].StringValue()

	// Try to parse as UUID (full ID)
	issueID, err := uuid.Parse(issueIDStr)
	if err != nil {
		// If not a valid UUID, try to find by partial ID
		issues, err := h.issueService.ListIssuesByChannel(ctx, i.ChannelID)
		if err != nil {
			h.logger.Error("Failed to get issues for partial ID search",
				zap.Error(err),
				zap.String("channel_id", i.ChannelID),
			)
			h.respondToInteraction(ctx, i, "❌ Failed to search for issues. Please try again.", true)
			return
		}

		// Look for issue with matching partial ID
		var foundIssue *domain.Issue
		for _, issue := range issues {
			if strings.HasPrefix(issue.ID.String(), issueIDStr) {
				foundIssue = issue
				break
			}
		}

		if foundIssue == nil {
			h.respondToInteraction(ctx, i, fmt.Sprintf("❌ No issue found with ID: `%s`", issueIDStr), true)
			return
		}
		issueID = foundIssue.ID
	}

	// Get the issue details
	issue, err := h.issueService.GetIssue(ctx, issueID)
	if err != nil {
		if err == domain.ErrIssueNotFound {
			h.respondToInteraction(ctx, i, fmt.Sprintf("❌ Issue not found: `%s`", issueIDStr), true)
			return
		}

		h.logger.Error("Failed to get issue details",
			zap.Error(err),
			zap.String("issue_id", issueID.String()),
		)
		h.respondToInteraction(ctx, i, "❌ Failed to retrieve issue details. Please try again.", true)
		return
	}

	// Build the detailed response
	var content strings.Builder

	// Status emoji
	statusEmoji := "🟢 **OPEN**"
	if issue.Status == domain.StatusClosed {
		statusEmoji = "🔴 **CLOSED**"
	}

	// Priority emoji
	priorityEmoji := "🟡"
	priorityText := "Medium"
	switch issue.Priority {
	case domain.PriorityLow:
		priorityEmoji = "🟢"
		priorityText = "Low"
	case domain.PriorityHigh:
		priorityEmoji = "🔴"
		priorityText = "High"
	}

	content.WriteString(fmt.Sprintf("🎫 **Issue Details**\n\n"))
	content.WriteString(fmt.Sprintf("**Title:** %s\n", issue.Title))
	content.WriteString(fmt.Sprintf("**ID:** `%s`\n", issue.ID.String()))
	content.WriteString(fmt.Sprintf("**Status:** %s\n", statusEmoji))
	content.WriteString(fmt.Sprintf("**Priority:** %s %s\n", priorityEmoji, priorityText))
	content.WriteString(fmt.Sprintf("**Reporter:** <@%s>\n", issue.ReporterID))
	content.WriteString(fmt.Sprintf("**Created:** %s\n", issue.CreatedAt.Format("January 2, 2006 at 3:04 PM")))

	if issue.Status == domain.StatusClosed && issue.ClosedAt != nil {
		content.WriteString(fmt.Sprintf("**Closed:** %s\n", issue.ClosedAt.Format("January 2, 2006 at 3:04 PM")))
	}

	if issue.ThreadID != "" {
		content.WriteString(fmt.Sprintf("**Discussion:** <#%s>\n", issue.ThreadID))
	}

	content.WriteString(fmt.Sprintf("\n**Description:**\n%s", issue.Description))

	// Create embed for image if present
	var embeds []*discordgo.MessageEmbed
	if issue.ImageURL != "" {
		embeds = []*discordgo.MessageEmbed{
			{
				Title: "Issue Screenshot",
				Image: &discordgo.MessageEmbedImage{URL: issue.ImageURL},
				Color: 0x3498db,
			},
		}
	}

	// Send response with embed if there's an image
	if len(embeds) > 0 {
		if err := h.session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content.String(),
				Embeds:  embeds,
			},
		}); err != nil {
			h.logger.Error("Failed to respond to interaction with embed", zap.Error(err))
		}
	} else {
		h.respondToInteraction(ctx, i, content.String(), false)
	}
}

// handleHelpCommand handles the /help slash command
func (h *Handler) handleHelpCommand(ctx context.Context, i *discordgo.InteractionCreate) {
	h.logger.Info("Handling help command",
		zap.String("user_id", i.Member.User.ID),
		zap.String("channel_id", i.ChannelID),
	)

	helpContent := `🤖 **Fix Track Bot - Help**

**Available Commands:**

🎫 ` + "`/issue`" + ` - Create a new issue or bug report
   Opens a form to submit a new issue with title, description, and optional screenshot

📋 ` + "`/issues`" + ` - List all issues in this channel
   Shows a summary of all issues in the current channel with status and priority

🔍 ` + "`/issue-status <id>`" + ` - Check the status of a specific issue
   Shows detailed information about an issue (use full UUID or first 8 characters)

📝 ` + "`/register`" + ` - Register this channel for issue tracking
   Register the channel with customer and project information (required before creating issues)

❓ ` + "`/help`" + ` - Show this help message

**Features:**

• **Issue Tracking** - Create and track issues with unique IDs
• **Thread Discussions** - Each issue gets its own discussion thread
• **Priority Levels** - Set priority as Low 🟢, Medium 🟡, or High 🔴
• **Status Management** - Issues can be Open 🟢 or Closed 🔴
• **Screenshots** - Attach images to issues via URL

**How to Use:**

1. Use ` + "`/issue`" + ` to create a new issue
2. Fill out the form with title, description, and optional image URL
3. The bot creates a thread for discussion
4. Set priority using the dropdown in the thread
5. Close issues using the "🔒 Close Issue" button

**Tips:**

• Use short, descriptive titles for better organization
• Include steps to reproduce in the description
• Use the thread for follow-up discussion and updates
• Close issues when they're resolved

Need more help? Contact your server administrators.`

	h.respondToInteraction(ctx, i, helpContent, false)
}

// handleRegisterCommand handles the /register slash command
func (h *Handler) handleRegisterCommand(ctx context.Context, i *discordgo.InteractionCreate) {
	h.logger.Info("Handling register command",
		zap.String("user_id", i.Member.User.ID),
		zap.String("channel_id", i.ChannelID),
		zap.String("guild_id", i.GuildID),
	)

	// Check if channel is already registered
	isRegistered, err := h.channelService.IsChannelRegistered(ctx, i.ChannelID)
	if err != nil {
		h.logger.Error("Failed to check channel registration status",
			zap.Error(err),
			zap.String("channel_id", i.ChannelID),
		)
		h.respondToInteraction(ctx, i, "❌ Failed to check channel registration status. Please try again.", true)
		return
	}

	if isRegistered {
		// Get existing registration details
		channel, err := h.channelService.GetChannelRegistration(ctx, i.ChannelID)
		if err != nil {
			h.logger.Error("Failed to get existing channel registration",
				zap.Error(err),
				zap.String("channel_id", i.ChannelID),
			)
			h.respondToInteraction(ctx, i, "❌ Failed to get existing registration details.", true)
			return
		}

		response := fmt.Sprintf("⚠️ **Channel Already Registered**\n\n"+
			"This channel is already registered for issue tracking:\n\n"+
			"**Customer:** %s\n"+
			"**Project:** %s\n"+
			"**Registered by:** <@%s>\n"+
			"**Registration Date:** %s\n\n"+
			"To update the registration, please contact an administrator.",
			channel.Project.Customer.Name,
			channel.Project.Name,
			channel.RegisteredByUser.DiscordID,
			channel.CreatedAt.Format("January 2, 2006"),
		)

		h.respondToInteraction(ctx, i, response, true)
		return
	}

	// Show registration modal
	modal := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "register_modal",
			Title:    "Register Channel for Issue Tracking",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "customer_name",
							Label:       "Customer/Organization Name",
							Style:       discordgo.TextInputShort,
							Placeholder: "e.g. Acme Corporation",
							Required:    true,
							MaxLength:   255,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "customer_email",
							Label:       "Customer Contact Email (Optional)",
							Style:       discordgo.TextInputShort,
							Placeholder: "e.g. contact@acme.com",
							Required:    false,
							MaxLength:   255,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "project_name",
							Label:       "Project Name",
							Style:       discordgo.TextInputShort,
							Placeholder: "e.g. E-commerce Platform",
							Required:    true,
							MaxLength:   255,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "project_description",
							Label:       "Project Description (Optional)",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "Brief description of the project...",
							Required:    false,
							MaxLength:   500,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "user_name",
							Label:       "Your Full Name (Optional)",
							Style:       discordgo.TextInputShort,
							Placeholder: "e.g. John Doe",
							Required:    false,
							MaxLength:   255,
						},
					},
				},
			},
		},
	}

	if err := h.session.InteractionRespond(i.Interaction, modal); err != nil {
		h.logger.Error("Failed to respond with register modal", zap.Error(err))
	}
}

// handleModalSubmit handles modal submission interactions
func (h *Handler) handleModalSubmit(ctx context.Context, i *discordgo.InteractionCreate) {
	modalID := i.ModalSubmitData().CustomID

	h.logger.Info("Handling modal submit",
		zap.String("modal_id", modalID),
		zap.String("user_id", i.Member.User.ID),
		zap.String("channel_id", i.ChannelID),
	)

	switch modalID {
	case "issue_modal":
		h.handleIssueModalSubmit(ctx, i)
	case "register_modal":
		h.handleRegisterModalSubmit(ctx, i)
	default:
		h.logger.Warn("Unknown modal ID", zap.String("modal_id", modalID))
		h.respondToInteraction(ctx, i, "Unknown modal", true)
	}
}

// handleRegisterModalSubmit handles the channel registration modal submission
func (h *Handler) handleRegisterModalSubmit(ctx context.Context, i *discordgo.InteractionCreate) {
	// Extract modal data
	components := i.ModalSubmitData().Components
	if len(components) < 3 {
		h.logger.Error("Invalid register modal components")
		h.respondToInteraction(ctx, i, "Invalid form data", true)
		return
	}

	customerName := components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	customerEmail := components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	projectName := components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	// Optional fields
	var projectDescription, userName string
	if len(components) > 3 {
		projectDescription = components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	}
	if len(components) > 4 {
		userName = components[4].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	}

	h.logger.Info("Processing channel registration",
		zap.String("customer_name", customerName),
		zap.String("customer_email", customerEmail),
		zap.String("project_name", projectName),
		zap.String("project_description", projectDescription),
		zap.String("user_name", userName),
		zap.String("user_id", i.Member.User.ID),
		zap.String("channel_id", i.ChannelID),
	)

	// Respond immediately to avoid timeout
	if err := h.session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🔄 Registering channel...",
		},
	}); err != nil {
		h.logger.Error("Failed to respond to registration interaction", zap.Error(err))
		return
	}

	// Register the channel through service
	channel, err := h.channelService.RegisterChannel(ctx, i.ChannelID, customerName, customerEmail, projectName, projectDescription, i.Member.User.ID, userName, i.GuildID)
	if err != nil {
		h.logger.Error("Failed to register channel", zap.Error(err))

		var errorMessage string
		switch err {
		case domain.ErrChannelAlreadyRegistered:
			errorMessage = "❌ This channel is already registered for issue tracking."
		case domain.ErrEmptyCustomerName:
			errorMessage = "❌ Customer name cannot be empty."
		case domain.ErrEmptyProjectName:
			errorMessage = "❌ Project name cannot be empty."
		default:
			errorMessage = "❌ Failed to register channel. Please try again."
		}

		h.editInteractionResponse(ctx, i, errorMessage)
		return
	}

	// Create success response
	successContent := fmt.Sprintf("✅ **Channel Registration Successful!**\n\n"+
		"This channel has been registered for issue tracking:\n\n"+
		"🏢 **Customer:** %s\n"+
		"📧 **Contact:** %s\n"+
		"📋 **Project:** %s\n"+
		"📝 **Description:** %s\n"+
		"📅 **Registered:** %s\n"+
		"👤 **Registered by:** %s (<@%s>)\n\n"+
		"You can now use the `/issue` command to create and track issues in this channel.\n\n"+
		"**Available Commands:**\n"+
		"• `/issue` - Create a new issue\n"+
		"• `/issues` - List all issues\n"+
		"• `/issue-status <id>` - Check issue status\n"+
		"• `/help` - Show help information",
		channel.Project.Customer.Name,
		getDisplayValue(channel.Project.Customer.ContactEmail, "Not provided"),
		channel.Project.Name,
		getDisplayValue(channel.Project.Description, "No description provided"),
		channel.CreatedAt.Format("January 2, 2006 at 3:04 PM"),
		getDisplayValue(channel.RegisteredByUser.Name, "Discord User"),
		channel.RegisteredByUser.DiscordID,
	)

	// Update the original response
	h.editInteractionResponse(ctx, i, successContent)

	// Log successful registration
	h.logger.Info("Channel registration completed successfully",
		zap.String("registration_id", channel.ID.String()),
		zap.String("channel_id", i.ChannelID),
		zap.String("customer_name", customerName),
		zap.String("project_name", projectName),
	)
}

// handleIssueModalSubmit handles the issue creation modal submission
func (h *Handler) handleIssueModalSubmit(ctx context.Context, i *discordgo.InteractionCreate) {
	// Extract modal data
	components := i.ModalSubmitData().Components
	if len(components) < 2 {
		h.logger.Error("Invalid modal components")
		h.respondToInteraction(ctx, i, "Invalid form data", true)
		return
	}

	title := components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	description := components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	imageURL := ""
	if len(components) > 2 {
		imageURL = components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	}

	h.logger.Info("Creating issue from modal",
		zap.String("title", title),
		zap.String("user_id", i.Member.User.ID),
	)

	// Respond immediately to avoid timeout
	if err := h.session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🔄 Creating issue...",
		},
	}); err != nil {
		h.logger.Error("Failed to respond to interaction", zap.Error(err))
		return
	}

	// Create issue through service
	issue, err := h.issueService.CreateIssue(ctx, title, description, imageURL, i.Member.User.ID, i.ChannelID)
	if err != nil {
		h.logger.Error("Failed to create issue", zap.Error(err))
		h.editInteractionResponse(ctx, i, "❌ Failed to create issue. Please try again.")
		return
	}

	// Generate issue ID for display
	displayID := domain.GenerateIssueID()

	// Create the main issue message
	mainContent := fmt.Sprintf("🎫 **New Issue: %s**\n\n**Issue ID:** %s\n**Description:** %s\n**Reporter:** <@%s>",
		title, displayID, description, i.Member.User.ID)

	// Create embeds for image if provided
	var embeds []*discordgo.MessageEmbed
	if imageURL != "" {
		embeds = []*discordgo.MessageEmbed{
			{
				Title: "Issue Screenshot",
				Image: &discordgo.MessageEmbedImage{URL: imageURL},
				Color: 0x3498db,
			},
		}
	}

	// Create message with close button
	message, err := h.session.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
		Content: mainContent,
		Embeds:  embeds,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "🔒 Close Issue",
						Style:    discordgo.DangerButton,
						CustomID: fmt.Sprintf("close_issue_%s", issue.ID.String()),
					},
				},
			},
		},
	})
	if err != nil {
		h.logger.Error("Failed to send issue message", zap.Error(err))
		h.editInteractionResponse(ctx, i, "❌ Failed to post issue message.")
		return
	}

	// Create thread for discussion
	thread, err := h.session.MessageThreadStart(i.ChannelID, message.ID, fmt.Sprintf("Issue: %s", title), 60)
	if err != nil {
		h.logger.Error("Failed to create thread", zap.Error(err))
		// Continue without thread
	} else {
		// Update issue with thread and message info
		if err := h.issueService.SetThreadInfo(ctx, issue.ID, thread.ID, message.ID); err != nil {
			h.logger.Error("Failed to update issue thread info", zap.Error(err))
		}

		// Update button with thread ID
		components := []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "🔒 Close Issue",
						Style:    discordgo.DangerButton,
						CustomID: fmt.Sprintf("close_issue_%s_%s", issue.ID.String(), thread.ID),
					},
				},
			},
		}

		if _, err := h.session.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    i.ChannelID,
			ID:         message.ID,
			Content:    &mainContent,
			Embeds:     &embeds,
			Components: &components,
		}); err != nil {
			h.logger.Error("Failed to update message with thread ID", zap.Error(err))
		}

		// Add priority selection in thread
		h.sendPrioritySelector(ctx, thread.ID, issue.ID.String())

		// Send welcome message in thread
		h.sendMessage(ctx, thread.ID, fmt.Sprintf("💬 Discussion thread for Issue **%s**\n\nFeel free to add comments, updates, or additional information here.", displayID))
	}

	// Update the original response
	h.editInteractionResponse(ctx, i, fmt.Sprintf("✅ Issue **%s** created successfully!", displayID))
}

// handleMessageComponent handles button clicks and select menu interactions
func (h *Handler) handleMessageComponent(ctx context.Context, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID

	h.logger.Info("Handling message component",
		zap.String("custom_id", customID),
		zap.String("user_id", i.Member.User.ID),
	)

	// Handle close issue button
	if strings.HasPrefix(customID, "close_issue_") {
		h.handleCloseIssueButton(ctx, i)
		return
	}

	// Handle priority selection
	if strings.HasPrefix(customID, "issue_priority_") {
		h.handlePrioritySelection(ctx, i)
		return
	}

	h.logger.Warn("Unknown message component", zap.String("custom_id", customID))
	h.respondToInteraction(ctx, i, "Unknown action", true)
}

// handleCloseIssueButton handles the close issue button click
func (h *Handler) handleCloseIssueButton(ctx context.Context, i *discordgo.InteractionCreate) {
	parts := strings.Split(i.MessageComponentData().CustomID, "_")
	if len(parts) < 3 {
		h.logger.Error("Invalid close button custom ID")
		h.respondToInteraction(ctx, i, "Invalid button action", true)
		return
	}

	issueIDStr := parts[2]
	issueID, err := uuid.Parse(issueIDStr)
	if err != nil {
		h.logger.Error("Invalid issue ID in button", zap.Error(err))
		h.respondToInteraction(ctx, i, "Invalid issue ID", true)
		return
	}

	// Close the issue through service
	if err := h.issueService.CloseIssue(ctx, issueID); err != nil {
		h.logger.Error("Failed to close issue", zap.Error(err))
		h.respondToInteraction(ctx, i, "❌ Failed to close issue", true)
		return
	}

	// Respond to user
	h.respondToInteraction(ctx, i, "🔒 Closing issue...", true)

	// Update the original message
	originalMessage := i.Message
	closedContent := fmt.Sprintf("🔒 **[CLOSED]** %s", originalMessage.Content)

	if _, err := h.session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    i.ChannelID,
		ID:         originalMessage.ID,
		Content:    &closedContent,
		Embeds:     &originalMessage.Embeds,
		Components: &[]discordgo.MessageComponent{}, // Remove all components
	}); err != nil {
		h.logger.Error("Failed to update message", zap.Error(err))
	}

	// If there's a thread, close it
	if len(parts) > 3 {
		threadID := parts[3]
		h.sendMessage(ctx, threadID, "🔒 **This issue has been closed.**\n\nThis thread will be archived.")

		if _, err := h.session.ChannelEditComplex(threadID, &discordgo.ChannelEdit{
			Archived: &[]bool{true}[0],
			Locked:   &[]bool{true}[0],
		}); err != nil {
			h.logger.Error("Failed to archive thread", zap.Error(err))
		}
	}
}

// handlePrioritySelection handles priority selection from select menu
func (h *Handler) handlePrioritySelection(ctx context.Context, i *discordgo.InteractionCreate) {
	if len(i.MessageComponentData().Values) == 0 {
		h.respondToInteraction(ctx, i, "No priority selected", true)
		return
	}

	priorityStr := i.MessageComponentData().Values[0]
	priority := domain.Priority(priorityStr)

	// Extract issue ID from custom ID (format: "issue_priority_<uuid>")
	customID := i.MessageComponentData().CustomID
	parts := strings.Split(customID, "_")
	if len(parts) < 3 {
		h.logger.Error("Invalid priority selector custom ID", zap.String("custom_id", customID))
		h.respondToInteraction(ctx, i, "❌ Invalid priority selector", true)
		return
	}

	issueIDStr := parts[2]
	issueID, err := uuid.Parse(issueIDStr)
	if err != nil {
		h.logger.Error("Invalid issue ID in priority selector",
			zap.Error(err),
			zap.String("issue_id", issueIDStr),
		)
		h.respondToInteraction(ctx, i, "❌ Invalid issue ID", true)
		return
	}

	// Update the issue priority
	if err := h.issueService.UpdateIssuePriority(ctx, issueID, priority); err != nil {
		h.logger.Error("Failed to update issue priority",
			zap.Error(err),
			zap.String("issue_id", issueID.String()),
			zap.String("priority", priorityStr),
		)
		h.respondToInteraction(ctx, i, "❌ Failed to update priority. Please try again.", true)
		return
	}

	// Disable the select menu after selection
	disabledSelectMenu := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    customID,
				Placeholder: fmt.Sprintf("Priority set to %s", strings.Title(priorityStr)),
				Disabled:    true,
				Options: []discordgo.SelectMenuOption{
					{Label: "Low", Value: "low", Emoji: &discordgo.ComponentEmoji{Name: "🟢"}},
					{Label: "Medium", Value: "medium", Emoji: &discordgo.ComponentEmoji{Name: "🟡"}},
					{Label: "High", Value: "high", Emoji: &discordgo.ComponentEmoji{Name: "🔴"}},
				},
			},
		},
	}

	// Update the message to disable the selector
	newContent := fmt.Sprintf("📊 **Priority set to %s %s**",
		func() string {
			switch priority {
			case domain.PriorityLow:
				return "🟢"
			case domain.PriorityHigh:
				return "🔴"
			default:
				return "🟡"
			}
		}(), strings.Title(priorityStr))

	if _, err := h.session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    i.ChannelID,
		ID:         i.Message.ID,
		Content:    &newContent,
		Components: &[]discordgo.MessageComponent{disabledSelectMenu},
	}); err != nil {
		h.logger.Error("Failed to update priority selector message", zap.Error(err))
	}

	h.respondToInteraction(ctx, i, fmt.Sprintf("✅ Priority set to **%s**", strings.Title(priorityStr)), true)
}

// Helper methods

func (h *Handler) sendMessage(ctx context.Context, channelID, content string) {
	if _, err := h.session.ChannelMessageSend(channelID, content); err != nil {
		h.logger.Error("Failed to send message",
			zap.Error(err),
			zap.String("channel_id", channelID),
		)
	}
}

func (h *Handler) respondToInteraction(ctx context.Context, i *discordgo.InteractionCreate, content string, ephemeral bool) {
	flags := discordgo.MessageFlags(0)
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	if err := h.session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   flags,
		},
	}); err != nil {
		h.logger.Error("Failed to respond to interaction", zap.Error(err))
	}
}

func (h *Handler) editInteractionResponse(ctx context.Context, i *discordgo.InteractionCreate, content string) {
	if _, err := h.session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	}); err != nil {
		h.logger.Error("Failed to edit interaction response", zap.Error(err))
	}
}

func (h *Handler) sendPrioritySelector(ctx context.Context, channelID, issueID string) {
	selectMenu := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    fmt.Sprintf("issue_priority_%s", issueID),
				Placeholder: "Select priority",
				Options: []discordgo.SelectMenuOption{
					{Label: "Low", Value: "low", Emoji: &discordgo.ComponentEmoji{Name: "🟢"}},
					{Label: "Medium", Value: "medium", Emoji: &discordgo.ComponentEmoji{Name: "🟡"}},
					{Label: "High", Value: "high", Emoji: &discordgo.ComponentEmoji{Name: "🔴"}},
				},
			},
		},
	}

	if _, err := h.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content:    "📊 **Set Issue Priority:**",
		Components: []discordgo.MessageComponent{selectMenu},
	}); err != nil {
		h.logger.Error("Failed to send priority selector", zap.Error(err))
	}
}

// getDisplayValue returns the value if not empty, otherwise returns the default value
func getDisplayValue(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}
