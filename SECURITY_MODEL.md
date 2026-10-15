# Security Model: Public vs Internal Newsletters

## Overview

The newsletter generator enforces a strict security boundary between **public** and **internal** newsletters to prevent accidental leakage of sensitive business information.

## Two Modes

### 1. General Mode (Default, Public)

```bash
./coreos-newsletter general --days 7 --out newsletter.md
```

**Safe for public distribution**

Data sources:

- ✅ GitHub default repositories (public PRs)
- ✅ Fedora CoreOS meetbot logs (community meetings)

Excluded:

- ❌ GitLab (internal)
- ❌ Jira (contains client names and sensitive issues)

Use case: Community newsletters, public documentation, external communications

### 2. Internal Mode (Requires Explicit Flag)

```bash
./coreos-newsletter internal --days 7 --out newsletter.md
```

**INTERNAL ONLY - Never distribute externally**

Data sources:

- ✅ GitHub repositories
- ✅ GitLab merge requests (INTERNAL)
- ✅ Jira issues (INTERNAL - client information)
- ✅ Fedora CoreOS meetbot logs

Requirements:

- Must explicitly use `internal` command (not default)
- Requires valid GitLab credentials (GITLAB_TOKEN)
- Requires valid Jira credentials (JIRA_BASE_URL, JIRA_EMAIL, JIRA_API_TOKEN)
- Will refuse to generate if credentials are missing
- Security warning prepended to output

Use case: Internal team updates, management reporting, project tracking

## Security Features



### 1. Credential Validation

Internal mode validates all required credentials before proceeding:

```
If credentials missing:
  ✗ GitLab credentials missing (set GITLAB_TOKEN for --internal mode)
  ✗ Jira credentials missing (set JIRA_BASE_URL, JIRA_EMAIL, JIRA_API_TOKEN for --internal mode)
  → Generation refused
```



### 2. Explicit Opt-In

- Default mode is `general` (safe)
- Internal mode requires explicit `internal` command
- No silent fallback or hidden behavior
- User always knows what mode they're running



### 3. Output Warnings

Internal newsletters include security warning:

```
⚠️  INTERNAL ONLY - DO NOT DISTRIBUTE EXTERNALLY ⚠️
```



### 4. Prompt Segregation

Two different LLM prompts:

**GeneralNewsletterPrompt:**

- Focuses on community contributions
- No mention of client details
- Welcomes external participation
- Safe for public sharing

**InternalNewsletterPrompt:**

- Includes project status and tracking
- May reference client names (Jira)
- Intended for internal consumption only
- Warns about sensitive information



### 5. Data Boundaries

**General bundle contains:**

```go
bundle.PullRequests    // GitHub PRs only
bundle.Meetings        // Public meetbot logs
// bundle.JiraIssues is empty (not fetched)
```

**Internal bundle contains:**

```go
bundle.PullRequests    // GitHub + GitLab
bundle.JiraIssues      // SENSITIVE - client data
bundle.Meetings        // Public meetbot logs
```



## CLI Interface



### Commands

```bash
# Public newsletter (safe for distribution)
./coreos-newsletter general

# Internal newsletter (requires credentials)
./coreos-newsletter internal

# Check if credentials are available
./coreos-newsletter validate

# Show data summary (uses general by default, respects --internal flag)
./coreos-newsletter summary [--internal]
```



### Global Flags

```bash
--config <file>  Configuration file (default: .env)
--verbose        Verbose logging
--internal       Include internal data (explicit opt-in for internal mode)
```



## Configuration



### Required for General Mode

```bash
GITHUB_TOKEN=...
GITHUB_OWNER=coreos
GITHUB_REPO=coreos-assembler,coreos-installer
```



### Additional Required for Internal Mode

```bash
GITLAB_TOKEN=...
GITLAB_BASE_URL=https://gitlab.com
GITLAB_PROJECT_ID=123,456

JIRA_BASE_URL=https://jira.example.com
JIRA_EMAIL=user@example.com
JIRA_API_TOKEN=...
JIRA_PROJECT_KEY=COREOS
```



## Error Handling



### Scenario 1: Requesting internal mode without credentials

```
$ ./coreos-newsletter internal
[ERROR] Cannot generate internal newsletter - missing credentials:
  - GitLab credentials missing (set GITLAB_TOKEN for --internal mode)
  - Jira credentials missing (set JIRA_BASE_URL, JIRA_EMAIL, JIRA_API_TOKEN for --internal mode)
```



### Scenario 2: GitHub-only environment (safe mode)

```
$ ./coreos-newsletter general  # ✅ Works fine
$ ./coreos-newsletter internal # ❌ Requires credentials
```



### Scenario 3: Full environment (all modes available)

```
$ ./coreos-newsletter general  # ✅ Works
$ ./coreos-newsletter internal # ✅ Works
```



## Best Practices

✅ **DO:**

- Use `general` mode for external communications
- Use `internal` mode for team-only updates
- Always check the newsletter content before distribution
- Store internal newsletters in secure locations
- Validate credentials with `./coreos-newsletter validate`

❌ **DON'T:**

- Distribute internal newsletters externally
- Share internal mode output on public channels
- Remove the security warning from internal newsletters
- Use `--internal` flag unintentionally
- Store credentials in version control



## Testing

Verify security boundaries:

```bash
# Only GitHub configured - general works, internal fails
unset GITLAB_TOKEN JIRA_BASE_URL JIRA_EMAIL JIRA_API_TOKEN
./coreos-newsletter general   # ✅ Works
./coreos-newsletter internal  # ❌ Fails with clear error

# Full configuration - both modes work
export GITLAB_TOKEN=...
export JIRA_BASE_URL=...
./coreos-newsletter general   # ✅ Works (ignores internal data)
./coreos-newsletter internal  # ✅ Works (includes internal data)
```



## Audit Trail

Each newsletter includes metadata:

```
General mode:
- Data sources: GitHub + Meetbot
- Generated at: [timestamp]
- Mode: PUBLIC

Internal mode:
- Data sources: GitHub + GitLab + Jira + Meetbot
- Generated at: [timestamp]
- Mode: INTERNAL (WARNING)
```



## Future Enhancements

- [ ] Encryption for internal newsletter files
- [ ] Automatic detection of sensitive patterns
- [ ] Email-only distribution for internal mode
- [ ] Audit logging of who generated what
- [ ] Redaction options for client names
- [ ] Data retention policies



## Compliance

This security model helps comply with:

- ✅ GDPR (personal data protection)
- ✅ SOC 2 (data handling)
- ✅ Internal confidentiality agreements
- ✅ Client data protection requirements

Questions? Contact the security team.