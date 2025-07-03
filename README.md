![GitHub release (latest by date)](https://img.shields.io/github/v/release/joeyomi/clipboardpro)
![GitHub](https://img.shields.io/github/license/joeyomi/clipboardpro)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/joeyomi/clipboardpro)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/joeyomi/clipboardpro/test.yaml)

# ClipBoard Pro

Cross-platform clipboard manager built with [Fyne](https://fyne.io) in Go.  
Manages clipboard history across Windows, macOS, and Linux.

## Features

- Cross-platform: Windows, macOS, Linux
- Clipboard history tracking
- Modern UI built with Fyne
- Lightweight with minimal resource usage
- Local storage (data stays on your device)
- Auto-update notifications and one-click updates (WIP)

## Installation

### Download

Download the latest release for your platform from the [Releases page](https://github.com/joeyomi/clipboardpro/releases).

**Downloads:**
- **Windows**: `ClipBoard-Pro-*-windows-amd64.zip`
- **macOS**: `ClipBoard-Pro-*-darwin-amd64.tar.gz` (Intel) or `ClipBoard-Pro-*-darwin-arm64.tar.gz` (Apple Silicon)
- **Linux**: `ClipBoard-Pro-*-linux-amd64.tar.xz`

### Installation Instructions

> **Security Notice**: All builds are unsigned, so your operating system may show security warnings.  
> This is expected as I don't have any commercial code signing certificates at the moment.

#### Windows

1. Download `ClipBoard-Pro-*-windows-amd64.zip`
2. Extract the ZIP file
3. **Security Warning**: Windows will show "Windows protected your PC" message
4. **To run the app**:
   - Click **"More info"** on the SmartScreen warning
   - Click **"Run anyway"** 
   - **Alternative**: Right-click the `.exe` file → **Properties** → Check **"Unblock"** → **Apply** → **OK**
5. **Optional**: Move the `.exe` to a permanent location (e.g., `C:\Program Files\ClipBoard Pro\`)

#### macOS

1. Download the appropriate `.tar.gz` file for your Mac:
   - Intel Macs: `ClipBoard-Pro-*-darwin-amd64.tar.gz`
   - Apple Silicon Macs: `ClipBoard-Pro-*-darwin-arm64.tar.gz`
2. Extract the archive (double-click or use `tar -xzf filename.tar.gz`)
3. **Installation**: Drag `ClipBoard Pro.app` to your Applications folder
4. **Security Warning**: macOS will show "ClipBoard Pro.app is damaged and can't be opened"
5. **To run the app**:
   ```bash
   # Open Terminal and run:
   sudo xattr -rds com.apple.quarantine "/Applications/ClipBoard Pro.app"
   
   # Enter your password when prompted
   # Now you can launch the app from Applications or Launchpad
   ```

#### Linux

1. Download `ClipBoard-Pro-*-linux-amd64.tar.xz`
2. Extract the archive:
   ```bash
   tar -xf ClipBoard-Pro-*-linux-amd64.tar.xz
   cd usr/local/bin  # Navigate to extracted folder
   ```
3. **Run the app**:
   ```bash
   ./ClipBoard\ Pro
   ```
4. **Optional**: Install system-wide:
   ```bash
   # Extract to system directories (requires sudo)
   sudo tar -xf ClipBoard-Pro-*-linux-amd64.tar.xz -C /
   
   # Run from anywhere
   clipboardpro
   ```

### Auto-Update (WIP)

Once installed, ClipBoard Pro will automatically:
- Check for updates when the app starts (configurable)
- Notify you when new versions are available
- Allow one-click updates without reinstalling
- Preserve your settings and clipboard history accross updates

## Development

### Prerequisites

- Go 1.23 or later
- C compiler (for Fyne's native bindings)
- Platform-specific development tools:
  - **macOS**: Xcode Command Line Tools
  - **Linux**: `gcc`, `libgl1-mesa-dev`, `xorg-dev`, `libxkbcommon-dev`
  - **Windows**: TDM-GCC or similar

### Building from Source

```bash
# Clone the repository
git clone https://github.com/joeyomi/clipboardpro.git
cd clipboardpro

# Install dependencies
go mod tidy

# Build for your current platform
go build -o clipboardpro

# Create a packaged app for distribution
go install fyne.io/tools/cmd/fyne@latest
fyne package -os linux   # Linux
fyne package -os darwin  # macOS  
fyne package -os windows # Windows
```

### Release Process

The project uses GitHub Actions for automated releases:

- **Tag a release**: `git tag v1.2.3 && git push origin v1.2.3`
- **Automatic builds**: Native builds on Windows, macOS, and Linux runners
- **Dual assets**: Both user-friendly packages and raw binaries for auto-updates
- **Checksums**: Automatic SHA256 verification files

## Built With

- [Fyne](https://fyne.io) - Cross-platform GUI framework
- [Go](https://golang.org) - Programming language
- [go-selfupdate](https://github.com/creativeprojects/go-selfupdate) - Auto-update functionality
