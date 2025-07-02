package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/scharissis/coh3-replay-analyser/vault"
)

// TestReplay represents a replay file and its expected results
type TestReplay struct {
	FileName        string             `json:"file_name"`
	ExpectedResults ExpectedReplayData `json:"expected_results"`
}

// ExpectedReplayData contains expected values for replay validation
type ExpectedReplayData struct {
	Duration        string           `json:"duration"` // Format: "MM:SS"
	DurationSeconds uint32           `json:"duration_seconds"`
	MapName         string           `json:"map_name"`
	PlayerCount     int              `json:"player_count"`
	TeamCount       int              `json:"team_count"`
	WinningTeam     *uint32          `json:"winning_team"`
	Players         []ExpectedPlayer `json:"players"`
}

// ExpectedPlayer contains expected data for a specific player
type ExpectedPlayer struct {
	ID            uint32   `json:"id"`
	Name          string   `json:"name"`
	TeamID        uint32   `json:"team_id"`
	MinCommands   int      `json:"min_commands"`   // Minimum number of commands expected
	MaxCommands   int      `json:"max_commands"`   // Maximum number of commands expected (0 = no limit)
	FirstCommands []string `json:"first_commands"` // Expected first few command types
}

// GetTestDataPath returns the absolute path to a test data file
func GetTestDataPath(filename string) (string, error) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Try different possible paths
	paths := []string{
		filepath.Join(wd, "tests", "testdata", filename),
		filepath.Join(wd, "testdata", filename),
		filepath.Join(filepath.Dir(wd), "tests", "testdata", filename),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return filepath.Abs(path)
		}
	}

	return "", fmt.Errorf("test data file not found: %s (searched in %v)", filename, paths)
}

// LoadTestFixture loads expected test results from a JSON fixture file
func LoadTestFixture(fixtureName string) (*TestReplay, error) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Try different possible paths for the fixture file
	paths := []string{
		filepath.Join(wd, "fixtures", fixtureName),                        // tests/fixtures/file.json (when run from tests/)
		filepath.Join(wd, "tests", "fixtures", fixtureName),               // tests/fixtures/file.json (when run from parent)
		filepath.Join(filepath.Dir(wd), "tests", "fixtures", fixtureName), // when run from different dir
	}

	var data []byte
	var lastErr error

	for _, path := range paths {
		data, lastErr = os.ReadFile(path)
		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to read fixture file %s (searched in %v): %w", fixtureName, paths, lastErr)
	}

	var testReplay TestReplay
	if err := json.Unmarshal(data, &testReplay); err != nil {
		return nil, fmt.Errorf("failed to parse fixture file %s: %w", fixtureName, err)
	}

	return &testReplay, nil
}

// ParseDuration parses a MM:SS duration string to seconds
func ParseDuration(duration string) (uint32, error) {
	var minutes, seconds int
	n, err := fmt.Sscanf(duration, "%d:%d", &minutes, &seconds)
	if err != nil || n != 2 {
		return 0, fmt.Errorf("invalid duration format: %s (expected MM:SS)", duration)
	}
	return uint32(minutes*60 + seconds), nil
}

// FormatDuration formats seconds to MM:SS string
func FormatDuration(seconds uint32) string {
	minutes := seconds / 60
	remainingSeconds := seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, remainingSeconds)
}

// AssertDuration validates the replay duration matches expected value
func AssertDuration(t *testing.T, actual uint32, expected string) {
	t.Helper()

	expectedSeconds, err := ParseDuration(expected)
	if err != nil {
		t.Fatalf("Invalid expected duration format: %v", err)
	}

	if actual != expectedSeconds {
		t.Errorf("Duration mismatch: expected %s (%d seconds), got %s (%d seconds)",
			expected, expectedSeconds, FormatDuration(actual), actual)
	}
}

// AssertPlayerExists validates that a player with the given name and ID exists
func AssertPlayerExists(t *testing.T, data *vault.ReplayData, expectedPlayer ExpectedPlayer) *vault.Player {
	t.Helper()

	for _, player := range data.Players {
		if player.PlayerName == expectedPlayer.Name {
			if player.PlayerID != expectedPlayer.ID {
				t.Errorf("Player %s has wrong ID: expected %d, got %d",
					expectedPlayer.Name, expectedPlayer.ID, player.PlayerID)
			}
			if player.TeamID != expectedPlayer.TeamID {
				t.Errorf("Player %s has wrong team: expected %d, got %d",
					expectedPlayer.Name, expectedPlayer.TeamID, player.TeamID)
			}
			return &player
		}
	}

	t.Errorf("Player %s (ID: %d) not found in replay data", expectedPlayer.Name, expectedPlayer.ID)
	return nil
}

// AssertCommandCount validates the number of commands for a player
func AssertCommandCount(t *testing.T, player *vault.Player, expected ExpectedPlayer) {
	t.Helper()

	if player == nil {
		return
	}

	commandCount := len(player.BuildCommands)

	if expected.MinCommands > 0 && commandCount < expected.MinCommands {
		t.Errorf("Player %s has too few commands: expected at least %d, got %d",
			player.PlayerName, expected.MinCommands, commandCount)
	}

	if expected.MaxCommands > 0 && commandCount > expected.MaxCommands {
		t.Errorf("Player %s has too many commands: expected at most %d, got %d",
			player.PlayerName, expected.MaxCommands, commandCount)
	}
}


