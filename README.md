# CoH3 Build Order Extractor

A Golang CLI tool that extracts build orders from Company of Heroes 3 replay files (*.rec) using the Rust [vault library](https://github.com/ryantaylor/vault).

## Features

- Parse Company of Heroes 3 replay files
- Extract build commands for specific players or all players
- **Rich unit name resolution** - Shows actual unit names like "Panzergrenadier Squad", "Riflemen Squad", "Infantry Section" instead of generic commands
- **Battlegroup name resolution** - Shows actual battlegroup names like "Armored (US)" instead of "select_battlegroup"
- **Upgrade name resolution** - Shows specific upgrade names like "T1 Unit Unlock (Afrika Korps)", "Medical Station (Wehrmacht)" instead of "build_global_upgrade"
- **Multi-faction support** - Full support for Wehrmacht, US Forces, Afrika Korps, and British factions
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

### Web Interface (Recommended)

Start the web server for an interactive timeline visualization:
```bash
./coh3-web-server
```

Then open http://localhost:8080 in your browser and drag & drop a .rec file to see:
- **Multi-player timeline columns** - Each player gets their own column for easy comparison
- **Rich build order visualization** - See actual unit names, upgrade names, and battlegroup selections
- **Interactive filters** - Filter by command type (units/buildings/upgrades), faction, or time range
- **Professional interface** - Clean, responsive design that works on any screen size

Optional: specify different port:
```bash
./coh3-web-server 3000
```

### Command Line Interface

#### Show Replay Information

Get high-level information about a replay:
```bash
./coh3-build-order info replay.rec
```

#### Extract Build Orders

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

#### Verbose Output

Enable verbose logging for build order extraction:
```bash
./coh3-build-order build-order -v replay.rec
```

#### Help

View all available commands:
```bash
./coh3-build-order --help
./coh3-build-order build-order --help
./coh3-build-order info --help
```

## Output Format

### Web Interface Timeline

The web interface displays an interactive timeline with:

- **Player columns** - Each player has their own column showing their build order chronologically
- **Rich command descriptions** - "🪖 Built: Panzergrenadier Squad", "🔬 Researched: T1 Unit Unlock (Afrika Korps)"
- **Color-coded players** - Each player gets a unique color for easy identification
- **Real-time filtering** - Filter by command type, faction, or time range
- **Responsive design** - Works on desktop, tablet, and mobile

### Command Line Replay Information

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

### Command Line Build Orders

The CLI tool outputs build commands with rich, human-readable names:

```
=== Player 0: IMPLACÁVEL ===
Build Order:
  1. [01:01] build_squad: Grenadier Squad
  2. [01:02] build_squad: Grenadier Squad
  3. [01:38] construct_entity: Wehrmacht Building (Structure #22)
  4. [02:06] construct_entity: Wehrmacht Building (Structure #57)
  5. [02:32] build_squad: Grenadier Squad
  6. [02:47] build_global_upgrade: build_global_upgrade
  7. [03:14] construct_entity: Wehrmacht Building (Structure #95)
  8. [03:24] select_battlegroup: Unknown Wehrmacht BG 1
  9. [03:25] select_battlegroup: select_battlegroup
 10. [03:26] build_global_upgrade: build_global_upgrade
 11. [04:52] build_global_upgrade: Medical Station (Wehrmacht)
 12. [05:13] construct_entity: Wehrmacht Building (Structure #525)

=== Player 1: Surgie ===
Build Order:
  1. [00:05] select_battlegroup: Armored (US)
  2. [00:14] build_squad: Engineer Squad
  3. [00:48] build_squad: M1 Mortar Team
  4. [01:45] build_squad: Riflemen Squad
  5. [02:54] build_squad: Riflemen Squad
  6. [07:14] build_squad: M3 Armored Personnel Carrier

=== Player 2: ftw ===
Build Order:
  1. [00:02] build_squad: Kradschützen Motorcycle Team
  2. [00:21] build_squad: Panzergrenadier Squad
  3. [01:28] build_squad: Panzergrenadier Squad
  4. [01:46] select_battlegroup: Unknown Afrika Korps BG
  5. [01:47] build_global_upgrade: T1 Unit Unlock (Afrika Korps)
  6. [02:28] construct_entity: AfrikaKorps Building (Structure #45)
  7. [03:04] build_squad: MG34 Team
  8. [03:56] build_squad: Panzerjäger Squad
  9. [05:19] build_squad: Assault Grenadier Squad
 10. [06:48] build_squad: Panzerpioneer Squad
```

## Architecture

This project uses Foreign Function Interface (FFI) to bridge between Go and Rust:

- **Rust Layer** (`vault-wrapper/`): Wraps the vault library and exposes C-compatible functions
- **Go Layer** (`vault/`): CGO bindings to call Rust functions from Go
- **CLI Layer** (`cmd/`): Command-line interface using Cobra

## Limitations

- Build order extraction depends on the parsing capabilities of the vault library
- Some battlegroup and upgrade names may show as generic text when PBGID mappings are not available
- Building names currently show as generic "Faction Building (Structure #X)" format
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