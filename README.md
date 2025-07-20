# Obsidian CLI

Automatically track programming work in Obsidian daily notes. Discovers git repositories, analyzes commits, and formats intelligent project entries.

## Installation

**Homebrew (recommended):**
```bash
brew tap dylansatow/formulae
brew install obsid
```

**Build from source:**
```bash
go build -o obsid
```

## Setup

Interactive configuration (recommended):
```bash
obsid init
```

Quick setup:
```bash
obsid init --vault ~/Obsidian/Main
```

## Usage

Log all discovered projects:
```bash
obsid log
```

Log current repository:
```bash
obsid log .
```

Log specific timeframe with details:
```bash
obsid log --git-summary --timeframe 2h
```

Create daily note when missing:
```bash
obsid log --create-note
```

View configuration:
```bash
obsid config
```

## Features

- **Smart Discovery**: Finds all git repositories in configured directories
- **Intelligent Formatting**: Converts commits to readable accomplishments  
- **Flexible Timeframes**: Support for `1h`, `2h30m`, `today`, `yesterday`
- **Safe Operation**: Requires explicit permission to create daily notes
- **Area Categorization**: Groups changes by functional areas (frontend, backend, tests)

## Configuration

Stored in `~/.config/obsid/config.yaml`. Configure vault path, project directories, git settings, and formatting preferences through interactive setup.

## Requirements

- Go 1.19+
- Git repositories for tracking
- Obsidian vault with daily notes