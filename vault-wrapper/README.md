# Vault Wrapper

This directory contains a Rust wrapper around the [vault](https://crates.io/crates/vault) library for parsing Company of Heroes 3 replay files.

## Purpose

The vault-wrapper provides:

1. **C FFI Library** (`libvault_wrapper.a`) - A static library that exposes vault functionality for use in Go applications
2. **Debug Binary** (`vault-debug`) - A standalone tool for examining raw vault library output

## Components

### FFI Library (lib.rs)

The main library (`lib.rs`) provides a C-compatible interface to the vault library:

- `parse_replay_full()` - Parses a replay file and returns comprehensive JSON data including:
  - Match metadata (map, duration, players, teams)
  - Player commands and build orders  
  - Chat messages
  - **Enhanced command parsing** with unit/building names hydrated from lookup data
  - Accurate timestamp conversion using tick rate constants

The library is used by the Go application via CGO bindings.

### Game Data Lookup System

The library includes an intelligent lookup system for converting game codes to friendly names:

- **JSON-based data files** (`data/units.json`, `data/buildings.json`, `data/abilities.json`)
- **Automatic fallback** to hardcoded data if JSON files are unavailable
- **PBGID resolution** - converts numeric game IDs to unit/building names
- **Faction-aware** - includes faction information for each unit
- **Offline operation** - all data is bundled with the application

Currently supports common units from:
- Wehrmacht (Grenadier Squad, Pioneer Squad, MG42, Panzer IV, etc.)
- Afrika Korps (Panzergrenadier Squad, 8 Rad Armored Car, etc.)
- British Forces (Section, Commando Section, Churchill Tank, etc.)
- US Forces (Rifleman Squad, Engineer Squad, Sherman Tank, etc.)

### Debug Binary (main.rs)

A debugging tool that shows the raw, unprocessed output from the vault library. This is useful for:

- Understanding what data the vault library actually provides
- Debugging issues with command parsing
- Seeing the exact structure of vault's debug output
- Comparing raw vault output with the processed JSON from the FFI library

## Usage

### Building

```bash
# Build the static library (for Go integration)
cargo build --release

# Build the debug binary
cargo build --bin vault-debug
```

### Running the Debug Tool

```bash
# Run on a replay file
./target/debug/vault-debug path/to/replay.rec

# Example with test data
./target/debug/vault-debug ../tests/testdata/temp_29_06_2025__22_49.rec
```

### Debug Output

The debug tool shows:

- **Basic Info**: Version, duration, timestamp, game type
- **Map Details**: Filename and localized name IDs
- **Player Data**: Names, factions, teams, Steam IDs, battlegroups
- **Commands**: First 5 commands per player with full debug structure
- **Messages**: Chat messages with timestamps
- **Full Debug**: Truncated view of the complete vault data structure

## Timing and Tick Rate

The library uses a configurable tick rate constant:
- **SECONDS_PER_TICK**: 0.125 seconds (each tick represents 1/8th of a second)
- **TICKS_PER_SECOND**: 8 ticks per second
- **MILLISECONDS_PER_TICK**: 125 milliseconds per tick

This allows accurate timestamp conversion from vault's internal tick-based timing to real-world time units.

## Dependencies

- `vault = "10.1.5"` - The core CoH3 replay parsing library
- `serde` + `serde_json` - JSON serialization for the FFI interface
- `libc` - C interop for the FFI layer

## Updating Lookup Data

See [LOOKUP_DATA.md](LOOKUP_DATA.md) for detailed instructions on:
- Adding new unit/building/ability definitions
- Discovering missing PBGIDs from replays
- Maintaining the JSON data files
- Troubleshooting lookup issues

## Integration

The static library (`libvault_wrapper.a`) is consumed by the Go application in the parent directory via CGO. The Go code calls `parse_replay_full()` and receives structured JSON data about the replay.