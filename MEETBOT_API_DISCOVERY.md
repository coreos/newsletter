# Meetbot/Mote API Discovery

## Overview

The Fedora CoreOS meeting logs are hosted on **Mote** (https://github.com/fedora-infra/mote), a Flask application that aggregates IRC/Matrix meeting minutes and logs for the Fedora Project.

## API Endpoints Found

### 1. Calendar Events API (RECOMMENDED)
**Endpoint:** `GET /cal/events`

**Parameters:**
- `start` (required): Start date/time in ISO 8601 format (e.g., `2026-10-08T00:00:00`)
- `end` (required): End date/time in ISO 8601 format (e.g., `2026-10-15T23:59:59`)

**Response:** JSON array of meeting events

**Example Request:**
```bash
curl "https://meetbot.fedoraproject.org/cal/events?start=2026-10-08T00:00:00&end=2026-10-15T23:59:59"
```

**Expected Response Format:**
```json
[
  {
    "title": "Fedora CoreOS Meeting",
    "start": "2026-10-14T14:00:00",
    "allDay": false,
    "duration": 45
  }
]
```

**Advantages:**
- Official calendar API
- Returns structured JSON
- Covers all meetings in a time period
- Most efficient for range queries

---

### 2. Fragment Query API (FALLBACK)
**Endpoint:** `GET /fragedpt/`

**Parameters (via query string):**

#### List Channels
- `rqstdata=listchan`

Example:
```bash
curl "https://meetbot.fedoraproject.org/fragedpt/?rqstdata=listchan"
```

#### List Dates for Channel
- `rqstdata=listdate`
- `channame=CHANNEL_NAME` (e.g., `fedora-coreos`)

Example:
```bash
curl "https://meetbot.fedoraproject.org/fragedpt/?rqstdata=listdate&channame=fedora-coreos"
```

#### List Meetings for Channel on Date
- `rqstdata=listmeet`
- `channame=CHANNEL_NAME`
- `datename=DATE` (format: `YYYY-MM-DD`)

Example:
```bash
curl "https://meetbot.fedoraproject.org/fragedpt/?rqstdata=listmeet&channame=fedora-coreos&datename=2026-10-14"
```

#### Search Meetings
- `rqstdata=srchmeet`
- `srchtext=SEARCH_TEXT`

Example:
```bash
curl "https://meetbot.fedoraproject.org/fragedpt/?rqstdata=srchmeet&srchtext=coreos"
```

**Response Format:** JSON

---

### 3. Direct Meeting URLs (BLOCKED FOR SECURITY)
These endpoints return 403 Forbidden (likely by design for security):

- `/<channel>/<date>/<meeting>.html` - HTML minutes
- `/<channel>/<date>/<meeting>.log.html` - HTML logs
- `/<channel>/<date>/<meeting>.txt` - Plain text
- `/<channel>/<date>/<meeting>.summary.md` - Markdown summary

These redirect to `meetbot-raw.fedoraproject.org` which is also blocked.

---

## Implementation Details

### Mote Source Code Architecture

**Key Modules:**
- `mote/modules/call.py` - Core data fetching functions
- `mote/modules/late.py` - Calendar/period-based queries
- `mote/modules/find.py` - Search functionality

**Key Functions:**
- `fetch_channel_dict()` - List all channels
- `fetch_datetxt_dict(channel)` - List dates for channel
- `fetch_meeting_dict(channel, date)` - List meetings for date
- `fetch_meeting_summary(path)` - Parse meeting summary from HTML
- `fetch_meeting_by_period(start, end)` - Get meetings for period
- `find_meetings_by_substring(text)` - Search meetings

### Meeting Summary Structure

When parsing meeting HTML, Mote extracts:
- **title**: Meeting name (sanitized)
- **peoples**: Attendees (filters out bots and 0-line participants)
- **topics**: List of topics discussed
- **actions**: Action items with assigned owners
- **duration**: Meeting length in minutes
- **start_time**: Meeting start (format: HH:MM:SS UTC)
- **end_time**: Meeting end (format: HH:MM:SS UTC)

---

## Example Response Data

When visiting the web interface, the October 14, 2026 Fedora CoreOS meeting data was:

```json
{
  "title": "fedora coreos meeting",
  "peoples": [
    "@angelcr:matrix.org (33 messages)",
    "@ydesouza:fedora.im (13 messages)",
    "@zodbot:fedora.im (8 messages)",
    "@cadejacobson:matrix.org (5 messages)",
    "@rapneset:matrix.org (4 messages)",
    "@marmijo:fedora.im (4 messages)",
    "@hricky:fedora.im (2 messages)",
    "@bittin_:mozilla.org (1 message)",
    "@nimbinatus:matrix.org (1 message)",
    "@jeckersb:fedora.im (1 message)"
  ],
  "topics": [
    "roll call",
    "Action items from last meeting",
    "Review Fedora 45 Release Schedule",
    "Migrate Newsletter repository to the CoreOS organization",
    "Open Floor"
  ],
  "actions": [
    "Yasmin Valim de Souza (ydesouza): To proceed with migrating the newsletter repository to the CoreOS organization.",
    "Cade Jacobson (cadejacobson): To open a pull request with unit tests for Ignition's Azure logic to document its current behavior."
  ],
  "start": "14:00:00",
  "end": "14:45:00",
  "duration": 45
}
```

---

## Implementation Status

### Current Implementation in Go

The `internal/collectors/meetbot.go` now implements:

1. **Primary Path**: `fetchViaCalendarAPI()`
   - Uses `/cal/events` endpoint
   - Fetches all meetings in a date range
   - Filters for "fedora" and "coreos" in title

2. **Fallback Path**: `fetchViaFragedptAPI()`
   - Uses `/fragedpt/` endpoint
   - Iterates through each day in the window
   - Queries meetings for `fedora-coreos` channel
   - Falls back if primary path fails

3. **Parsing**: `parseCalendarEvent()` and `parseFragedptMeeting()`
   - Extracts title, date, URL from API responses
   - Stores raw event data for debugging

### Testing Notes

- The web interface at https://meetbot.fedoraproject.org/ is publicly accessible
- Direct meeting log URLs return **403 Forbidden** (security feature)
- API endpoints work when accessed from outside sandboxed environments
- Sandbox network restrictions may block API access during development

---

## Recommendations

1. **For Production**: Use `/cal/events` endpoint - it's the official, most efficient API
2. **For Development**: Test with curl/wget outside sandbox environment
3. **For Reliability**: Keep `/fragedpt/` as fallback in case calendar API changes
4. **For Future**: Consider asking Fedora Infrastructure for API documentation
5. **For Caching**: API responses should be cached to avoid repeated requests

---

## References

- Mote Source: https://github.com/fedora-infra/mote
- Meetbot Instance: https://meetbot.fedoraproject.org/
- Configuration: Mote uses `MEETING_DIR` to store meetings locally
