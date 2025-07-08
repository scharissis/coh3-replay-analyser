package tests

import (
	"strings"
	"testing"

	"github.com/scharissis/coh3-replay-analyser/vault"
)

func TestBuildingNameLookupForConstructEntityCommands(t *testing.T) {
	// Test that construct_entity commands get appropriate building names
	replayPath := "testdata/2_07_2025__12_22_PM.rec"
	dataDir := "../data/coh3-data"
	
	filter := vault.NewBuildOnlyFilter()
	data, err := vault.ParseReplayWithFilter(replayPath, dataDir, filter)
	if err != nil {
		t.Fatalf("Failed to parse replay: %v", err)
	}
	
	// Test PCMD_PlaceAndConstructEntities commands (should have "Building Placement Command")
	found := false
	for _, player := range data.Players {
		if player.PlayerName == "ftw" {
			for _, cmd := range player.Commands {
				if cmd.CommandType == "construct_entity" {
					found = true
					if cmd.BuildingName == nil {
						t.Errorf("Expected building name for ftw's construct_entity command, got nil")
					} else if *cmd.BuildingName != "Light Support Kompanie" {
						t.Errorf("Expected 'Light Support Kompanie' for ftw's construct_entity command, got '%s'", *cmd.BuildingName)
					}
					t.Logf("✅ ftw construct_entity: %s", *cmd.BuildingName)
					break // Only test first one
				}
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find construct_entity command for ftw")
	}
	
	// Test SCMD_BuildStructure commands (should have "Building Construction Command (Unknown Type)")
	found = false
	for _, player := range data.Players {
		if player.PlayerName == "Surgie" {
			for _, cmd := range player.Commands {
				if cmd.CommandType == "construct_entity" {
					// Look for the specific SCMD_BuildStructure at ~12:31
					timestampSeconds := cmd.Timestamp / 1000
					if timestampSeconds >= 750 && timestampSeconds <= 752 {
						found = true
						if cmd.BuildingName == nil {
							t.Errorf("Expected building name for Surgie's SCMD_BuildStructure command, got nil")
						} else if !strings.Contains(*cmd.BuildingName, "Americans Building (Structure #182)") {
							t.Errorf("Expected 'Americans Building (Structure #182)' for Surgie's SCMD_BuildStructure command, got '%s'", *cmd.BuildingName)
						}
						t.Logf("✅ Surgie SCMD_BuildStructure: %s", *cmd.BuildingName)
						break
					}
				}
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find SCMD_BuildStructure command for Surgie around 12:31")
	}
}

func TestConstructEntityCommandsHaveDescriptiveNames(t *testing.T) {
	// Test that all construct_entity commands have non-generic names
	replayPath := "testdata/2_07_2025__12_22_PM.rec"
	dataDir := "../data/coh3-data"
	
	filter := vault.NewBuildOnlyFilter()
	data, err := vault.ParseReplayWithFilter(replayPath, dataDir, filter)
	if err != nil {
		t.Fatalf("Failed to parse replay: %v", err)
	}
	
	constructEntityCount := 0
	descriptiveNameCount := 0
	
	for _, player := range data.Players {
		for _, cmd := range player.Commands {
			if cmd.CommandType == "construct_entity" {
				constructEntityCount++
				
				if cmd.BuildingName != nil {
					buildingName := *cmd.BuildingName
					// Check that we're not using the old generic name
					if strings.Contains(buildingName, "Building") {
						descriptiveNameCount++
						t.Logf("✅ %s: %s", player.PlayerName, buildingName)
					} else {
						t.Errorf("Unexpected building name format for %s: %s", player.PlayerName, buildingName)
					}
				} else {
					t.Errorf("construct_entity command for %s has nil building name", player.PlayerName)
				}
			}
		}
	}
	
	if constructEntityCount == 0 {
		t.Error("Expected to find construct_entity commands")
	}
	
	if descriptiveNameCount != constructEntityCount {
		t.Errorf("Expected all %d construct_entity commands to have descriptive names, got %d", 
			constructEntityCount, descriptiveNameCount)
	}
	
	t.Logf("Successfully verified %d construct_entity commands with descriptive names", descriptiveNameCount)
}