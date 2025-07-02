package tests

import (
	"os"
	"testing"

	"github.com/scharissis/coh3-replay-analyser/vault"
)

// BenchmarkParseReplayFull benchmarks the full replay parsing function
func BenchmarkParseReplayFull(b *testing.B) {
	replayPath, err := GetTestDataPath("temp_29_06_2025__22_49.rec")
	if err != nil {
		b.Skipf("Test replay file not found: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vault.ParseReplayFull(replayPath)
		if err != nil {
			b.Fatalf("Failed to parse replay: %v", err)
		}
	}
}

// BenchmarkMemoryUsage provides insights into memory allocation patterns
func BenchmarkMemoryUsage(b *testing.B) {
	replayPath, err := GetTestDataPath("temp_29_06_2025__22_49.rec")
	if err != nil {
		b.Skipf("Test replay file not found: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, err := vault.ParseReplayFull(replayPath)
		if err != nil {
			b.Fatalf("Failed to parse replay: %v", err)
		}

		// Access data to prevent optimization
		_ = data.DurationSeconds
		_ = len(data.Players)
		_ = len(data.Teams)
	}
}

// BenchmarkConcurrentParsing tests performance under concurrent load
func BenchmarkConcurrentParsing(b *testing.B) {
	replayPath, err := GetTestDataPath("temp_29_06_2025__22_49.rec")
	if err != nil {
		b.Skipf("Test replay file not found: %v", err)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := vault.ParseReplayFull(replayPath)
			if err != nil {
				b.Fatalf("Failed to parse replay: %v", err)
			}
		}
	})
}

// BenchmarkFileOperations benchmarks just the file reading part
func BenchmarkFileOperations(b *testing.B) {
	replayPath, err := GetTestDataPath("temp_29_06_2025__22_49.rec")
	if err != nil {
		b.Skipf("Test replay file not found: %v", err)
	}

	b.Run("FileRead", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			data, err := os.ReadFile(replayPath)
			if err != nil {
				b.Fatalf("Failed to read file: %v", err)
			}
			_ = len(data) // Prevent optimization
		}
	})

	b.Run("FileReadAndParse", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := vault.ParseReplayFull(replayPath)
			if err != nil {
				b.Fatalf("Failed to parse replay: %v", err)
			}
		}
	})
}
