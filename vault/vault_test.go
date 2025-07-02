package vault

import (
	"testing"
	"path/filepath"
	"os"
)

// NOTE: For comprehensive end-to-end testing, see the tests/ directory
// which contains a complete testing framework with fixtures, CLI tests,
// benchmarks, and table-driven tests.

const testReplayFile = "../tests/testdata/temp_29_06_2025__22_49.rec"

func TestParseReplayFull_Integration(t *testing.T) {
	// Skip if test file doesn't exist
	if _, err := os.Stat(testReplayFile); os.IsNotExist(err) {
		t.Skipf("Test replay file not found: %s", testReplayFile)
	}

	absPath, err := filepath.Abs(testReplayFile)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	data, err := ParseReplayFull(absPath)
	if err != nil {
		t.Fatalf("Failed to parse replay: %v", err)
	}

	// Test basic structure
	if !data.Success {
		t.Errorf("Expected success=true, got false. Error: %v", data.ErrorMessage)
	}

	// Test expected duration: 38:27 = 38*60 + 27 = 2307 seconds
	expectedDuration := uint32(2307)
	if data.DurationSeconds != expectedDuration {
		t.Errorf("Expected duration %d seconds (38:27), got %d seconds (%02d:%02d)", 
			expectedDuration, data.DurationSeconds,
			data.DurationSeconds/60, data.DurationSeconds%60)
	}

	// Test that we have 4 players
	if len(data.Players) != 4 {
		t.Errorf("Expected 4 players, got %d", len(data.Players))
	}

	// Test that we have 2 teams
	if len(data.Teams) != 2 {
		t.Errorf("Expected 2 teams, got %d", len(data.Teams))
	}

	// Test that Tomsch is player 0
	var tomschPlayer *Player
	for _, player := range data.Players {
		if player.PlayerName == "Tomsch" {
			tomschPlayer = &player
			break
		}
	}

	if tomschPlayer == nil {
		t.Fatalf("Tomsch not found in players")
	}

	if tomschPlayer.PlayerID != 0 {
		t.Errorf("Expected Tomsch to be player ID 0, got %d", tomschPlayer.PlayerID)
	}

	// Test Tomsch has build commands
	if len(tomschPlayer.BuildCommands) == 0 {
		t.Fatalf("Tomsch has no build commands")
	}

	firstCommand := tomschPlayer.BuildCommands[0]
	
	// Test that command type is not empty
	if firstCommand.CommandType == "" {
		t.Errorf("First command has empty command type: %+v", firstCommand)
	}

	// Print actual data for debugging
	t.Logf("Actual duration: %d seconds (%02d:%02d)", 
		data.DurationSeconds, data.DurationSeconds/60, data.DurationSeconds%60)
	t.Logf("Tomsch's first command: %+v", firstCommand)
	t.Logf("Map name: %s", data.MapName)
	t.Logf("Winning team: %v", data.WinningTeam)
	
	if len(tomschPlayer.BuildCommands) >= 3 {
		t.Logf("First 3 build commands for Tomsch:")
		for i := 0; i < 3; i++ {
			cmd := tomschPlayer.BuildCommands[i]
			var unitName string
			if cmd.UnitName != nil {
				unitName = *cmd.UnitName
			} else {
				unitName = cmd.CommandType
			}
			t.Logf("  %d. [%02d:%02d] %s: %s", 
				i+1,
				cmd.Timestamp/60000, (cmd.Timestamp/1000)%60,
				cmd.CommandType, 
				unitName)
		}
	}

	// Test team structure
	for _, team := range data.Teams {
		if len(team.Players) == 0 {
			t.Errorf("Team %d has no players", team.TeamID)
		}
		
		for _, player := range team.Players {
			if player.PlayerName == "" {
				t.Errorf("Player in team %d has empty name", team.TeamID)
			}
		}
	}
}

func TestParseReplayFull_ErrorHandling(t *testing.T) {
	// Test with non-existent file
	_, err := ParseReplayFull("/nonexistent/file.rec")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	// Test with empty filename
	_, err = ParseReplayFull("")
	if err == nil {
		t.Error("Expected error for empty filename, got nil")
	}
}