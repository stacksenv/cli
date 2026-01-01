# StacksEnv Installer

This repository contains installation scripts for [StacksEnv CLI](https://github.com/stacksenv/cli).

## Installation

### Linux, macOS, and BSD Systems

You can install StacksEnv using Homebrew, `curl`, or `wget`:

#### Using Homebrew

```bash
brew install stacksenv/tap/stacksenv
```

#### Using curl

```bash
curl -fsSL https://raw.githubusercontent.com/stacksenv/get/main/get.sh | bash
```

#### Using wget

```bash
wget -qO- https://raw.githubusercontent.com/stacksenv/get/main/get.sh | bash
```

**Note:** The installation script may require `sudo` privileges to install StacksEnv to `/usr/local/bin` (or `/usr/bin` if `/usr/local/bin` doesn't exist).

### Windows

For Windows systems, use PowerShell (run as Administrator):

```powershell
iwr -useb https://raw.githubusercontent.com/stacksenv/get/main/get.ps1 | iex
```

## Requirements

### Unix-like Systems (Linux, macOS, BSD)

- `bash`
- `curl` or `wget`
- `tar` (for Linux/BSD) or `unzip` (for macOS/Windows)
- `sudo` (may be required for installation to system directories)

### Windows

- PowerShell (run as Administrator)
- Internet connection

## What the Installer Does

1. Detects your operating system and architecture
2. Downloads the latest StacksEnv release from GitHub
3. Extracts the binary
4. Installs it to a directory in your PATH
5. Makes the binary executable (Unix-like systems)

## Supported Platforms

- **Linux**: amd64, 386, arm64, armv5, armv6, armv7
- **macOS (Darwin)**: amd64, arm64
- **FreeBSD**: amd64, 386
- **NetBSD**: amd64, 386
- **OpenBSD**: amd64, 386
- **Windows**: amd64, 386

## Verification

After installation, verify that StacksEnv is installed correctly:

```bash
stacksenv --version
```

## Troubleshooting

### Installation fails with "Aborted, could not find curl or wget"

Install either `curl` or `wget`:

- **Ubuntu/Debian**: `sudo apt-get install curl` or `sudo apt-get install wget`
- **CentOS/RHEL**: `sudo yum install curl` or `sudo yum install wget`
- **macOS**: Usually pre-installed, or install via Homebrew: `brew install curl`

### Permission denied errors

The installer needs to write to system directories. Use `sudo` if needed:

```bash
curl -fsSL https://raw.githubusercontent.com/stacksenv/get/main/get.sh | sudo bash
```

### StacksEnv not found after installation

1. Ensure the installation completed successfully
2. Check that the installation directory is in your PATH
3. Try opening a new terminal session
4. Verify the installation path: `which stacksenv` (Unix) or `where.exe stacksenv` (Windows)

## Manual Installation

If you prefer to install manually:

1. Visit the [StacksEnv CLI releases page](https://github.com/stacksenv/cli/releases)
2. Download the appropriate binary for your platform
3. Extract and place it in a directory in your PATH
4. Make it executable (Unix-like systems): `chmod +x stacksenv`

## Links

- **GitHub Repository**: https://github.com/stacksenv/cli
- **Issues**: https://github.com/stacksenv/cli/issues

## License

See the main StacksEnv CLI repository for license information.

