# COH3 Build Order - Integration Tests

This directory contains comprehensive end-to-end integration tests for the COH3 Build Order CLI tool.

## Directory Structure

```
tests/
├── go.mod                  # Go module for tests
├── README.md              # This file
├── helpers.go             # Test utilities and assertions
├── integration_test.go    # Core integration tests
├── cli_test.go           # CLI command tests
├── benchmark_test.go     # Performance benchmarks
├── testdata/             # Replay files for testing
│   └── temp_29_06_2025__22_49.rec
└── fixtures/             # Expected test results in JSON
    └── temp_29_06_2025__22_49.json
```

## Adding New Test Replays

1. **Add replay file**: Place your `.rec` file in the `testdata/` directory
2. **Create fixture**: Create a corresponding JSON fixture in `fixtures/` with expected results
3. **Update tests**: Add your replay to the test cases in `integration_test.go`

### Example Fixture Format

```json
{
  "file_name": "your_replay.rec",
  "expected_results": {
    "duration": "25:30",
    "duration_seconds": 1530,
    "map_name": "expected_map_name",
    "player_count": 4,
    "team_count": 2,
    "winning_team": 1,
    "players": [
      {
        "id": 0,
        "name": "PlayerName",
        "team_id": 1,
        "min_commands": 5,
        "max_commands": 20,
        "has_real_commands": true,
        "first_commands": ["train", "build", "ability"]
      }
    ]
  }
}
```

## Running Tests

### All Tests
```bash
cd tests
go test -v
```

### Specific Test Categories
```bash
# Integration tests only
go test -v -run "TestReplayFiles"

# CLI tests only  
go test -v -run "TestCLI"

# Error handling tests
go test -v -run "TestError"

# Consistency tests
go test -v -run "TestConsistency"
```

### Performance Benchmarks
```bash
# All benchmarks
go test -bench=.

# Specific benchmarks
go test -bench=BenchmarkParseReplayFull
go test -bench=BenchmarkConcurrent

# With memory allocation stats
go test -bench=. -benchmem

# Multiple runs for accuracy
go test -bench=. -count=5
```

### Benchmark Results Analysis
```bash
# Compare performance over time
go test -bench=. -count=10 | tee benchmark_results.txt

# Profile memory usage
go test -bench=BenchmarkMemoryUsage -memprofile=mem.prof
go tool pprof mem.prof

# Profile CPU usage  
go test -bench=BenchmarkParseReplayFull -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

## Test Categories

### 1. Integration Tests (`integration_test.go`)
- **TestReplayFiles**: Table-driven tests using JSON fixtures
- **TestParseReplayFull_Comprehensive**: Detailed validation of full parsing
- **TestParseReplayPlayer_ByName**: Player filtering by name
- **TestParseReplayPlayer_ByID**: Player filtering by ID  
- **TestGetReplayInfo**: High-level replay information
- **TestAllReplaysInTestData**: Validates all replay files can be parsed
- **TestErrorHandling**: Various error conditions
- **TestConsistencyBetweenFunctions**: Ensures all parsing functions return consistent data

### 2. CLI Tests (`cli_test.go`)
- **TestCLI_InfoCommand**: Tests `info` command output
- **TestCLI_BuildOrderCommand**: Tests `build-order` command with various options
- **TestCLI_FullCommand**: Tests `full` command comprehensive output
- **TestCLI_ErrorHandling**: CLI error conditions and messages
- **TestCLI_Help**: Help functionality
- **TestCLI_OutputConsistency**: Ensures CLI output matches library results

### 3. Performance Tests (`benchmark_test.go`)
- **BenchmarkParseReplayFull**: Full parsing performance
- **BenchmarkParseReplayPlayer**: Player-specific parsing
- **BenchmarkGetReplayInfo**: Info extraction performance
- **BenchmarkCompareParsingMethods**: Compares different approaches
- **BenchmarkMemoryUsage**: Memory allocation patterns
- **BenchmarkConcurrentParsing**: Performance under concurrent load
- **BenchmarkFileOperations**: File I/O vs parsing overhead

## Test Utilities (`helpers.go`)

### Core Functions
- `GetTestDataPath(filename)`: Get absolute path to test data files
- `LoadTestFixture(fixtureName)`: Load expected results from JSON
- `ValidateFullReplay(data, expected)`: Comprehensive replay validation

### Assertion Helpers
- `AssertDuration(actual, expected)`: Validate replay duration
- `AssertPlayerExists(data, expected)`: Validate player presence and attributes
- `AssertCommandCount(player, expected)`: Validate command counts
- `AssertRealCommands(player, shouldHaveReal)`: Ensure commands aren't placeholder data
- `AssertFirstCommands(player, expected)`: Validate first few command types

### Utility Functions
- `ParseDuration(duration)`: Convert "MM:SS" to seconds
- `FormatDuration(seconds)`: Convert seconds to "MM:SS"
- `GetAllTestReplays()`: List all replay files in testdata
- `LogReplayStats(data)`: Debug logging of replay statistics

## Continuous Integration

These tests are designed to:
1. **Prevent regressions**: Ensure functionality doesn't break with code changes
2. **Validate accuracy**: Confirm parsed data matches expected real values
3. **Performance monitoring**: Track performance changes over time
4. **Cross-platform testing**: Verify behavior across different environments

## Best Practices

1. **Always add fixtures**: Create JSON fixtures for expected results when adding replays
2. **Test edge cases**: Include replays with unusual characteristics (short games, disconnects, etc.)
3. **Validate real data**: Ensure tests check for actual game data, not placeholder values
4. **Performance baselines**: Regularly run benchmarks to establish performance baselines
5. **Error coverage**: Test various error conditions and edge cases

## Troubleshooting

### Test Failures
- Check that the CLI binary is built: `make build` in parent directory
- Verify replay files exist in `testdata/`
- Ensure fixtures have correct expected values
- Check Go module dependencies are up to date

### Performance Issues
- Profile using Go's built-in profiling tools
- Compare benchmark results over time
- Check for memory leaks in long-running tests
- Monitor file I/O vs parsing performance ratios