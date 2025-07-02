package vault

/*
#cgo LDFLAGS: -L./lib -lvault_wrapper -ldl -lm
#include <stdlib.h>
#include <stdbool.h>

typedef struct {
    bool include_build_squad;
    bool include_construct_entity;
    bool include_build_global_upgrade;
    bool include_use_ability;
    bool include_use_battlegroup_ability;
    bool include_select_battlegroup;
    bool include_select_battlegroup_ability;
    bool include_cancel_construction;
    bool include_cancel_production;
    bool include_ai_takeover;
    bool include_unknown;
} CCommandFilter;

char* parse_replay_full(const char* file_path);
char* parse_replay_with_filter(const char* file_path, const CCommandFilter* filter);
void free_string(char* s);
*/
import "C"
import (
	"encoding/json"
	"errors"
	"strconv"
	"unsafe"

	"github.com/scharissis/coh3-replay-analyser/pkg/lookup"
)

// CommandFilter represents configuration for filtering commands
type CommandFilter struct {
	IncludeBuildSquad              bool `json:"include_build_squad"`
	IncludeConstructEntity         bool `json:"include_construct_entity"`
	IncludeBuildGlobalUpgrade      bool `json:"include_build_global_upgrade"`
	IncludeUseAbility              bool `json:"include_use_ability"`
	IncludeUseBattlegroupAbility   bool `json:"include_use_battlegroup_ability"`
	IncludeSelectBattlegroup       bool `json:"include_select_battlegroup"`
	IncludeSelectBattlegroupAbility bool `json:"include_select_battlegroup_ability"`
	IncludeCancelConstruction      bool `json:"include_cancel_construction"`
	IncludeCancelProduction        bool `json:"include_cancel_production"`
	IncludeAITakeover              bool `json:"include_ai_takeover"`
	IncludeUnknown                 bool `json:"include_unknown"`
}

// NewBuildOnlyFilter creates a filter that only includes build-related commands
func NewBuildOnlyFilter() CommandFilter {
	return CommandFilter{
		IncludeBuildSquad:              true,
		IncludeConstructEntity:         true,
		IncludeBuildGlobalUpgrade:      true,
		IncludeUseAbility:              false,
		IncludeUseBattlegroupAbility:   false,
		IncludeSelectBattlegroup:       true,
		IncludeSelectBattlegroupAbility: true,
		IncludeCancelConstruction:      false,
		IncludeCancelProduction:        false,
		IncludeAITakeover:              false,
		IncludeUnknown:                 false,
	}
}

// NewAllCommandsFilter creates a filter that includes all command types
func NewAllCommandsFilter() CommandFilter {
	return CommandFilter{
		IncludeBuildSquad:              true,
		IncludeConstructEntity:         true,
		IncludeBuildGlobalUpgrade:      true,
		IncludeUseAbility:              true,
		IncludeUseBattlegroupAbility:   true,
		IncludeSelectBattlegroup:       true,
		IncludeSelectBattlegroupAbility: true,
		IncludeCancelConstruction:      true,
		IncludeCancelProduction:        true,
		IncludeAITakeover:              true,
		IncludeUnknown:                 true,
	}
}

// NewCombatOnlyFilter creates a filter that only includes combat-related commands
func NewCombatOnlyFilter() CommandFilter {
	return CommandFilter{
		IncludeBuildSquad:              false,
		IncludeConstructEntity:         false,
		IncludeBuildGlobalUpgrade:      false,
		IncludeUseAbility:              true,
		IncludeUseBattlegroupAbility:   true,
		IncludeSelectBattlegroup:       false,
		IncludeSelectBattlegroupAbility: false,
		IncludeCancelConstruction:      false,
		IncludeCancelProduction:        false,
		IncludeAITakeover:              false,
		IncludeUnknown:                 false,
	}
}

