# GoFind

GoFind is a lightweight household inventory management application designed to help you store and retrieve information about items and their locations in your home. With minimal system resource requirements, GoFind offers a practical solution for keeping track of belongings, whether you're organizing a closet, managing storage, or simply remembering where you placed that important document.

## Overview

GoFind provides a web-based interface for managing your household inventory. The application combines a Go backend for reliable performance with a responsive HTML and CSS frontend, creating a self-hosted solution that runs efficiently on modest hardware. Perfect for users who value privacy and control over their personal data.

## Features

Built with simplicity and efficiency in mind, GoFind allows you to:

- Store detailed information about household items
- Track the location and position of belongings
- Access your inventory through a clean web interface
- Run as a self-hosted service on your own infrastructure
- Manage items with minimal system overhead

## Technology Stack

The application is built using:

- Go: Backend server and core logic
- SQLite: Lightweight database for persistent storage
- HTML/CSS: Responsive user interface
- JavaScript: Client-side interactions

## Installation

### Prerequisites

- Go 1.16 or later
- CGO enabled compiler
- Linux or Windows system (ARM64 and x86-64 support)

### Quick Start

Clone the repository and build the application:

```
git clone https://github.com/ykarus117/GoFind.git
cd GoFind
make build
```

This creates a binary in the `bin/` directory.

### System Installation (Linux)

For a production setup as a systemd service:

```
sudo make install
```

The installer will:

- Create a dedicated system user for GoFind
- Install the binary to `/usr/local/bin/GoFind/`
- Configure systemd service management
- Set appropriate permissions

Manage the service with:

```
sudo systemctl start gofind
sudo systemctl stop gofind
sudo systemctl status gofind
```

Access the web interface at `http://localhost:8080`

### Configuration

Set the database path using the `DB_PATH` environment variable:

```
export DB_PATH=/path/to/custom/location/GoFind.db
```

If not specified, the database defaults to `./GoFind.db` in the current directory.

## Cross-Platform Builds

The Makefile includes targets for building on different platforms:

- `make cc_build`: Builds for Windows x64, Linux x64, and Linux ARM64

This is useful for creating distribution binaries for different target systems.

## Uninstall

To remove the system installation:

```
sudo make uninstall
```

This will stop the service, remove systemd configuration, and clean up all installed files and the dedicated system user.

## Contributing

Contributions are welcome. Whether you have suggestions for improvements, bug reports, or want to contribute code, please feel free to open an issue or pull request.

## License

GoFind is released under the GNU General Public License v3.0. See the LICENSE file for full details. The GPL v3.0 requires that any modifications or derivative works also be licensed under the same terms and made available to users.
