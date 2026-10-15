# Fedora CoreOS Newsletter Generator

A simple Go CLI tool that generates community newsletters by collecting GitHub pull requests and Fedora CoreOS meeting logs, then using Claude or Gemini to create a friendly summary.

## Quick Start

### Prerequisites

- Go 1.22+ (to build, or use the pre-built binary)
- API key from either:
  - **Claude** (Anthropic): https://console.anthropic.com/
  - **Gemini** (Google): https://ai.google.dev/
- GitHub token (optional, recommended for better rate limits)

### Setup

1. **Get the binary** (already built in `dist/coreos-newsletter`)

2. **Set your API keys as environment variables:**

**For Claude (Anthropic):**
```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

**For Gemini (Google):**
```bash
export GOOGLE_API_KEY=your_api_key...
```

3. **Create `.env` file:**
```bash
cp .env.example .env
```

4. **Edit `.env` to set:**
```bash
# GitHub configuration (always needed)
GITHUB_OWNER=coreos
GITHUB_REPO=coreos-assembler,coreos-installer

# Choose your LLM provider
LLM_PROVIDER=claude    # or "gemini"
CLAUDE_MODEL=claude-sonnet-4-20250514    # if using claude
GEMINI_MODEL=gemini-2.5-flash             # if using gemini
```

### Run

Generate a newsletter for the last 7 days:

```bash
./dist/coreos-newsletter generate --days 7 --out newsletter.md
```

Your newsletter will be in `newsletter.md`

## Commands

**Generate newsletter:**
```bash
./dist/coreos-newsletter generate --days 7 --out output.md
```

**Show data summary:**
```bash
./dist/coreos-newsletter summary --days 7
```

**Validate configuration:**
```bash
./dist/coreos-newsletter validate
```

**Show version:**
```bash
./dist/coreos-newsletter version
```

## How It Works

1. Collects GitHub PRs from your specified repositories
2. Fetches Fedora CoreOS community meeting logs
3. Ranks PRs by importance and identifies contributors
4. Sends data to your chosen LLM (Claude or Gemini) for newsletter generation
5. Outputs clean markdown file

## Choosing an LLM Provider

### Claude (Recommended)
- **Provider**: Anthropic
- **Setup**: Set `ANTHROPIC_API_KEY` environment variable
- **Config**: `LLM_PROVIDER=claude`
- **Model**: `claude-sonnet-4-20250514` (or other Claude models)

### Gemini
- **Provider**: Google
- **Setup**: Set `GOOGLE_API_KEY` environment variable
- **Config**: `LLM_PROVIDER=gemini`
- **Model**: `gemini-2.5-flash` (or other Gemini models)

Both providers read the API key from environment variables - no need to paste keys in `.env` file!

## Configuration

### Required
- `GITHUB_OWNER` - GitHub organization (e.g., `coreos`)
- `GITHUB_REPO` - Comma-separated list of repos

### For Claude
- `LLM_PROVIDER=claude`
- `ANTHROPIC_API_KEY` - Set in environment: `export ANTHROPIC_API_KEY=sk-ant-...`

### For Gemini
- `LLM_PROVIDER=gemini`
- `GOOGLE_API_KEY` - Set in environment: `export GOOGLE_API_KEY=...`

### Optional
- `GITHUB_TOKEN` - GitHub token for better rate limits

## Output Example

```markdown
# Fedora CoreOS Newsletter

## This Week's Highlights

The Fedora CoreOS project moves forward with valuable contributions and active participation.

## Notable Contributions

- Pull requests improving core functionality have been merged
- Meeting discussions focused on feature planning and roadmap alignment
- Documentation improvements make it easier for new contributors to get started

## Thanks to All Contributors

We appreciate the ongoing support and contributions from all who help drive Fedora CoreOS forward!
```

## Development

Build:
```bash
make dev      # Development build
make build    # Release build
make clean    # Clean artifacts
```

Test:
```bash
make test
```

Code quality:
```bash
make fmt
make lint
make vet
```

## License

Apache License 2.0
