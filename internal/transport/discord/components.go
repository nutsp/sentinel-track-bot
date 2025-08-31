package discord

import (
	"fix-track-bot/internal/domain"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// CreateIssueCard creates a Discord embed with action buttons for an issue
func CreateIssueCard(issue *domain.Issue) (*discordgo.MessageEmbed, []discordgo.MessageComponent) {

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🐛 %s", issue.Title),
		Description: issue.Description,
		Color:       getStatusColorInt(issue.Status),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Status",
				Value:  fmt.Sprintf("%s %s", getStatusEmoji(issue.Status), issue.GetStatusDisplayName()),
				Inline: true,
			},
			{
				Name:   "Priority",
				Value:  fmt.Sprintf("%s %s", getPriorityEmoji(issue.Priority), string(issue.Priority)),
				Inline: true,
			},
			{
				Name:   "Reporter",
				Value:  fmt.Sprintf("<@%s>", issue.Reporter.DiscordID),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Issue ID: %s", issue.ID.String()),
		},
		Timestamp: issue.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Add assignees if any
	if len(issue.Assignees) > 0 {
		assigneeText := ""
		for _, assignee := range issue.Assignees {
			roleEmoji := getRoleEmoji(assignee.Role)
			assigneeText += fmt.Sprintf("%s %s <@%s>\n", roleEmoji, assignee.Role.GetDisplayName(), assignee.User.DiscordID)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Assignees",
			Value:  assigneeText,
			Inline: false,
		})
	}

	// Create action buttons based on current status
	buttons := createActionButtons(issue)

	return embed, buttons
}

// createActionButtons creates context-aware action buttons based on issue status
func createActionButtons(issue *domain.Issue) []discordgo.MessageComponent {
	// Get possible next statuses
	nextStatuses := issue.GetNextPossibleStatuses()

	// Create buttons for each possible action
	var buttons []discordgo.MessageComponent

	for _, status := range nextStatuses {
		button := createStatusButton(issue.ID.String(), status)
		if button != nil {
			buttons = append(buttons, button)
		}
	}

	// Add utility buttons
	// buttons = append(buttons,
	// 	&discordgo.Button{
	// 		Label:    "📋 Details",
	// 		Style:    discordgo.SecondaryButton,
	// 		CustomID: fmt.Sprintf("issue_details_%s", issue.ID.String()),
	// 		Emoji: &discordgo.ComponentEmoji{
	// 			Name: "📋",
	// 		},
	// 	},
	// 	&discordgo.Button{
	// 		Label:    "📊 History",
	// 		Style:    discordgo.SecondaryButton,
	// 		CustomID: fmt.Sprintf("issue_history_%s", issue.ID.String()),
	// 		Emoji: &discordgo.ComponentEmoji{
	// 			Name: "📊",
	// 		},
	// 	},
	// )

	// Group buttons into rows (max 5 buttons per row)
	var rows []discordgo.MessageComponent
	for i := 0; i < len(buttons); i += 5 {
		end := i + 5
		if end > len(buttons) {
			end = len(buttons)
		}

		row := discordgo.ActionsRow{
			Components: buttons[i:end],
		}
		rows = append(rows, row)
	}

	return rows
}

