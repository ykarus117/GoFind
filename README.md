# GoFind

GoFind is a lightweight household inventory management application designed to help you store and retrieve information about items and their locations in your home. With minimal system resource requirements, GoFind offers a practical solution for keeping track of belongings, whether you're organizing a closet, managing storage, or simply remembering where you placed that important document. 

I'm developing this application to help myself at home as I'm the kind of person that struggles with remembering the existance of household objects ("out of sight out of mind"). The project is not yet complete and has to be considered on ongoing development. This project at the current state not suitable for any production environment (it has no TSL/SSL support for example) if you choose to self-host it and expose it to the internet consider using a reverse proxy.

## Features

Built with simplicity and efficiency in mind, GoFind allows you to:

- Store detailed information about household items
- Track the location and position of belongings
- Run as a self-hosted service on your own infrastructure
- Manage items with minimal system overhead

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

For a quick setup as a systemd service:

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

## Uninstall

To remove the system installation:

```
sudo make uninstall
```

This will stop the service, remove systemd configuration, and clean up all installed files and the dedicated system user.

## Contributing

Contributions are welcome. Whether you have suggestions for improvements, bug reports, or want to contribute code, please feel free to open an issue or pull request. Please refrain from submitting unvetted AI generated code.

## License

GoFind is released under the GNU General Public License v3.0. See the LICENSE file for full details. The GPL v3.0 requires that any modifications or derivative works also be licensed under the same terms and made available to users.

Part of the tree rendering code has been developed starting from D3 examples gallery that fall under the following license:

"Copyright 2018–2020 Observable, Inc.

Permission to use, copy, modify, and/or distribute this software for any
purpose with or without fee is hereby granted, provided that the above
copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE."

## AI usage disclaimer
No AI model was used to explicitly write code, some code suggestions were derived from llms (as google intregrated AI search feature) this readme has been auto generated from github Copilot and manually edited.
