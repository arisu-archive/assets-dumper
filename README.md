# 🎮 Assets Dumper

<div align="center">

<picture><img src="https://img.shields.io/badge/dynamic/yaml?url=https%3A%2F%2Fba.pokeguy.dev%2Fcom.nexon.bluearchive%2Fversion.txt&query=%24&prefix=v&style=for-the-badge&logo=nexon&label=Global&color=0099ff" alt="Nexon BlueArchive Latest Version" style="visibility:visible;max-width:100%;"></picture><picture><img src="https://img.shields.io/badge/dynamic/yaml?url=https%3A%2F%2Fba.pokeguy.dev%2Fcom.YostarJP.BlueArchive%2Fversion.txt&query=%24&prefix=v&style=for-the-badge&logo=googleplay&label=Yostar&color=7d3cc8" alt="Yostar BlueArchive Latest Version" style="visibility:visible;max-width:100%;"></picture>

<picture><img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go" style="visibility:visible;max-width:100%;"></picture>

</div>

## 🚀 Overview

A powerful CLI tool designed to extract and download game assets from Blue Archive servers. This utility simplifies the process of accessing game resources by downloading and organizing them into a structured format for easier analysis and usage.

### ✨ Key Features

- **📊 Version Checking**: Verify game versions before extraction
- **📦 Complete Asset Dump**: Extract all assets from game files
- **⚡ Parallel Downloads**: Accelerate extraction with concurrent processing
- **🔍 Asset Filtering**: Target specific assets using glob patterns
- **🔐 Decryption Support**: Handle encrypted game assets automatically
- **🌍 Multi-Region Support**: Compatible with both Global and Japan game servers

## 📋 Prerequisites

| Tool | Minimum Version |
|:----:|:---------------:|
| Go   | ≥ 1.22          |

## ⚙️ Setup Instructions

### 1. Clone the Repository

```bash
# Clone this repository
git clone https://github.com/arisu-archive/assets-dumper.git
cd assets-dumper
```

### 2. Build or Run Directly

```bash
# Build the tool
go build -o assets-dumper

# Or run directly with Go
go run main.go [command] [flags]
```

## 🛠️ Usage Instructions

### Checking Game Version

Verify the current game version before proceeding with asset extraction:

```bash
# Check version (using go run)
go run main.go version --server global

# Or with built binary
./assets-dumper version --server japan
```

### Downloading Game Assets

Extract assets from a specific game server with optional filtering:

```bash
# Download all assets from global server
go run main.go download --server global --output ./output

# Download only specific assets using glob pattern
go run main.go download --server japan --output ./japan-assets --filter "**/*.png" --max-concurrency 20
```

**Parameters:**
| Flag | Description |
|------|-------------|
| `-s, --server <server>` | Server to download from (global/japan) |
| `-o, --output <path>` | Path to download assets to |
| `-f, --filter <glob>` | Glob pattern to filter assets (default: "**") |
| `-c, --max-concurrency <N>` | Maximum number of concurrent downloads (default: 10) |
| `-v, --verbose` | Enable verbose logging |

## 🔍 How It Works

1. **Version Check**: First checks the current game version on the selected server
2. **Resource Discovery**: Identifies all available assets on the server
3. **Filtering**: Applies the specified glob pattern to select desired assets
4. **Parallel Download**: Downloads selected assets using concurrent workers
5. **Storage**: Saves all downloaded assets to the specified output directory, preserving path structure

## 📁 Project Structure

```
├── cmd/
│   ├── root/          # Main command definitions
│   ├── version/       # Version checking functionality
│   └── download/      # Asset downloading functionality
├── pkg/
│   ├── global/        # Global server implementation
│   ├── japan/         # Japan server implementation
│   └── resources/     # Shared resource handling
├── internal/          # Internal utilities
├── LICENSE            # MIT License
└── README.md          # This documentation
```

## 📈 Roadmap

- [x] Check game versions (Global/Japan)
- [x] Download assets with filtering
- [x] Parallel downloading
- [ ] Improved asset decryption
- [ ] Resource caching for faster re-downloads
- [ ] Support for additional game regions
- [ ] Enhanced progress tracking for downloads
- [ ] Automatic version tracking

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## ⚠️ Disclaimer

This tool is intended for legitimate research and analysis purposes only. Always ensure you comply with the terms of service of any game you analyze and all relevant laws and regulations.

## 📬 Contact & Support

- Create an [Issue](https://github.com/arisu-archive/assets-dumper/issues) for bug reports or feature requests
- Star ⭐ the repo if you find it useful
- Follow for updates on new features and improvements

---

<div align="center">
<strong>Built with ❤️ for the Blue Archive community</strong>
</div>
