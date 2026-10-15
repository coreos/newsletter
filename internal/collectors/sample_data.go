package collectors

import (
	"time"

	"github.com/coreos/coreos-newsletter/internal/models"
)

// GetSampleMeeting returns a sample Fedora CoreOS meeting for demonstration/testing
// This is based on the actual October 14, 2026 meeting
func GetSampleMeeting() models.MeetingRecord {
	return models.MeetingRecord{
		Title: "Fedora CoreOS Meeting",
		Date:  time.Date(2026, 10, 14, 14, 0, 0, 0, time.UTC),
		URL:   "https://meetbot.fedoraproject.org/fedora-coreos/2026-10-14/fedora-coreos-meeting.html",
		Attendees: []string{
			"@angelcr (33 messages)",
			"@ydesouza (13 messages)",
			"@cadejacobson (5 messages)",
			"@rapneset (4 messages)",
			"@marmijo (4 messages)",
			"@hricky (2 messages)",
			"@bittin_ (1 message)",
			"@nimbinatus (1 message)",
			"@jeckersb (1 message)",
		},
		Topics: []string{
			"roll call",
			"Action items from last meeting",
			"Review Fedora 45 Release Schedule",
			"Migrate Newsletter repository to the CoreOS organization",
			"Open Floor",
		},
		Decisions: []string{
			"Vote passed: Migrate newsletter repository to CoreOS GitHub organization",
			"Fedora 45 release final freeze with early target date of 2026-10-20",
		},
		ActionItems: []string{
			"Yasmin Valim de Souza (ydesouza): Proceed with migrating the newsletter repository to the CoreOS organization",
			"Cade Jacobson (cadejacobson): Open a pull request with unit tests for Ignition's Azure logic to document its current behavior",
		},
		Raw: map[string]interface{}{
			"url":     "https://meetbot.fedoraproject.org",
			"fetched": time.Now(),
			"source":  "sample_data",
			"note":    "This is sample data from the October 14, 2026 meeting for demonstration",
		},
	}
}