// createStatusButton creates a button for a specific status transition
func createStatusButton(issueID string, status domain.Status) discordgo.MessageComponent {
	switch status {
	case domain.StatusDraft:
		return &discordgo.Button{
			Label:    "🔵 Open",
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("open_issue_%s", issueID),
			Emoji: &discordgo.ComponentEmoji{
				Name: "🔵",
			},
		}
	case domain.StatusOpen:
		return &discordgo.Button{
			Label:    "🔵 Open",
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("open_issue_%s", issueID),
			Emoji: &discordgo.ComponentEmoji{
				Name: "🔵",
			},
		}
	case domain.StatusInProgress:
		return &discordgo.Button{
			Label:    "🟡 Start Work",
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("start_work_%s", issueID),
			Emoji: &discordgo.ComponentEmoji{
				Name: "🟡",
			},
		}
	case domain.StatusResolved:
		return &discordgo.Button{
			Label:    "🟢 Resolve",
			Style:    discordgo.SuccessButton,
			CustomID: fmt.Sprintf("resolve_issue_%s", issueID),
			Emoji: &discordgo.ComponentEmoji{
				Name: "🟢",
			},
		}
	// case domain.StatusAssignedQA:
	// 	return &discordgo.Button{
	// 		Label:    "🔷 Assign QA",
	// 		Style:    discordgo.PrimaryButton,
	// 		CustomID: fmt.Sprintf("assign_qa_%s", issueID),
	// 		Emoji: &discordgo.ComponentEmoji{
	// 			Name: "🔷",
	// 		},
	// 	}
	// case domain.StatusVerified:
	// 	return &discordgo.Button{
	// 		Label:    "✅ Verify",
	// 		Style:    discordgo.SuccessButton,
	// 		CustomID: fmt.Sprintf("verify_issue_%s", issueID),
	// 		Emoji: &discordgo.ComponentEmoji{
	// 			Name: "✅",
	// 		},
	// 	}
	case domain.StatusClosed:
		return &discordgo.Button{
			Label:    "🟣 Close",
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("close_issue_%s", issueID),
			Emoji: &discordgo.ComponentEmoji{
				Name: "🟣",
			},
		}
		// case domain.StatusRejected:
		// 	return &discordgo.Button{
		// 		Label:    "🔴 Reject",
		// 		Style:    discordgo.DangerButton,
		// 		CustomID: fmt.Sprintf("reject_issue_%s", issueID),
		// 		Emoji: &discordgo.ComponentEmoji{
		// 			Name: "🔴",
		// 		},
		// 	}
		// case domain.StatusReopened:
		// 	return &discordgo.Button{
		// 		Label:    "🟠 Reopen",
		// 		Style:    discordgo.SecondaryButton,
		// 		CustomID: fmt.Sprintf("reopen_issue_%s", issueID),
		// 		Emoji: &discordgo.ComponentEmoji{
		// 			Name: "🟠",
		// 		},
		// 	}
		// case domain.StatusOpen:
		// 	return &discordgo.Button{
		// 		Label:    "🔵 Back to Open",
		// 		Style:    discordgo.SecondaryButton,
		// 		CustomID: fmt.Sprintf("back_to_open_%s", issueID),
		// 		Emoji: &discordgo.ComponentEmoji{
		// 			Name: "🔵",
		// 		},
		// 	}
	}
	return nil
}

// CreateUserSelectMenu creates a select menu for choosing users
func CreateUserSelectMenu(customID string, placeholder string, minValues, maxValues int) discordgo.MessageComponent {
	return discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    customID,
				Placeholder: placeholder,
				MinValues:   &minValues,
				MaxValues:   maxValues,
				MenuType:    discordgo.UserSelectMenu,
			},
		},
	}
}

// CreateRoleSelectMenu creates a select menu for choosing roles
func CreateRoleSelectMenu(customID string) discordgo.MessageComponent {
	return discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    customID,
				Placeholder: "Choose role...",
				Options: []discordgo.SelectMenuOption{
					{
						Label:       "👨‍💻 Developer",
						Value:       "dev",
						Description: "Software developer",
						Emoji: &discordgo.ComponentEmoji{
							Name: "👨‍💻",
						},
					},
					{
						Label:       "🧪 QA Tester",
						Value:       "qa",
						Description: "Quality assurance tester",
						Emoji: &discordgo.ComponentEmoji{
							Name: "🧪",
						},
					},
					{
						Label:       "👀 Reviewer",
						Value:       "reviewer",
						Description: "Code reviewer",
						Emoji: &discordgo.ComponentEmoji{
							Name: "👀",
						},
					},
					{
						Label:       "👤 Other",
						Value:       "other",
						Description: "Other role",
						Emoji: &discordgo.ComponentEmoji{
							Name: "👤",
						},
					},
				},
			},
		},
	}
}

