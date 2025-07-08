package entity

import (
	"fmt"
	"strings"
)

// EntityTracker tracks entities and infers building types from production patterns
type EntityTracker struct {
	entities map[string]*TrackedEntity
	// Known unit-to-building mappings for different factions
	unitToBuildingMap map[string]map[string]string
}

// TrackedEntity represents a single entity (building or unit) and its activity
type TrackedEntity struct {
	Index               string
	FirstSeenTimestamp  uint32
	LastSeenTimestamp   uint32
	CommandHistory      []EntityCommand
	InferredBuildingID  *string
	InferredBuildingName *string
	Faction             string
}

// EntityCommand represents a command associated with an entity
type EntityCommand struct {
	Timestamp   uint32
	CommandType string
	PBGID       *string
	Details     string
}

// NewEntityTracker creates a new entity tracker
func NewEntityTracker() *EntityTracker {
	return &EntityTracker{
		entities:          make(map[string]*TrackedEntity),
		unitToBuildingMap: initializeUnitToBuildingMap(),
	}
}

// Command represents a command for entity tracking (avoiding import cycle)
type Command struct {
	Timestamp   uint32
	CommandType string
	Details     string
	PBGID       *string
	Index       *string
}

// TrackCommand processes a command and updates entity tracking
func (et *EntityTracker) TrackCommand(cmd Command, playerFaction string) {
	if cmd.Index == nil {
		return
	}

	index := *cmd.Index
	
	// Initialize entity if not exists
	if et.entities[index] == nil {
		et.entities[index] = &TrackedEntity{
			Index:              index,
			FirstSeenTimestamp: cmd.Timestamp,
			LastSeenTimestamp:  cmd.Timestamp,
			CommandHistory:     []EntityCommand{},
			Faction:            playerFaction,
		}
	}

	entity := et.entities[index]
	entity.LastSeenTimestamp = cmd.Timestamp

	// Add command to history
	entityCmd := EntityCommand{
		Timestamp:   cmd.Timestamp,
		CommandType: cmd.CommandType,
		PBGID:       cmd.PBGID,
		Details:     cmd.Details,
	}
	entity.CommandHistory = append(entity.CommandHistory, entityCmd)

	// Try to infer building type
	et.inferBuildingType(entity)
}

// inferBuildingType attempts to determine what type of building this entity is
func (et *EntityTracker) inferBuildingType(entity *TrackedEntity) {
	if entity.InferredBuildingID != nil {
		return // Already inferred
	}

	// Look for construct_entity commands (indicates this is a building)
	hasConstruct := false
	for _, cmd := range entity.CommandHistory {
		if cmd.CommandType == "construct_entity" {
			hasConstruct = true
			break
		}
	}

	if !hasConstruct {
		return // Not a building
	}

	// Look for build_squad commands that might indicate building type
	for _, cmd := range entity.CommandHistory {
		if cmd.CommandType == "build_squad" && cmd.PBGID != nil {
			if buildingInfo := et.inferBuildingFromUnit(*cmd.PBGID, entity.Faction); buildingInfo != nil {
				entity.InferredBuildingID = &buildingInfo.ID
				entity.InferredBuildingName = &buildingInfo.Name
				return
			}
		}
	}

	// If no direct production, look for units produced shortly after construction
	// This handles cases where units are produced by index X but constructed at index Y
	constructTimestamp := et.getConstructTimestamp(entity)
	if constructTimestamp == 0 {
		return
	}

	// Look for units produced within 1 minute of construction (after only)
	timeWindow := uint32(1 * 60 * 1000) // 1 minute in milliseconds
	
	// Check all entities for unit production in the time window (before or after construction)
	var candidateUnits []string
	for _, otherEntity := range et.entities {
		for _, cmd := range otherEntity.CommandHistory {
			if cmd.CommandType == "build_squad" && cmd.PBGID != nil {
				// Look for units produced within time window AFTER construction only
				if cmd.Timestamp >= constructTimestamp && 
				   cmd.Timestamp <= constructTimestamp + timeWindow {
					candidateUnits = append(candidateUnits, *cmd.PBGID)
				}
			}
		}
	}

	// Try to infer building type from the most likely candidate unit
	for _, unitPBGID := range candidateUnits {
		if buildingInfo := et.inferBuildingFromUnit(unitPBGID, entity.Faction); buildingInfo != nil {
			// Skip HQ inference for constructed buildings (HQ is a starting building)
			if buildingInfo.ID != "HQ" {
				entity.InferredBuildingID = &buildingInfo.ID
				entity.InferredBuildingName = &buildingInfo.Name
				return
			}
		}
	}
}

