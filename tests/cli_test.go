package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scharissis/coh3-replay-analyser/vault"
)

// getCLIPath returns the path to the compiled CLI binary
func getCLIPath() (string, error) {
	// Look for the binary in the parent directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	parentDir := filepath.Dir(wd)
	binaryPath := filepath.Join(parentDir, "coh3-build-order")

	// Check if binary exists
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, nil
	}

	// If not found, try to build it
	cmd := exec.Command("make", "build")
	cmd.Dir = parentDir
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return binaryPath, nil
}

// runCLICommand executes a CLI command and returns stdout, stderr, and error
func runCLICommand(args ...string) (string, string, error) {
	cliPath, err := getCLIPath()
	if err != nil {
		return "", "", err
	}

	cmd := exec.Command(cliPath, args...)

	// Set working directory to parent for relative file paths
	wd, _ := os.Getwd()
	cmd.Dir = filepath.Dir(wd)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	return stdout.String(), stderr.String(), err
}

// TestCLI_InfoCommand tests the info command output
func TestCLI_InfoCommand(t *testing.T) {
	// Use relative path from parent directory
	replayFile := "tests/testdata/temp_29_06_2025__22_49.rec"

	stdout, stderr, err := runCLICommand("info", replayFile)
	if err != nil {
		t.Fatalf("CLI info command failed: %v\nStderr: %s", err, stderr)
	}

	// Verify expected content in output
	expectedStrings := []string{
		"=== Replay Information ===",
		"Duration: 38:27",
		"Winning Team: Unknown",
		"=== Teams ===",
		"Tomsch",
		"Ted 'Seaman' Silk",
		"Angrybirds",
		"Thomas Smooth",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s'\nActual output:\n%s", expected, stdout)
		}
	}

	// Verify team structure
	if !strings.Contains(stdout, "Team 1:") || !strings.Contains(stdout, "Team 2:") {
		t.Errorf("Expected output to show both teams\nActual output:\n%s", stdout)
	}

	// Verify player IDs are shown
	if !strings.Contains(stdout, "ID 0:") || !strings.Contains(stdout, "ID 1:") {
		t.Errorf("Expected output to show player IDs\nActual output:\n%s", stdout)
	}
}