// CreateStatusFilterMenu creates a select menu for filtering by status
func CreateStatusFilterMenu(customID string) discordgo.MessageComponent {
	return discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    customID,
				Placeholder: "Filter by status...",
				Options: []discordgo.SelectMenuOption{
					{
						Label: "🔵 Open",
						Value: "open",
						Emoji: &discordgo.ComponentEmoji{Name: "🔵"},
					},
					{
						Label: "🔷 In Progress",
						Value: "in_progress",
						Emoji: &discordgo.ComponentEmoji{Name: "🔷"},
					},
					{
						Label: "🟡 In Progress",
						Value: "in_progress",
						Emoji: &discordgo.ComponentEmoji{Name: "🟡"},
					},
					{
						Label: "🟢 Resolved",
						Value: "resolved",
						Emoji: &discordgo.ComponentEmoji{Name: "🟢"},
					},
					{
						Label: "🔷 Verified",
						Value: "verified",
						Emoji: &discordgo.ComponentEmoji{Name: "🔷"},
					},
					{
						Label: "✅ Verified",
						Value: "verified",
						Emoji: &discordgo.ComponentEmoji{Name: "✅"},
					},
					{
						Label: "🟣 Closed",
						Value: "closed",
						Emoji: &discordgo.ComponentEmoji{Name: "🟣"},
					},
					{
						Label: "🔴 Rejected",
						Value: "rejected",
						Emoji: &discordgo.ComponentEmoji{Name: "🔴"},
					},
					{
						Label: "🟠 Reopened",
						Value: "reopened",
						Emoji: &discordgo.ComponentEmoji{Name: "🟠"},
					},
				},
			},
		},
	}
}

// CreateConfirmationButtons creates Yes/No confirmation buttons
func CreateConfirmationButtons(actionID string, issueID string) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					Label:    "✅ Confirm",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("confirm_%s_%s", actionID, issueID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "✅",
					},
				},
				&discordgo.Button{
					Label:    "❌ Cancel",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("cancel_%s_%s", actionID, issueID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "❌",
					},
				},
			},
		},
	}
}

// Helper functions for emojis and colors
func getStatusEmoji(status domain.Status) string {
	switch status {
	case domain.StatusDraft:
		return "⚪"
	case domain.StatusOpen:
		return "🔵"
	case domain.StatusInProgress:
		return "🔷"
	case domain.StatusResolved:
		return "🟢"
	case domain.StatusVerified:
		return "✅"
	case domain.StatusClosed:
		return "🟣"
	case domain.StatusRejected:
		return "🔴"
	case domain.StatusReopened:
		return "🟠"
	default:
		return "⚪"
	}
}

func getPriorityEmoji(priority domain.Priority) string {
	switch priority {
	case domain.PriorityLow:
		return "🟢"
	case domain.PriorityMedium:
		return "🟡"
	case domain.PriorityHigh:
		return "🔴"
	default:
		return "⚪"
	}
}

func getRoleEmoji(role domain.AssigneeRole) string {
	switch role {
	case domain.AssigneeRoleDev:
		return "👨‍💻"
	case domain.AssigneeRoleQA:
		return "🧪"
	case domain.AssigneeRoleReviewer:
		return "👀"
	case domain.AssigneeRoleOther:
		return "👤"
	default:
		return "👤"
	}
}

func getStatusColorInt(status domain.Status) int {
	switch status {
	case domain.StatusDraft:
		return 0x95a5a6 // Silver gray
	case domain.StatusOpen:
		return 0x3498db // Bright blue
	case domain.StatusInProgress:
		return 0x9b59b6 // Amethyst purple
	case domain.StatusResolved:
		return 0xf39c12 // Vivid orange
	case domain.StatusVerified:
		return 0x2ecc71 // Emerald green
	case domain.StatusClosed:
		return 0x34495e // Navy gray
	case domain.StatusRejected:
		return 0xe74c3c // Strong red
	case domain.StatusReopened:
		return 0xe67e22 // Carrot orange
	default:
		return 0x7f8c8d // Default gray
	}
}
