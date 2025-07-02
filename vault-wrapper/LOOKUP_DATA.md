# Game Data Lookup System

## Overview

The vault-wrapper includes a sophisticated lookup system that converts numeric game codes (PBGIDs) to human-readable unit, building, and ability names. This system supports both JSON-based data files and hardcoded fallbacks.

## Data Storage

### JSON Files (Recommended)

Data is stored in JSON files in the `data/` directory:

- `data/units.json` - Unit definitions (infantry, vehicles, etc.)
- `data/buildings.json` - Building definitions (structures, fortifications)
- `data/abilities.json` - Ability definitions (spells, upgrades)

### File Structure

Each JSON file maps PBGID strings to data objects:

```json
{
  "198355": {
    "name": "Grenadier Squad",
    "faction": "Wehrmacht",
    "category": "Infantry",
    "description": "Basic infantry squad"
  },
  "170321": {
    "name": "Sherman Tank",
    "faction": "US Forces", 
    "category": "Vehicle",
    "description": "Medium tank"
  }
}
```

## How to Update Lookup Data

### Method 1: Adding New Units to JSON Files

1. **Identify Missing PBGIDs**:
   ```bash
   # Use the debug binary to see raw PBGIDs
   cd vault-wrapper
   ./target/debug/vault-debug ../tests/testdata/your_replay.rec | grep "pbgid:"
   ```

2. **Find Unknown Units**:
   ```bash
   # Look for generic fallback names in the output
   ./coh3-build-order build-order tests/testdata/your_replay.rec | grep "Infantry Squad\|Building\|Ability"
   ```

3. **Add to JSON file**:
   Edit `data/units.json` and add the new PBGID:
   ```json
   {
     "existing_units": "...",
     "NEW_PBGID": {
       "name": "Unit Name",
       "faction": "Faction Name",
       "category": "Infantry|Vehicle|Support|Engineer",
       "description": "Brief description"
     }
   }
   ```

4. **Rebuild and test**:
   ```bash
   make build
   ./coh3-build-order build-order tests/testdata/your_replay.rec
   ```

### Method 2: Automated Data Discovery

For systematic discovery of new units:

1. **Extract all PBGIDs from multiple replays**:
   ```bash
   # Create a script to extract PBGIDs from many replays
   for replay in tests/testdata/*.rec; do
     echo "=== $replay ==="
     cd vault-wrapper
     ./target/debug/vault-debug ../$replay | grep "pbgid:" | sort -u
     cd ..
   done > all_pbgids.txt
   ```

2. **Compare with existing data**:
   ```bash
   # Find PBGIDs not in our JSON files
   grep "pbgid:" all_pbgids.txt | sed 's/.*pbgid: \([0-9]*\).*/\1/' | sort -u > discovered_pbgids.txt
   jq -r 'keys[]' data/units.json | sort > known_pbgids.txt
   comm -23 discovered_pbgids.txt known_pbgids.txt > missing_pbgids.txt
   ```

### Method 3: Using External Data Sources

1. **From coh3-data repository**:
   - Monitor https://github.com/cohstats/coh3-data for updates
   - Download latest unit data: `https://data.coh3stats.com/cohstats/coh3-data/latest/data/`
   - Convert their JSON format to our lookup format

2. **Manual research**:
   - Use community wikis and databases
   - Test in-game to verify unit names
   - Cross-reference with other CoH3 tools

## Fallback System

The lookup system has multiple layers:

1. **JSON files** (primary) - loaded from `data/` directory
2. **Hardcoded data** (fallback) - embedded in `src/lookup.rs`
3. **Heuristic parsing** (last resort) - pattern matching on debug output

### Updating Hardcoded Fallbacks

Edit `vault-wrapper/src/lookup.rs` in the `initialize_hardcoded_data()` function:

```rust
self.units.insert(NEW_PBGID, UnitData {
    name: "Unit Name".to_string(),
    faction: Some("Faction".to_string()),
    category: "Category".to_string(),
    description: Some("Description".to_string()),
});
```

## Data Quality Guidelines

### Naming Conventions

- **Units**: Use official in-game names (e.g., "Grenadier Squad", "Sherman Tank")
- **Factions**: Use standard names ("Wehrmacht", "US Forces", "British", "Afrika Korps")
- **Categories**: Use consistent categories ("Infantry", "Vehicle", "Support", "Engineer")
- **Descriptions**: Keep brief and descriptive

### Validation

Before adding new data:

1. **Verify accuracy**: Test with actual replays
2. **Check for duplicates**: Ensure PBGID isn't already defined
3. **Validate JSON**: Use `jq` to check syntax
4. **Test thoroughly**: Run on multiple replay files

## Deployment

### Local Development

1. Edit JSON files in `data/`
2. Run `make build`
3. Test with replay files

### Production Updates

1. Update JSON files in the repository
2. Rebuild the static library: `make build-rust`
3. Rebuild the Go binary: `make build-go`
4. Deploy the updated `coh3-build-order` binary

## Troubleshooting

### Common Issues

1. **JSON syntax errors**: Use `jq . data/units.json` to validate
2. **Missing file**: Library falls back to hardcoded data automatically
3. **Wrong directory**: JSON files must be in `data/` relative to binary execution
4. **PBGID parsing**: Ensure PBGID strings are quoted in JSON

### Debug Commands

```bash
# Check if JSON files are loading
./coh3-build-order build-order tests/testdata/replay.rec | grep "Infantry Squad"

# View raw vault output
cd vault-wrapper
./target/debug/vault-debug ../tests/testdata/replay.rec | less

# Validate JSON syntax
jq . data/units.json > /dev/null && echo "Valid JSON" || echo "Invalid JSON"
```

## Future Enhancements

1. **Automatic data fetching**: Download latest data from coh3-data repository
2. **Version management**: Track data version compatibility
3. **Localization**: Support multiple languages
4. **Dynamic updates**: Hot-reload data without rebuilding
5. **AI-assisted naming**: Use LLMs to suggest unit names from descriptions