// TestCLI_BuildOrderCommand tests the build-order command output
func TestCLI_BuildOrderCommand(t *testing.T) {
	replayFile := "tests/testdata/temp_29_06_2025__22_49.rec"

	t.Run("AllPlayers", func(t *testing.T) {
		stdout, stderr, err := runCLICommand("build-order", replayFile)
		if err != nil {
			t.Fatalf("CLI build-order command failed: %v\nStderr: %s", err, stderr)
		}

		// Should show all 4 players
		expectedPlayers := []string{
			"=== Player 0: Tomsch ===",
			"=== Player 1: Ted 'Seaman' Silk ===",
			"=== Player 2: Angrybirds ===",
			"=== Player 3: Thomas Smooth ===",
		}

		for _, expected := range expectedPlayers {
			if !strings.Contains(stdout, expected) {
				t.Errorf("Expected output to contain '%s'\nActual output:\n%s", expected, stdout)
			}
		}

		// Should show build orders with timestamps
		if !strings.Contains(stdout, "Build Order:") {
			t.Errorf("Expected output to contain 'Build Order:'\nActual output:\n%s", stdout)
		}

		// Should show numbered commands with timestamps
		if !strings.Contains(stdout, "1. [") || !strings.Contains(stdout, "]") {
			t.Errorf("Expected output to show numbered commands with timestamps\nActual output:\n%s", stdout)
		}
	})

	t.Run("SpecificPlayerByName", func(t *testing.T) {
		stdout, stderr, err := runCLICommand("build-order", "-p", "Tomsch", replayFile)
		if err != nil {
			t.Fatalf("CLI build-order command failed: %v\nStderr: %s", err, stderr)
		}

		// Should show only Tomsch
		if !strings.Contains(stdout, "=== Player 0: Tomsch ===") {
			t.Errorf("Expected output to contain Tomsch's info\nActual output:\n%s", stdout)
		}

		// Should NOT show other players
		otherPlayers := []string{"Ted 'Seaman' Silk", "Angrybirds", "Thomas Smooth"}
		for _, player := range otherPlayers {
			if strings.Contains(stdout, player) {
				t.Errorf("Output should not contain '%s' when filtering for Tomsch\nActual output:\n%s", player, stdout)
			}
		}

		// Should show build commands
		if !strings.Contains(stdout, "Build Order:") {
			t.Errorf("Expected output to contain 'Build Order:'\nActual output:\n%s", stdout)
		}
	})

	t.Run("SpecificPlayerByID", func(t *testing.T) {
		stdout, stderr, err := runCLICommand("build-order", "-p", "0", replayFile)
		if err != nil {
			t.Fatalf("CLI build-order command failed: %v\nStderr: %s", err, stderr)
		}

		// Should show only player 0 (Tomsch)
		if !strings.Contains(stdout, "=== Player 0: Tomsch ===") {
			t.Errorf("Expected output to contain player 0 info\nActual output:\n%s", stdout)
		}

		// Should NOT show other players
		if strings.Contains(stdout, "Player 1:") || strings.Contains(stdout, "Player 2:") || strings.Contains(stdout, "Player 3:") {
			t.Errorf("Output should only show player 0\nActual output:\n%s", stdout)
		}
	})

	t.Run("VerboseMode", func(t *testing.T) {
		stdout, stderr, err := runCLICommand("build-order", "-v", "-p", "Tomsch", replayFile)
		if err != nil {
			t.Fatalf("CLI build-order command with verbose failed: %v\nStderr: %s", err, stderr)
		}

		// Verbose mode should show parsing information
		expectedVerboseStrings := []string{
			"Parsing replay file:",
			"Extracting build order for player: Tomsch",
		}

		for _, expected := range expectedVerboseStrings {
			if !strings.Contains(stdout, expected) {
				t.Errorf("Expected verbose output to contain '%s'\nActual output:\n%s", expected, stdout)
			}
		}
	})
}

// TestCLI_FullCommand tests the full command output
func TestCLI_FullCommand(t *testing.T) {
	replayFile := "tests/testdata/temp_29_06_2025__22_49.rec"

	stdout, stderr, err := runCLICommand("full", replayFile)
	if err != nil {
		t.Fatalf("CLI full command failed: %v\nStderr: %s", err, stderr)
	}

	// Verify comprehensive output sections
	expectedSections := []string{
		"=== Comprehensive Replay Data ===",
		"=== Teams ===",
		"=== All Players with Build Orders ===",
		"=== Messages ===",
	}

	for _, expected := range expectedSections {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain section '%s'\nActual output:\n%s", expected, stdout)
		}
	}

	// Verify it shows more detail than other commands
	if !strings.Contains(stdout, "Duration: 38:27") {
		t.Errorf("Expected output to contain duration\nActual output:\n%s", stdout)
	}

	// Should show all players with team info
	if !strings.Contains(stdout, "(Team 1)") || !strings.Contains(stdout, "(Team 2)") {
		t.Errorf("Expected output to show team information for players\nActual output:\n%s", stdout)
	}

	// Should show message section (even if placeholder)
	if !strings.Contains(stdout, "Messages") {
		t.Errorf("Expected output to contain messages section\nActual output:\n%s", stdout)
	}
}

