[![GitHub Stars](https://img.shields.io/github/stars/backendsystems/nibble?style=social)](https://github.com/backendsystems/nibble)
[![PPA](https://img.shields.io/badge/Ubuntu-PPA-E95420?logo=ubuntu&logoColor=white)](#install-apt)
[![COPR](https://img.shields.io/badge/Fedora-COPR-51A2DA?logo=fedora&logoColor=white)](#install-dnf)
[![AUR](https://img.shields.io/aur/version/nibble-bin)](#install-aur)
[![brew](https://img.shields.io/badge/brew-tap-FBB040)](#install-brew)
[![winget](https://img.shields.io/badge/winget-package-0078D4?logo=windows&logoColor=white)](#install-winget)
[![npm](https://img.shields.io/npm/v/@backendsystems/nibble)](#install-npm)
[![PyPI](https://img.shields.io/pypi/v/nibble-cli)](#install-pip)
[![Go Report Card](https://goreportcard.com/badge/github.com/backendsystems/nibble)](https://goreportcard.com/report/github.com/backendsystems/nibble)

<div align="center">
  <img src="assets/nibble.svg" alt="Nibble" width="200">
</div>


Nibble is a CLI tool for local network scanning that focuses on speed and ease of use.

Select a network interface, and Nibble scans your local subnet. Lists hosts, hardware manufacturer, open ports and their services.

![Nibble demo](demo.gif "Made with Bubble Tea VHS")

- Lightning fast scans using lightweight threads
- Stealthy, emits no network signals before a scan is started
- Colors uses your terminal theme colors
- Skips loopback and irrelevant adapters
- Defaults to SSH, Telnet, HTTP, HTTPS, SMB, RDP, and more
- Can be set to a list of custom ports that are stored for future use
- Target mode for targeted network scans
- Reads service banners on open ports (for example, OpenSSH or nginx versions)
- Looks up hardware vendors:
  - Raspberry Pi, Ubiquiti, Apple and 40,000 other vendor ids
- [Headless mode](#headless-mode) with JSON output for scripting and automation

## History
See past scans, the found hosts and re-scan all hosts ports. hotkey: `r`  
History remembers your position between sessions, so jump right back in to your last viewed scan.

![Nibble history](history.gif "Made with Bubble Tea VHS")

## Hotkeys
`↑/↓/←/→`, `w/s/a/d`, `h/j/k/l`: selection
`Enter`: confirm
`p`: select ports
`r`: history
`t`: target mode
`q`: cancel
`Ctrl+C`: quit
`?`: help

## Mouse
Full mouse support. Click to select, click again to confirm. Scroll to navigate lists.
Hold `Shift` and drag to select text.

![Nibble click interface](click.gif "Clickable interface with mouse support")

## Installation
<a id="install-go"></a>
<img src="https://cdn.simpleicons.org/go/00ADD8" width="16" style="vertical-align:middle"> go (https://go.dev/):
```bash
go install github.com/backendsystems/nibble@latest
```
<a id="install-apt"></a>
<img src="https://cdn.simpleicons.org/ubuntu/E95420" width="16" style="vertical-align:middle"> apt (Ubuntu, Mint, Pop!_OS, Zorin, Elementary, KDE Neon):
```bash
sudo add-apt-repository ppa:backendsystems/ppa
sudo apt install nibble
```
<a id="install-dnf"></a>
<img src="https://cdn.simpleicons.org/fedora/51A2DA" width="16" style="vertical-align:middle"> dnf (Fedora, RHEL, CentOS Stream):
```bash
sudo dnf copr enable @backendsystems/nibble
sudo dnf install nibble
```
<a id="install-aur"></a>
<img src="https://cdn.simpleicons.org/archlinux/1793D1" width="16" style="vertical-align:middle"> aur (Arch Linux):
```bash
yay -S nibble-bin
```
<a id="install-brew"></a>
<img src="https://cdn.simpleicons.org/homebrew/FBB040" width="16" style="vertical-align:middle"> brew (macOS):
```bash
brew install backendsystems/tap/nibble
```
<a id="install-winget"></a>
🪟 winget (Windows):
```bash
winget install backendsystems.nibble
```
<a id="install-pip"></a>
<img src="https://cdn.simpleicons.org/python/3776AB" width="16" style="vertical-align:middle"> pip:
```bash
pipx install nibble-cli
```
<a id="install-npm"></a>
<img src="https://cdn.simpleicons.org/npm/CB3837" width="16" style="vertical-align:middle"> npm:
```bash
npm install -g @backendsystems/nibble
```
or run without install
```bash
npx @backendsystems/nibble
```

#### Manual download:
Pre-built binaries for Linux, macOS and Windows (amd64/arm64) are available on the [Releases](https://github.com/backendsystems/nibble/releases) page.

## Usage
Run the CLI with `nibble`, select a network interface.  
Interface icons: `🔌`Ethernet, `📶`Wi-Fi, `📦`Container, `🔒`VPN.

## Headless Mode
Run scans without the TUI. Outputs JSON.  
Headless scans are not saved in history.

`-i` scan target(s), comma-separated or a file ([example_input](internal/parameters/example_input.txt))  
`-p` custom ports (e.g. `22,80,8000-8100` or `-` for all)  
`-o` write output to file instead of stdout ([example_output](internal/parameters/example_output.json))

```bash
nibble -i 192.168.0.0/24
nibble -i 192.168.1.223,10.0.0.12/32 -p - -o results.json
nibble -i targets.txt -p 22,80,443,8000-8100
```

Exit codes: `0` success, `1` error, `2` invalid usage.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)

## License

This project is MIT licensed. See the [LICENSE](LICENSE) file for details.

Note: The "nibble" name and branding assets are excluded from this license, see the separate [LICENSE](assets/LICENSE) for branding terms.