// Command represents a command with detailed information
type Command struct {
	Timestamp    uint32  `json:"timestamp"`
	CommandType  string  `json:"command_type"`
	Details      string  `json:"details"`
	PBGID        *string `json:"pbgid,omitempty"`
	UnitName     *string `json:"unit_name,omitempty"`
	BuildingName *string `json:"building_name,omitempty"`
}

// Team represents a team in the replay
type Team struct {
	TeamID  uint32       `json:"team_id"`
	Players []PlayerInfo `json:"players"`
}

// PlayerInfo represents basic player information
type PlayerInfo struct {
	PlayerID   uint32  `json:"player_id"`
	PlayerName string  `json:"player_name"`
	Faction    *string `json:"faction,omitempty"`
	IsHuman    bool    `json:"is_human"`
	SteamID    *string `json:"steam_id,omitempty"`
	ProfileID  *string `json:"profile_id,omitempty"`
}

// ReplayData represents comprehensive replay information
type ReplayData struct {
	Success      bool    `json:"success"`
	ErrorMessage *string `json:"error_message,omitempty"`
	// Match Information
	MapName         string  `json:"map_name"`
	MapFilename     string  `json:"map_filename"`
	DurationSeconds uint32  `json:"duration_seconds"`
	DurationTicks   uint32  `json:"duration_ticks"`
	GameVersion     *uint16 `json:"game_version,omitempty"`
	Timestamp       *string `json:"timestamp,omitempty"`
	GameType        *string `json:"game_type,omitempty"`
	MatchHistoryID  *string `json:"matchhistory_id,omitempty"`
	// Teams and Players
	Teams       []Team   `json:"teams"`
	WinningTeam *uint32  `json:"winning_team,omitempty"`
	Players     []Player `json:"players"`
	// Messages and Events
	Messages []GameMessage `json:"messages"`
}

// Player represents a player with comprehensive command and metadata information
type Player struct {
	PlayerID      uint32        `json:"player_id"`
	PlayerName    string        `json:"player_name"`
	TeamID        uint32        `json:"team_id"`
	Faction       *string       `json:"faction,omitempty"`
	IsHuman       bool          `json:"is_human"`
	SteamID       *string       `json:"steam_id,omitempty"`
	ProfileID     *string       `json:"profile_id,omitempty"`
	BattlegroupID *string       `json:"battlegroup_id,omitempty"`
	Commands      []Command     `json:"commands"`
	BuildCommands []Command     `json:"build_commands"`
	ChatMessages  []GameMessage `json:"chat_messages"`
}

// GameMessage represents a chat message or game event
type GameMessage struct {
	Timestamp   uint32  `json:"timestamp"`
	PlayerID    *uint32 `json:"player_id,omitempty"`
	Content     string  `json:"content"`
	MessageType string  `json:"message_type"`
}

// ParseReplayFull parses a Company of Heroes 3 replay file and extracts all available information
// This is the comprehensive parsing function that extracts everything at once
func ParseReplayFull(filePath string) (*ReplayData, error) {
	cFilePath := C.CString(filePath)
	defer C.free(unsafe.Pointer(cFilePath))

	cResult := C.parse_replay_full(cFilePath)
	if cResult == nil {
		return nil, errors.New("failed to parse replay file")
	}
	defer C.free_string(cResult)

	resultStr := C.GoString(cResult)

	var result ReplayData
	if err := json.Unmarshal([]byte(resultStr), &result); err != nil {
		return nil, err
	}

	if !result.Success && result.ErrorMessage != nil {
		return &result, errors.New(*result.ErrorMessage)
	}
	if !result.Success {
		return nil, errors.New("failed to parse replay; got success=false")
	}

	return &result, nil
}

// ParseReplayWithLookup parses a replay file and enhances commands with friendly names
func ParseReplayWithLookup(filePath string, dataDir string) (*ReplayData, error) {
	// Use default build-only filter for backwards compatibility
	filter := NewBuildOnlyFilter()
	return ParseReplayWithFilter(filePath, dataDir, filter)
}

