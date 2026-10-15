package llm

import (
	"fmt"
	"strings"
)

// NewsletterPrompt creates the prompt for newsletter generation
// This generates an informative newsletter highlighting contributions and discussions
func NewsletterPrompt(payload map[string]interface{}) string {
	var builder strings.Builder

	builder.WriteString(`You are writing a weekly newsletter for Fedora CoreOS.

This newsletter highlights contributions and activities. The tone should be calm, simple, and informative. Present facts clearly and directly without speculation. Be professional and appreciative of all contributors.

## Activity Summary
`)

	// Window information
	if start, ok := payload["window_start"].(string); ok {
		if end, ok := payload["window_end"].(string); ok {
			builder.WriteString(fmt.Sprintf("- Week of: %s to %s\n", start, end))
		}
	}

	// PR statistics
	if totalPRs, ok := payload["total_prs"].(int); ok {
		builder.WriteString(fmt.Sprintf("- Pull requests: %d total\n", totalPRs))
	}

	// Meeting statistics
	if totalMeetings, ok := payload["total_meetings"].(int); ok {
		builder.WriteString(fmt.Sprintf("- Community meetings: %d\n", totalMeetings))
	}

	builder.WriteString("\n## Highlights\n\n")

	// Top PRs
	if topPRs, ok := payload["top_prs"].([]map[string]interface{}); ok {
		builder.WriteString("### Notable Contributions\n\n")
		for i, pr := range topPRs {
			if i >= 5 {
				break
			}
			title, _ := pr["title"].(string)
			url, _ := pr["url"].(string)
			author, _ := pr["author"].(*string)

			authorStr := "unknown"
			if author != nil {
				authorStr = *author
			}

			builder.WriteString(fmt.Sprintf("- **%s** ([PR](%s)) by @%s\n", title, url, authorStr))
		}
		builder.WriteString("\n")
	}

	// Contributor recognition
	if contributors, ok := payload["human_contributors"].([]string); ok {
		if len(contributors) > 0 {
			builder.WriteString("### Thanks to Our Contributors\n\n")
			for _, contributor := range contributors {
				builder.WriteString(fmt.Sprintf("- @%s\n", contributor))
			}
			builder.WriteString("\n")
		}
	}

	// Meetings summary
	if meetings, ok := payload["meetings"].([]map[string]interface{}); ok {
		if len(meetings) > 0 {
			builder.WriteString("### Community Meetings\n\n")
			for _, meeting := range meetings {
				date, _ := meeting["date"].(string)
				url, _ := meeting["url"].(string)
				topics, _ := meeting["topics"].([]string)
				decisions, _ := meeting["decisions"].([]string)
				actions, _ := meeting["actions"].([]string)
				attendees, _ := meeting["attendees"].([]string)

				builder.WriteString(fmt.Sprintf("**Meeting on %s** ([View Meeting Log](%s))\n\n", date, url))

				if len(attendees) > 0 {
					builder.WriteString(fmt.Sprintf("Attendees: %d participants\n\n", len(attendees)))
				}

				if len(topics) > 0 {
					builder.WriteString("**Topics Discussed:**\n")
					for _, topic := range topics {
						builder.WriteString(fmt.Sprintf("- %s\n", topic))
					}
					builder.WriteString("\n")
				}

				if len(decisions) > 0 {
					builder.WriteString("**Key Decisions:**\n")
					for _, decision := range decisions {
						builder.WriteString(fmt.Sprintf("- %s\n", decision))
					}
					builder.WriteString("\n")
				}

				if len(actions) > 0 {
					builder.WriteString("**Action Items:**\n")
					for _, action := range actions {
						builder.WriteString(fmt.Sprintf("- %s\n", action))
					}
					builder.WriteString("\n")
				}
			}
		}
	}

	builder.WriteString("\n## Instructions\n\n")
	builder.WriteString(`Based on the data above, write a newsletter that:

1. Welcomes and appreciates contributors
2. Highlights key contributions and pull requests
3. Recognizes contributors
4. Summarizes discussions from meetings
5. Provides links to get involved
6. Maintains a welcoming, inclusive tone

Format the output in Markdown. Keep it brief, positive, and focused on value.
`)

	return builder.String()
}

// ExtractSummaryFromResponse extracts the newsletter content from LLM response
func ExtractSummaryFromResponse(response string) string {
	// For now, return the response as-is
	// In a real implementation, you might parse structured output
	return strings.TrimSpace(response)
}
