# Building Name Resolution - Implementation Guide

## Quick Reference

**Problem**: SCMD_BuildStructure commands show "Americans Building" instead of actual building names like "Infanterie Kompanie"

**Root Cause**: SCMD commands use `index` fields (e.g., `index: 182`) instead of `pbgid` fields for entity references

**Current Parser**: Only extracts `pbgid:` patterns, fails on SCMD commands that use `index:` patterns

## Immediate Fix Implementation

### 1. Update Rust Command Structure

**File**: `/vault-wrapper/src/lib.rs`

Add index field to Command struct:
```rust
#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct Command {
    pub timestamp: u32,
    pub command_type: String,
    pub details: String,
    pub pbgid: Option<String>,
    pub index: Option<String>,     // NEW: For SCMD commands
    pub unit_name: Option<String>,
    pub building_name: Option<String>,
}
```

### 2. Add Index Extraction Function

**File**: `/vault-wrapper/src/lib.rs`

```rust
fn extract_index(command_debug: &str) -> Option<String> {
    if let Some(start) = command_debug.find("index: ") {
        let substring = &command_debug[start + 7..];
        if let Some(end) = substring.find(',').or_else(|| substring.find('}')) {
            return Some(substring[..end].trim().to_string());
        }
    }
    None
}
```

### 3. Update Command Parsing

**File**: `/vault-wrapper/src/lib.rs` in `parse_command_simple()`

```rust
fn parse_command_simple(command_debug: &str) -> (String, String, Option<String>, Option<String>) {
    // ... existing logic ...
    
    let pbgid = extract_pbgid(command_debug);
    let index = extract_index(command_debug);   // NEW
    
    (command_type, details, pbgid, index)       // Updated return
}
```

### 4. Update Go Command Structure

**File**: `/vault/vault.go`

```go
type Command struct {
    Timestamp    uint32  `json:"timestamp"`
    CommandType  string  `json:"command_type"`
    Details      string  `json:"details"`
    PBGID        *string `json:"pbgid,omitempty"`
    Index        *string `json:"index,omitempty"`      // NEW
    UnitName     *string `json:"unit_name,omitempty"`
    BuildingName *string `json:"building_name,omitempty"`
}
```

### 5. Update Building Name Logic

**File**: `/vault/vault.go` in `enhanceCommandsWithPlayerInfo()`

```go
case "construct_entity":
    if cmd.PBGID != nil {
        // Try PBGID lookup first (for PCMD commands)
        if pbgid, err := strconv.ParseUint(*cmd.PBGID, 10, 32); err == nil {
            if unitInfo, err := resolver.ResolvePBGID(uint32(pbgid)); err == nil {
                cmd.BuildingName = &unitInfo.Name
                continue
            }
        }
    }
    
    // Fallback for SCMD commands with index
    if cmd.Index != nil {
        buildingName := *player.Faction + " Building (Structure #" + *cmd.Index + ")"
        cmd.BuildingName = &buildingName
    } else {
        buildingName := *player.Faction + " Building"
        cmd.BuildingName = &buildingName
    }
```

## Test Updates

Update tests to expect: `"Americans Building (Structure #182)"` instead of `"Americans Building"`

## Expected Results

**Before**: 
```
29. [12:31] construct_entity: Americans Building
```

**After**:
```  
29. [12:31] construct_entity: Americans Building (Structure #182)
```

## Complete Solution (Future Work)

For actual building names, implement entity tracking system:
1. Parse entire replay to track entity creation events
2. Map entity indices to building PBGIDs when entities are spawned
3. Resolve SCMD commands to actual building names

**Reference**: See `/BUILDING_NAME_INVESTIGATION.md` for detailed analysis and entity lifecycle patterns.