// BuildingInfo represents information about a building type
type BuildingInfo struct {
	ID   string
	Name string
}

// inferBuildingFromUnit attempts to determine building type from a unit PBGID
func (et *EntityTracker) inferBuildingFromUnit(unitPBGID, faction string) *BuildingInfo {
	factionMap, exists := et.unitToBuildingMap[strings.ToLower(faction)]
	if !exists {
		return nil
	}

	buildingID, exists := factionMap[unitPBGID]
	if !exists {
		return nil
	}

	// Map building IDs to names
	buildingNames := map[string]string{
		"HQ": "Headquarters",
		"198236": "Light Support Kompanie",
		"198237": "Mechanized Kompanie", 
		"UNKNOWN": "Unknown Building Type",
	}

	name, exists := buildingNames[buildingID]
	if !exists {
		name = fmt.Sprintf("Unknown Building (ID: %s)", buildingID)
	}

	return &BuildingInfo{
		ID:   buildingID,
		Name: name,
	}
}

// getConstructTimestamp returns the timestamp when this entity was constructed
func (et *EntityTracker) getConstructTimestamp(entity *TrackedEntity) uint32 {
	for _, cmd := range entity.CommandHistory {
		if cmd.CommandType == "construct_entity" {
			return cmd.Timestamp
		}
	}
	return 0
}

// GetTrackedEntities returns all tracked entities
func (et *EntityTracker) GetTrackedEntities() map[string]*TrackedEntity {
	return et.entities
}

// GetBuildings returns only entities that are identified as buildings
func (et *EntityTracker) GetBuildings() []*TrackedEntity {
	var buildings []*TrackedEntity
	
	for _, entity := range et.entities {
		hasConstruct := false
		for _, cmd := range entity.CommandHistory {
			if cmd.CommandType == "construct_entity" {
				hasConstruct = true
				break
			}
		}
		
		if hasConstruct {
			buildings = append(buildings, entity)
		}
	}
	
	return buildings
}

// FinalizeTracking performs final analysis after all commands are processed
func (et *EntityTracker) FinalizeTracking() {
	// Re-run inference for all entities to catch cross-entity patterns
	for _, entity := range et.entities {
		et.inferBuildingType(entity)
	}
}

// FormatTimestamp formats a timestamp for display
func (et *EntityTracker) FormatTimestamp(timestamp uint32) string {
	seconds := timestamp / 1000
	minutes := seconds / 60
	remainingSeconds := seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, remainingSeconds)
}

// initializeUnitToBuildingMap creates mappings from unit PBGIDs to building PBGIDs
func initializeUnitToBuildingMap() map[string]map[string]string {
	return map[string]map[string]string{
		"afrikakorps": {
			// Headquarters (HQ) - Starting building - PBGID TBD
			"198340": "HQ", // Panzergrenadier Squad
			"198341": "HQ", // Panzerpioneer Squad  
			"198355": "HQ", // Kradschützen Motorcycle Team
			
			// Light Support Kompanie - 198236
			"198347": "198236", // MG34 Machine Gun Team
			"198342": "198236", // Panzerjäger Squad
			"2072237": "198236", // 2.5-tonne Medical Truck
			"2063111": "198236", // Flakvierling Half-track
			
			// Mechanized Kompanie - 198237
			"2033664": "198237", // StuG III D Assault Gun
			"198357": "198237", // Marder III Tank Destroyer
			"198361": "198237", // Panzer III Medium Tank
			
			// Special units (need to identify building)
			"198413": "UNKNOWN", // Walking Stuka Rocket Launcher
		},
		"wehrmacht": {
			// Similar mapping for Wehrmacht
		},
		"americans": {
			// Similar mapping for Americans
		},
		"british": {
			// Similar mapping for British
		},
	}
}