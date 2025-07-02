package tests

import (
	"testing"

	"github.com/scharissis/coh3-replay-analyser/vault"
)

// TestReplayFiles contains test cases for individual replay files
func TestReplayFiles(t *testing.T) {
	testCases := []struct {
		name        string
		replayFile  string
		fixtureFile string
	}{
		{
			name:        "Rails and Sand 4P Replay",
			replayFile:  "temp_29_06_2025__22_49.rec",
			fixtureFile: "temp_29_06_2025__22_49.json",
		},
		// Add more test cases here as you add more replay files
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Load test fixture
			fixture, err := LoadTestFixture(tc.fixtureFile)
			if err != nil {
				t.Fatalf("Failed to load test fixture: %v", err)
			}

			// Get replay file path
			replayPath, err := GetTestDataPath(tc.replayFile)
			if err != nil {
				t.Fatalf("Failed to get replay path: %v", err)
			}

			// Parse the replay
			data, err := vault.ParseReplayFull(replayPath)
			if err != nil {
				t.Fatalf("Failed to parse replay: %v", err)
			}

			// Validate against expected results
			ValidateFullReplay(t, data, fixture.ExpectedResults)

			// Log detailed stats for debugging
			LogReplayStats(t, data)
		})
	}
}

// TestParseReplayFull_Comprehensive performs detailed testing of full replay parsing
func TestParseReplayFull_Comprehensive(t *testing.T) {
	replayPath, err := GetTestDataPath("temp_29_06_2025__22_49.rec")
	if err != nil {
		t.Skipf("Test replay file not found: %v", err)
	}

	data, err := vault.ParseReplayFull(replayPath)
	if err != nil {
		t.Fatalf("Failed to parse replay: %v", err)
	}

	// Test basic structure
	if !data.Success {
		t.Errorf("Expected success=true, got false. Error: %v", data.ErrorMessage)
	}

	// Test expected duration: 38:27 = 38*60 + 27 = 2307 seconds
	AssertDuration(t, data.DurationSeconds, "38:27")

	// Test that we have exactly 4 players
	if len(data.Players) != 4 {
		t.Errorf("Expected 4 players, got %d", len(data.Players))
	}

	// Test that we have exactly 2 teams
	if len(data.Teams) != 2 {
		t.Errorf("Expected 2 teams, got %d", len(data.Teams))
	}

	// Verify Tomsch is player 0 with specific expectations
	expectedTomsch := ExpectedPlayer{
		ID:            0,
		Name:          "Tomsch",
		TeamID:        1,
		MinCommands:   5,
		FirstCommands: []string{"build_squad", "build_squad", "build_squad"},
	}
	tomsch := AssertPlayerExists(t, data, expectedTomsch)
	AssertCommandCount(t, tomsch, expectedTomsch)
	AssertFirstCommands(t, tomsch, expectedTomsch.FirstCommands)


	// Verify all players have valid team assignments
	for _, player := range data.Players {
		if player.TeamID != 1 && player.TeamID != 2 {
			t.Errorf("Player %s has invalid team ID: %d (expected 1 or 2)",
				player.PlayerName, player.TeamID)
		}
	}

	// Verify teams contain correct players
	var team1Players, team2Players []string
	for _, team := range data.Teams {
		for _, player := range team.Players {
			if team.TeamID == 1 {
				team1Players = append(team1Players, player.PlayerName)
			} else if team.TeamID == 2 {
				team2Players = append(team2Players, player.PlayerName)
			}
		}
	}

	if len(team1Players) != 2 || len(team2Players) != 2 {
		t.Errorf("Expected 2 players per team, got team1: %d, team2: %d",
			len(team1Players), len(team2Players))
	}
}

// TestAllReplaysInTestData ensures all replay files in testdata can be parsed
func TestAllReplaysInTestData(t *testing.T) {
	replays, err := GetAllTestReplays()
	if err != nil {
		t.Fatalf("Failed to get test replays: %v", err)
	}

	if len(replays) == 0 {
		t.Skip("No replay files found in testdata directory")
	}

	for _, replayFile := range replays {
		t.Run(replayFile, func(t *testing.T) {
			replayPath, err := GetTestDataPath(replayFile)
			if err != nil {
				t.Fatalf("Failed to get path for %s: %v", replayFile, err)
			}

			data, err := vault.ParseReplayFull(replayPath)
			if err != nil {
				t.Errorf("Failed to parse %s: %v", replayFile, err)
				return
			}

			// Basic validation for all replays
			if !data.Success {
				t.Errorf("Parsing %s failed: %v", replayFile, data.ErrorMessage)
			}

			if data.DurationSeconds == 0 {
				t.Errorf("Replay %s has zero duration", replayFile)
			}

			if len(data.Players) == 0 {
				t.Errorf("Replay %s has no players", replayFile)
			}

			// Log basic stats for debugging
			t.Logf("✓ %s: %s, %d players, %d teams",
				replayFile, FormatDuration(data.DurationSeconds), len(data.Players), len(data.Teams))
		})
	}
}

// TestErrorHandling tests various error conditions
func TestErrorHandling(t *testing.T) {
	t.Run("NonExistentFile", func(t *testing.T) {
		_, err := vault.ParseReplayFull("/nonexistent/file.rec")
		if err == nil {
			t.Error("Expected error for nonexistent file, got nil")
		}
	})

	t.Run("EmptyFilename", func(t *testing.T) {
		_, err := vault.ParseReplayFull("")
		if err == nil {
			t.Error("Expected error for empty filename, got nil")
		}
	})
}
