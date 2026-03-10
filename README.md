# Nibble
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
Click to select, click again to confirm. Scroll to navigate lists.
Right-click to go back (same as pressing `q`).  Hold `Shift` and drag to select text.

## Installation
you may have to restart terminal to run `nibble` after install.


go:
```bash
go install github.com/backendsystems/nibble@latest
```
brew:
```bash
brew install backendsystems/tap/nibble
```
pip:
```bash
pipx install nibble-cli
```
npm:
```bash
npm install -g @backendsystems/nibble
```
or run without install
```bash
npx @backendsystems/nibble
```

## Usage
Run the CLI with `nibble`, select a network interface.  
Interface icons: `🔌` = Ethernet, `📶` = Wi-Fi, `📦` = Container, `🔒` = VPN.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