// AssertFirstCommands validates the first few command types match expected values
func AssertFirstCommands(t *testing.T, player *vault.Player, expectedCommands []string) {
	t.Helper()

	if player == nil {
		return
	}

	if len(expectedCommands) == 0 {
		return
	}

	actualCount := len(player.BuildCommands)
	expectedCount := len(expectedCommands)

	if actualCount < expectedCount {
		t.Errorf("Player %s has only %d commands, cannot check first %d commands",
			player.PlayerName, actualCount, expectedCount)
		return
	}

	for i, expectedType := range expectedCommands {
		actual := player.BuildCommands[i].CommandType
		if actual != expectedType {
			t.Errorf("Player %s command %d type mismatch: expected %s, got %s",
				player.PlayerName, i+1, expectedType, actual)
		}
	}
}

// ValidateFullReplay performs comprehensive validation of replay data against expected values
func ValidateFullReplay(t *testing.T, data *vault.ReplayData, expected ExpectedReplayData) {
	t.Helper()

	// Validate basic fields
	if !data.Success {
		t.Errorf("Replay parsing failed: %v", data.ErrorMessage)
		return
	}

	// Validate duration
	AssertDuration(t, data.DurationSeconds, expected.Duration)

	// Validate player count
	if len(data.Players) != expected.PlayerCount {
		t.Errorf("Player count mismatch: expected %d, got %d",
			expected.PlayerCount, len(data.Players))
	}

	// Validate team count
	if len(data.Teams) != expected.TeamCount {
		t.Errorf("Team count mismatch: expected %d, got %d",
			expected.TeamCount, len(data.Teams))
	}

	// Validate winning team
	if expected.WinningTeam != nil {
		if data.WinningTeam == nil {
			t.Error("Expected winning team to be set, but it's nil")
		} else if *data.WinningTeam != *expected.WinningTeam {
			t.Errorf("Winning team mismatch: expected %d, got %d",
				*expected.WinningTeam, *data.WinningTeam)
		}
	}

	// Validate map name if specified
	if expected.MapName != "" && data.MapName != expected.MapName {
		t.Errorf("Map name mismatch: expected %s, got %s",
			expected.MapName, data.MapName)
	}

	// Validate individual players
	for _, expectedPlayer := range expected.Players {
		player := AssertPlayerExists(t, data, expectedPlayer)
		AssertCommandCount(t, player, expectedPlayer)
		AssertFirstCommands(t, player, expectedPlayer.FirstCommands)
	}
}

// GetAllTestReplays returns a list of all replay files in the testdata directory
func GetAllTestReplays() ([]string, error) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Try different possible paths for testdata directory
	testDataDirs := []string{
		filepath.Join(wd, "testdata"),                        // testdata/ (when run from tests/)
		filepath.Join(wd, "tests", "testdata"),               // tests/testdata/ (when run from parent)
		filepath.Join(filepath.Dir(wd), "tests", "testdata"), // when run from different dir
	}

	var entries []os.DirEntry
	var lastErr error

	for _, dir := range testDataDirs {
		entries, lastErr = os.ReadDir(dir)
		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to read testdata directory (searched in %v): %w", testDataDirs, lastErr)
	}

	var replays []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".rec" {
			replays = append(replays, entry.Name())
		}
	}

	return replays, nil
}

// BenchmarkReplayParsing provides performance benchmarking for replay parsing
func BenchmarkReplayParsing(b *testing.B, filename string) {
	path, err := GetTestDataPath(filename)
	if err != nil {
		b.Fatalf("Failed to get test data path: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vault.ParseReplayFull(path)
		if err != nil {
			b.Fatalf("Failed to parse replay: %v", err)
		}
	}
}

// LogReplayStats logs detailed statistics about a replay for debugging
func LogReplayStats(t *testing.T, data *vault.ReplayData) {
	t.Helper()

	t.Logf("=== Replay Statistics ===")
	t.Logf("Duration: %s (%d seconds)", FormatDuration(data.DurationSeconds), data.DurationSeconds)
	t.Logf("Map: %s", data.MapName)
	t.Logf("Teams: %d", len(data.Teams))
	t.Logf("Players: %d", len(data.Players))

	if data.WinningTeam != nil {
		t.Logf("Winning Team: %d", *data.WinningTeam)
	}

	for _, player := range data.Players {
		t.Logf("Player %d: %s (Team %d, %d commands)",
			player.PlayerID, player.PlayerName, player.TeamID, len(player.BuildCommands))

		if len(player.BuildCommands) > 0 {
			first := player.BuildCommands[0]
			var unitName string
			if first.UnitName != nil {
				unitName = *first.UnitName
			} else {
				unitName = first.CommandType
			}
			t.Logf("  First command: [%s] %s: %s",
				FormatDuration(first.Timestamp/1000), first.CommandType, unitName)
		}
	}
}
