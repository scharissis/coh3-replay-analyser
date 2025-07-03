package tests

import (
	"testing"

	"github.com/scharissis/coh3-replay-analyser/vault"
)

func TestSCMDBuildStructureCommandParsing(t *testing.T) {
	// Test that SCMD_BuildStructure commands are parsed as construct_entity
	replayPath := "testdata/2_07_2025__12_22_PM.rec"
	dataDir := "../data/coh3-data"
	
	// Parse the replay with all commands filter to ensure construct_entity commands are included
	filter := vault.NewAllCommandsFilter()
	data, err := vault.ParseReplayWithFilter(replayPath, dataDir, filter)
	if err != nil {
		t.Fatalf("Failed to parse replay: %v", err)
	}
	
	// Find Surgie (player 1) who should have the SCMD_BuildStructure command
	var surgiePlayer *vault.Player
	for _, player := range data.Players {
		if player.PlayerName == "Surgie" {
			surgiePlayer = &player
			break
		}
	}
	
	if surgiePlayer == nil {
		t.Fatal("Could not find player 'Surgie' in replay data")
	}
	
	// Look for construct_entity commands around tick 6010 (12:30-12:31 game time)
	var foundConstructEntity bool
	var constructEntityCommand vault.Command
	
	for _, cmd := range surgiePlayer.Commands {
		if cmd.CommandType == "construct_entity" {
			// Check if this command is around the expected time (12:30-12:31 = 750-751 seconds)
			timestampSeconds := cmd.Timestamp / 1000
			if timestampSeconds >= 750 && timestampSeconds <= 752 {
				foundConstructEntity = true
				constructEntityCommand = cmd
				break
			}
		}
	}
	
	if !foundConstructEntity {
		t.Error("Expected to find construct_entity command around 12:30-12:31 for Surgie, but none found")
		
		// Debug: List all construct_entity commands for Surgie
		t.Log("All construct_entity commands for Surgie:")
		for _, cmd := range surgiePlayer.Commands {
			if cmd.CommandType == "construct_entity" {
				timestampSeconds := cmd.Timestamp / 1000
				minutes := timestampSeconds / 60
				seconds := timestampSeconds % 60
				pbgid := "nil"
				if cmd.PBGID != nil {
					pbgid = *cmd.PBGID
				}
				t.Logf("  [%02d:%02d] %s: %s (pbgid: %s)", 
					minutes, seconds, cmd.CommandType, cmd.Details, pbgid)
			}
		}
		return
	}
	
	// Verify the command properties
	expectedPbgid := "182"
	if constructEntityCommand.PBGID == nil || *constructEntityCommand.PBGID != expectedPbgid {
		actualPbgid := "nil"
		if constructEntityCommand.PBGID != nil {
			actualPbgid = *constructEntityCommand.PBGID
		}
		t.Errorf("Expected pbgid '%s', got '%s'", expectedPbgid, actualPbgid)
	}
	
	// Log success
	timestampSeconds := constructEntityCommand.Timestamp / 1000
	minutes := timestampSeconds / 60
	seconds := timestampSeconds % 60
	pbgid := "nil"
	if constructEntityCommand.PBGID != nil {
		pbgid = *constructEntityCommand.PBGID
	}
	t.Logf("Successfully found SCMD_BuildStructure parsed as construct_entity at [%02d:%02d] with pbgid %s", 
		minutes, seconds, pbgid)
}

func TestConstructEntityCommandsIncludedInBuildOrder(t *testing.T) {
	// Test that construct_entity commands appear in default build order filter
	replayPath := "testdata/2_07_2025__12_22_PM.rec"
	dataDir := "../data/coh3-data"
	
	// Use default build filter (should include construct_entity commands)
	filter := vault.NewBuildOnlyFilter()
	data, err := vault.ParseReplayWithFilter(replayPath, dataDir, filter)
	if err != nil {
		t.Fatalf("Failed to parse replay: %v", err)
	}
	
	// Count construct_entity commands across all players
	constructEntityCount := 0
	for _, player := range data.Players {
		for _, cmd := range player.Commands {
			if cmd.CommandType == "construct_entity" {
				constructEntityCount++
			}
		}
	}
	
	if constructEntityCount == 0 {
		t.Error("Expected to find construct_entity commands in build order, but found none")
	} else {
		t.Logf("Found %d construct_entity commands in build order across all players", constructEntityCount)
	}
}

func TestPbgidExtractionForSCMDCommands(t *testing.T) {
	// Test that SCMD commands have their index field extracted as pbgid
	replayPath := "testdata/2_07_2025__12_22_PM.rec"
	dataDir := "../data/coh3-data"
	
	filter := vault.NewAllCommandsFilter()
	data, err := vault.ParseReplayWithFilter(replayPath, dataDir, filter)
	if err != nil {
		t.Fatalf("Failed to parse replay: %v", err)
	}
	
	// Find the specific SCMD_BuildStructure command we know exists
	var foundCommand bool
	for _, player := range data.Players {
		if player.PlayerName != "Surgie" {
			continue
		}
		
		for _, cmd := range player.Commands {
			if cmd.CommandType == "construct_entity" && cmd.PBGID != nil && *cmd.PBGID == "182" {
				// Check if timestamp matches our expected tick 6010
				expectedTimestamp := uint32(6010 * 125) // 6010 ticks * 125ms per tick
				if cmd.Timestamp >= expectedTimestamp-500 && cmd.Timestamp <= expectedTimestamp+500 {
					foundCommand = true
					t.Logf("Found SCMD_BuildStructure command with correctly extracted pbgid: %s", *cmd.PBGID)
					break
				}
			}
		}
	}
	
	if !foundCommand {
		t.Error("Expected to find SCMD_BuildStructure command with pbgid '182', but none found")
	}
}