// ParseReplayWithFilter parses a replay file with a custom command filter and enhances commands with friendly names
func ParseReplayWithFilter(filePath string, dataDir string, filter CommandFilter) (*ReplayData, error) {
	// Convert Go filter to C filter
	cFilter := C.CCommandFilter{
		include_build_squad:               C.bool(filter.IncludeBuildSquad),
		include_construct_entity:          C.bool(filter.IncludeConstructEntity),
		include_build_global_upgrade:      C.bool(filter.IncludeBuildGlobalUpgrade),
		include_use_ability:               C.bool(filter.IncludeUseAbility),
		include_use_battlegroup_ability:   C.bool(filter.IncludeUseBattlegroupAbility),
		include_select_battlegroup:        C.bool(filter.IncludeSelectBattlegroup),
		include_select_battlegroup_ability: C.bool(filter.IncludeSelectBattlegroupAbility),
		include_cancel_construction:       C.bool(filter.IncludeCancelConstruction),
		include_cancel_production:         C.bool(filter.IncludeCancelProduction),
		include_ai_takeover:               C.bool(filter.IncludeAITakeover),
		include_unknown:                   C.bool(filter.IncludeUnknown),
	}

	// Call the Rust function with filter
	cFilePath := C.CString(filePath)
	defer C.free(unsafe.Pointer(cFilePath))

	cResult := C.parse_replay_with_filter(cFilePath, &cFilter)
	if cResult == nil {
		return nil, errors.New("failed to parse replay: null result")
	}
	defer C.free_string(cResult)

	result := C.GoString(cResult)

	var replayData ReplayData
	if err := json.Unmarshal([]byte(result), &replayData); err != nil {
		return nil, err
	}

	if !replayData.Success {
		if replayData.ErrorMessage != nil {
			return nil, errors.New(*replayData.ErrorMessage)
		}
		return nil, errors.New("failed to parse replay: unknown error")
	}

	// Initialize the lookup resolver for enhancing command names
	resolver, err := lookup.NewDataResolver(dataDir)
	if err != nil {
		// If lookup fails, just return the basic data without enhancement
		return &replayData, nil
	}

	// Enhance all player commands with friendly names
	for i := range replayData.Players {
		enhanceCommands(replayData.Players[i].Commands, resolver)
		enhanceCommands(replayData.Players[i].BuildCommands, resolver)
	}

	return &replayData, nil
}

// enhanceCommands adds friendly names to commands using the lookup resolver
func enhanceCommands(commands []Command, resolver *lookup.DataResolver) {
	for i := range commands {
		cmd := &commands[i]
		
		// Handle commands without PBGID
		if cmd.PBGID == nil {
			// For construction commands without PBGID, provide a generic name
			switch cmd.CommandType {
			case "construct_entity":
				buildingName := "Building Construction"
				cmd.BuildingName = &buildingName
			case "build_global_upgrade":
				upgradeName := "Technology Upgrade"
				cmd.UnitName = &upgradeName
			case "select_battlegroup":
				name := "Battlegroup Selection"
				cmd.UnitName = &name
			case "select_battlegroup_ability":
				name := "Battlegroup Ability Selection"
				cmd.UnitName = &name
			}
			continue
		}

		// Parse PBGID string to uint32
		pbgidStr := *cmd.PBGID
		pbgid, err := strconv.ParseUint(pbgidStr, 10, 32)
		if err != nil {
			continue
		}

		// Resolve the PBGID to friendly info
		if unitInfo, err := resolver.ResolvePBGID(uint32(pbgid)); err == nil {
			// Set the appropriate field based on command type
			switch cmd.CommandType {
			case "build_squad":
				cmd.UnitName = &unitInfo.Name
			case "construct_entity":
				cmd.BuildingName = &unitInfo.Name
			case "use_ability":
				cmd.UnitName = &unitInfo.Name
			default:
				// For other commands, use unit name
				cmd.UnitName = &unitInfo.Name
			}
		}
	}
}
