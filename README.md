# EDCITR (ED: Chat Is This Real)

# Warning: in-development rough state, not ready for release yet 
Didn't get time and lost interest to polish this up for a proper release yet. Soon™
YT quick APP overview: https://youtu.be/xYEAxzrO1rA

[![Latest Release](https://img.shields.io/github/v/release/gethiox/edcitr?sort=semver)](https://github.com/gethiox/edcitr/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/gethiox/edcitr)](https://goreportcard.com/report/github.com/gethiox/edcitr)
[![Build Status](https://img.shields.io/github/actions/workflow/status/gethiox/edcitr/build.yml?branch=main)](https://github.com/gethiox/edcitr/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/gethiox/edcitr/pkg.svg)](https://pkg.go.dev/github.com/gethiox/edcitr/pkg)
[![License: GPLv3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://github.com/gethiox/edcitr/blob/main/LICENSE)
[![EDDN Discord](https://img.shields.io/badge/Discord-EDDN-7289DA?logo=discord&logoColor=white)](https://discord.gg/your-invite-link-here)

EDCITR is a helper application for Elite Dangerous designed for advanced realtime chat monitoring.
It provides sound notifications and robust filtering capabilities across different channels,
allowing commanders to stay informed without constantly checking their (heavily bugged and cumbersome) comms panel.

todo: https://shields.io/

## Features

* **Multi-Channel Monitoring:** Track specific in-game communication channels.
* **Sound Notifications:** Customizable audio alerts for specific keywords or sender types.
* **Advanced Filtering:** Ignore noise and focus only on the messages that matter to your gameplay.
* **Inara Lookup integration:** (opt-in) Instantly fetch Inara profile information about encountered commanders.
* **Game Server Monitoring:** (opt-in) Historical tracking of server health response endpoint (the one ED launcher reports).
* **Meaningful Logs:** Monitor app behavior under terminal output or by revealing mirrored stdout logs by dragging status bar up.
* **Personalization:** Many features have detailed customization configuration options, including color scheme.
* **Blazingly Fast Meme Journal Processing:** Load a years worth of your journal collection in a matters of few seconds.

## Installation

### Pre-compiled Binaries
You can download the latest compiled binaries for your operating system from the [Releases page](https://github.com/gethiox/edcitr/releases).

### Building from Source
Ensure you have Go 1.21 or later installed.

    git clone https://github.com/gethiox/edcitr.git
    cd edcitr
    go build -o edcitr ./cmd/edcitr

## Usage as a Library

Certain core components of EDCITR, such as the log parsing and filtering engines, are exposed as reusable Go packages. If you are building your own Elite Dangerous tools, you can import them:

    import "github.com/gethiox/edcitr/pkg/chatparser"

Refer to the [Go Reference](https://pkg.go.dev/github.com/gethiox/edcitr/pkg) for API documentation.

## Contributing

Contributions are welcome! Please review the [CONTRIBUTING.md](CONTRIBUTING.md) file for guidelines on how to open issues and submit pull requests.

Note: All contributors are required to sign a Contributor License Agreement (CLA) before their pull requests can be merged. A bot will assist you with this process when you open your first PR.

## License and Branding

* **Code:** The source code is licensed under the [GNU GPLv3](LICENSE).
* **Branding:** The name "EDCITR", "ED: Chat Is This Real", and associated API identifiers are trademarks of Sławomir Kur. If you fork or redistribute this project, you must comply with the [Trademark Policy](TRADEMARK.md) by removing and replacing these identifiers.










### TODO 

in FID not found at boot, keep looking in live event streem for Load Game events.

profiles are not updated until column is scrolled

### NOTES
1. gcc is needed for cgo, on windows you can handle it with MSYS2 `winget install -e --id MSYS2.MSYS2`
> `pacman -S mingw-w64-x86_64-gcc 

2. Add `C:\msys64\mingw64\bin` to `PATH` environment variable

3. jsonv2 experinet have to be set and cgo enabled:
> `$env:GOEXPERIMENT = "jsonv2"`
 
> `$env:GOEXPERIMENT = "jsonv2"; $env:CGO_ENABLED=1; go run .\cmd\EDCITR\`



