# Fedora CoreOS Newsletter Generator

A community newsletter generator for Fedora CoreOS. Collects pull requests from GitHub and meeting logs, then generates a markdown newsletter.

## Requirements

- Go 1.22 or later
- GitHub token (for API access)
- LLM API key (Google Gemini or Anthropic Claude)

## Installation

Clone and build:

```bash
git clone https://github.com/coreos/coreos-newsletter.git
cd coreos-newsletter
make build
```

Binary will be in `dist/coreos-newsletter`

## Configuration

Create a `.env` file:

```bash
GITHUB_TOKEN=your_github_token
GITHUB_OWNER=coreos
GITHUB_REPO=coreos-assembler,coreos-installer

# Choose one LLM provider:
GOOGLE_API_KEY=your_gemini_key          # for Google Gemini
# OR
ANTHROPIC_API_KEY=your_claude_key       # for Anthropic Claude
```

All configuration comes from `.env` or environment variables.

## Usage

### Generate Newsletter

```bash
./dist/coreos-newsletter generate --days 7 --out newsletter.md
```

Options:
- `--days N`: Look back this many days (default: 7)
- `--out FILE`: Output to file (default: stdout)

### Show Data Summary

```bash
./dist/coreos-newsletter summary --days 7
```

### Validate Configuration

```bash
./dist/coreos-newsletter validate
```

### Show Version

```bash
./dist/coreos-newsletter version
```

## Data Sources

- **GitHub**: Pull requests from specified repositories
- **Meetbot**: Fedora CoreOS community meeting logs (parsed from HTML)

## Development

Build for development:

```bash
make dev
```

Run tests:

```bash
make test
```

Code quality:

```bash
make lint
make fmt
make vet
```

## License

Apache License 2.0
