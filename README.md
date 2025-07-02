# CoH3 Build Order Extractor

A Golang CLI tool that extracts build orders from Company of Heroes 3 replay files (*.rec) using the Rust [vault library](https://github.com/ryantaylor/vault).

## Features

- Parse Company of Heroes 3 replay files
- Extract the first 10 build commands for specific players or all players
- Reference players by name (case-insensitive) or by ID
- Show high-level replay information (teams, players, map, duration, winning team)
- Cross-platform support (Windows, Linux, macOS)
- Fast parsing using Rust backend with Go frontend
- Single statically-linked binary with no external dependencies

## Prerequisites

- Go 1.21 or later
- Rust (latest stable)
- Cargo

## Building

1. Clone this repository
2. Install dependencies:
   ```bash
   make install
   ```

3. Build the project:
   ```bash
   make build
   ```

This will:
- Build the Rust wrapper library with FFI bindings
- Compile the Go CLI application
- Create the `coh3-build-order` executable

## Usage

### Show Replay Information

Get high-level information about a replay:
```bash
./coh3-build-order info replay.rec
```

### Extract Build Orders

Extract build orders for all players:
```bash
./coh3-build-order build-order replay.rec
```

Extract build order for a specific player by name:
```bash
./coh3-build-order build-order -p PlayerName replay.rec
./coh3-build-order build-order -p Tomsch replay.rec
./coh3-build-order build-order -p "Ted 'Seaman' Silk" replay.rec
```

Extract build order for a specific player by ID:
```bash
./coh3-build-order build-order -p 0 replay.rec
./coh3-build-order build-order -p 1 replay.rec
```

### Verbose Output

Enable verbose logging for build order extraction:
```bash
./coh3-build-order build-order -v replay.rec
```

### Help

View all available commands:
```bash
./coh3-build-order --help
./coh3-build-order build-order --help
./coh3-build-order info --help
```

## Output Format

### Replay Information

```
=== Replay Information ===
Map: data:scenarios\multiplayer\rails_and_sand_4p\rails_and_sand_4p
Duration: 30:00
Winning Team: 1

=== Teams ===
Team 1:
  ID 0: Tomsch
  ID 1: Ted 'Seaman' Silk

Team 2:
  ID 2: Angrybirds
  ID 3: Thomas Smooth
```

### Build Orders

The tool outputs the first 10 build commands in a human-readable format:

```
=== Player 0: PlayerName ===
Build Order:
  1. [00:30] build: engineer
  2. [01:15] build: riflemen from barracks
  3. [02:45] build: at_gun from barracks
  4. [03:30] train: panzerfaust
  5. [04:15] construct: mortar from weapon_support_center
  6. [05:00] build: tank
  7. [05:45] train: scout
  8. [06:30] build: medic from mechanized_company
  9. [07:15] construct: flamethrower
 10. [08:00] build: engineer from airborne_company
```

## Architecture

This project uses Foreign Function Interface (FFI) to bridge between Go and Rust:

- **Rust Layer** (`vault-wrapper/`): Wraps the vault library and exposes C-compatible functions
- **Go Layer** (`vault/`): CGO bindings to call Rust functions from Go
- **CLI Layer** (`cmd/`): Command-line interface using Cobra

## Limitations

- Build order extraction depends on the parsing capabilities of the vault library
- Currently extracts the first 10 commands only
- Some command types may not be fully supported yet  
- The vault library is still under active development
- Player names are case-insensitive but must match exactly
- Replay information (duration, winning team) uses placeholder data as the vault library's parsing capabilities expand

## Development

### Building Individual Components

Build only the Rust wrapper:
```bash
make build-rust
```

Build only the Go CLI:
```bash
make build-go
```

### Cleaning

Remove all build artifacts:
```bash
make clean
```

### Testing

This project includes comprehensive end-to-end testing with a dedicated testing framework in the `tests/` directory.

#### Running Tests

Run all tests:
```bash
go test -v ./tests/
```

Run specific test categories:
```bash
# Integration tests
go test -v -run "TestReplayFiles" ./tests/

# CLI command tests
go test -v -run "TestCLI" ./tests/

# Error handling tests
go test -v -run "TestError" ./tests/

# Performance benchmarks
go test -bench=. ./tests/
go test -bench=. -benchmem ./tests/
```

#### Adding Test Replays

To add new replay files for testing:

1. Place `.rec` files in `tests/testdata/`
2. Create corresponding JSON fixtures in `tests/fixtures/` with expected results
3. Add test cases to `tests/integration_test.go`

Example fixture format:
```json
{
  "file_name": "your_replay.rec",
  "expected_results": {
    "duration": "25:30",
    "player_count": 4,
    "team_count": 2,
    "players": [
      {
        "id": 0,
        "name": "PlayerName",
        "team_id": 1,
        "min_commands": 5,
        "has_real_commands": true,
        "first_commands": ["train", "build"]
      }
    ]
  }
}
```

The testing framework provides:
- **Integration tests**: Validate parsing against expected results using JSON fixtures
- **CLI tests**: Test all command-line interface functionality
- **Performance benchmarks**: Monitor parsing performance and memory usage
- **Error handling**: Verify proper error conditions and messages
- **Consistency tests**: Ensure all parsing functions return consistent data

See `tests/README.md` for detailed testing documentation.

## Contributing

This project uses the [vault library](https://github.com/ryantaylor/vault) for replay parsing. If you encounter parsing issues or missing command types, consider contributing to the upstream vault project.

## License

This project follows the same license as the vault library it depends on.