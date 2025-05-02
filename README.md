# ScanExpress

A simple and efficient document scanning utility written in Go with a terminal user interface.

## Overview

ScanExpress is a command-line application that provides a convenient way to control document scanners, scan multiple pages, and automatically generate PDF documents from the scanned images. It uses the `scanimage` utility under the hood and provides a friendly TUI (Terminal User Interface) built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- Auto-detect available scanners
- Save scanner configuration for future use
- Simple TUI for selecting scanners and configuring scan options
- Support for scanning multiple pages
- Support for duplex (double-sided) scanning
- Automatic PDF generation from scanned images
- Auto-deskew and auto document size detection

## Demo



https://github.com/user-attachments/assets/1fa8c911-374f-431b-b151-b9ecba2f8b88



## Requirements

- `scanimage` (SANE backend) for scanner access
- `img2pdf` for PDF generation

## Installation

### From Source

Build and install the application to `$HOME/.local/bin` (default) or a custom path:

```bash
git clone https://github.com/midoBB/scanexpress.git
cd scanexpress
make build
# Install to default location
make install
# Install to a custom location
make install INSTALL_PATH=/path/to/install
```

### From Binary

Download the latest release from the [releases page](https://github.com/midoBB/scanexpress/releases).

## Usage

Run the application:

```bash
./scanexpress
```

### Command-line Options

- `-s, --select`: Force scanner selection even if one is already configured

### Configuration

The application stores configuration in `~/.config/scanexpress/config.yaml`. This includes:

- Previously selected scanner
- Default save folder

## Workflow

1. Select a scanner from the list of available devices
2. Choose a folder to save scanned documents
3. Enter the number of pages to scan
4. Select scan mode (single-sided or duplex)
5. Follow the prompts to scan documents
6. A PDF will be automatically generated when scanning is complete

## Todo / Roadmap

- Better input for duplex/single page scans
- Add a progress bar
- Add the option to customize the DPI and other scan parameters in the TUI or in the config file