// TestCLI_ErrorHandling tests CLI error conditions
func TestCLI_ErrorHandling(t *testing.T) {
	t.Run("NonexistentFile", func(t *testing.T) {
		stdout, stderr, err := runCLICommand("info", "nonexistent.rec")
		if err == nil {
			t.Error("Expected error for nonexistent file, but command succeeded")
		}

		// Should show helpful error message
		combinedOutput := stdout + stderr
		if !strings.Contains(combinedOutput, "does not exist") && !strings.Contains(combinedOutput, "no such file") {
			t.Errorf("Expected error message about file not existing\nOutput: %s", combinedOutput)
		}
	})

	t.Run("InvalidCommand", func(t *testing.T) {
		stdout, stderr, err := runCLICommand("invalid-command")
		if err == nil {
			t.Error("Expected error for invalid command, but command succeeded")
		}

		// Should show help or usage information
		combinedOutput := stdout + stderr
		if !strings.Contains(combinedOutput, "unknown command") && !strings.Contains(combinedOutput, "Usage") {
			t.Errorf("Expected help/usage information for invalid command\nOutput: %s", combinedOutput)
		}
	})

	t.Run("MissingArguments", func(t *testing.T) {
		stdout, stderr, err := runCLICommand("info")
		if err == nil {
			t.Error("Expected error for missing arguments, but command succeeded")
		}

		// Should show usage information
		combinedOutput := stdout + stderr
		if !strings.Contains(combinedOutput, "accepts") && !strings.Contains(combinedOutput, "requires") {
			t.Errorf("Expected error about missing arguments\nOutput: %s", combinedOutput)
		}
	})
}

// TestCLI_Help tests help functionality
func TestCLI_Help(t *testing.T) {
	t.Run("RootHelp", func(t *testing.T) {
		stdout, stderr, err := runCLICommand("--help")
		if err != nil {
			t.Fatalf("Help command failed: %v\nStderr: %s", err, stderr)
		}

		// Should show available commands
		expectedCommands := []string{
			"build-order",
			"info",
			"full",
		}

		for _, cmd := range expectedCommands {
			if !strings.Contains(stdout, cmd) {
				t.Errorf("Expected help to mention command '%s'\nActual output:\n%s", cmd, stdout)
			}
		}

		// Should show usage
		if !strings.Contains(stdout, "Usage:") {
			t.Errorf("Expected help to show usage\nActual output:\n%s", stdout)
		}
	})

	t.Run("CommandSpecificHelp", func(t *testing.T) {
		stdout, stderr, err := runCLICommand("build-order", "--help")
		if err != nil {
			t.Fatalf("Build-order help failed: %v\nStderr: %s", err, stderr)
		}

		// Should show command-specific options
		expectedContent := []string{
			"--player",
			"-p",
			"--verbose",
			"-v",
			"Examples:",
		}

		for _, content := range expectedContent {
			if !strings.Contains(stdout, content) {
				t.Errorf("Expected build-order help to contain '%s'\nActual output:\n%s", content, stdout)
			}
		}
	})
}

// TestCLI_OutputConsistency ensures CLI output matches library function results
func TestCLI_OutputConsistency(t *testing.T) {
	replayFile := "tests/testdata/temp_29_06_2025__22_49.rec"

	// Get library results
	replayPath, err := GetTestDataPath("temp_29_06_2025__22_49.rec")
	if err != nil {
		t.Skipf("Test replay file not found: %v", err)
	}

	data, err := vault.ParseReplayFull(replayPath)
	if err != nil {
		t.Fatalf("Failed to get replay data: %v", err)
	}

	// Get CLI output
	stdout, stderr, err := runCLICommand("info", replayFile)
	if err != nil {
		t.Fatalf("CLI info command failed: %v\nStderr: %s", err, stderr)
	}

	// Check duration consistency
	expectedDuration := FormatDuration(data.DurationSeconds)
	if !strings.Contains(stdout, "Duration: "+expectedDuration) {
		t.Errorf("CLI duration doesn't match library result. Expected: %s\nCLI output:\n%s", expectedDuration, stdout)
	}

	// Check winning team consistency
	if data.WinningTeam != nil {
		expectedWinning := fmt.Sprintf("Winning Team: %d", *data.WinningTeam)
		if !strings.Contains(stdout, expectedWinning) {
			t.Errorf("CLI winning team doesn't match library result. Expected: %s\nCLI output:\n%s", expectedWinning, stdout)
		}
	}

	// Check that all players are mentioned
	for _, team := range data.Teams {
		for _, player := range team.Players {
			if !strings.Contains(stdout, player.PlayerName) {
				t.Errorf("CLI output missing player: %s\nCLI output:\n%s", player.PlayerName, stdout)
			}
		}
	}
}
