use libc::c_char;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::ffi::{CStr, CString};
use std::ptr;

const TICKS_PER_SECOND: u32 = 8;
const MILLISECONDS_PER_TICK: u32 = 125;

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct Command {
    pub timestamp: u32,
    pub command_type: String,
    pub details: String,
    pub pbgid: Option<String>,         // Raw PBGID for reference
    pub unit_name: Option<String>,     // Resolved unit name if available
    pub building_name: Option<String>, // Building context if applicable
}


#[derive(Serialize, Deserialize, Debug)]
pub struct Team {
    pub team_id: u32,
    pub players: Vec<PlayerInfo>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct PlayerInfo {
    pub player_id: u32,
    pub player_name: String,
    pub team_id: u32,
    pub faction: Option<String>,       // Player faction
    pub is_human: bool,               // Human vs AI
    pub steam_id: Option<String>,     // Steam ID if available
    pub profile_id: Option<String>,   // Relic profile ID if available
}

#[derive(Serialize, Deserialize, Debug)]
pub struct ReplayData {
    pub success: bool,
    pub error_message: Option<String>,
    // Match Information
    pub map_name: String,
    pub map_filename: String,
    pub duration_seconds: u32,
    pub duration_ticks: u32,
    pub game_version: Option<u16>,
    pub timestamp: Option<String>,
    pub game_type: Option<String>,
    pub matchhistory_id: Option<String>,
    // Teams and Players
    pub teams: Vec<Team>,
    pub winning_team: Option<u32>,
    pub players: Vec<Player>,
    // Messages and Events
    pub messages: Vec<GameMessage>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Player {
    pub player_id: u32,
    pub player_name: String,
    pub team_id: u32,
    pub faction: Option<String>,
    pub is_human: bool,
    pub steam_id: Option<String>,
    pub profile_id: Option<String>,
    pub battlegroup_id: Option<String>,
    pub commands: Vec<Command>,
    pub build_commands: Vec<Command>,    // Subset of commands that are build-related
    pub chat_messages: Vec<GameMessage>, // Player's chat messages
}

#[derive(Serialize, Deserialize, Debug)]
pub struct GameMessage {
    pub timestamp: u32,
    pub player_id: Option<u32>,
    pub content: String,
    pub message_type: String,
}


#[no_mangle]
pub extern "C" fn parse_replay_full(file_path: *const c_char) -> *mut c_char {
    if file_path.is_null() {
        let error_result = ReplayData {
            success: false,
            error_message: Some("File path is null".to_string()),
            map_name: String::new(),
            map_filename: String::new(),
            duration_seconds: 0,
            duration_ticks: 0,
            game_version: None,
            timestamp: None,
            game_type: None,
            matchhistory_id: None,
            teams: vec![],
            winning_team: None,
            players: vec![],
            messages: vec![],
        };
        return serialize_replay_data(&error_result);
    }

    let c_str = unsafe { CStr::from_ptr(file_path) };
    let file_path_str = match c_str.to_str() {
        Ok(s) => s,
        Err(_) => {
            let error_result = ReplayData {
                success: false,
                error_message: Some("Invalid file path encoding".to_string()),
                map_name: String::new(),
                map_filename: String::new(),
                duration_seconds: 0,
                duration_ticks: 0,
                game_version: None,
                timestamp: None,
                game_type: None,
                matchhistory_id: None,
                teams: vec![],
                winning_team: None,
                players: vec![],
                messages: vec![],
            };
            return serialize_replay_data(&error_result);
        }
    };

    match parse_replay_full_internal(file_path_str) {
        Ok(result) => serialize_replay_data(&result),
        Err(e) => {
            let error_result = ReplayData {
                success: false,
                error_message: Some(e),
                map_name: String::new(),
                map_filename: String::new(),
                duration_seconds: 0,
                duration_ticks: 0,
                game_version: None,
                timestamp: None,
                game_type: None,
                matchhistory_id: None,
                teams: vec![],
                winning_team: None,
                players: vec![],
                messages: vec![],
            };
            serialize_replay_data(&error_result)
        }
    }
}




// C-compatible command filter structure
#[repr(C)]
#[derive(Debug, Clone)]
pub struct CCommandFilter {
    pub include_build_squad: bool,
    pub include_construct_entity: bool,
    pub include_build_global_upgrade: bool,
    pub include_use_ability: bool,
    pub include_use_battlegroup_ability: bool,
    pub include_select_battlegroup: bool,
    pub include_select_battlegroup_ability: bool,
    pub include_cancel_construction: bool,
    pub include_cancel_production: bool,
    pub include_ai_takeover: bool,
    pub include_unknown: bool,
}

impl From<CCommandFilter> for CommandFilter {
    fn from(c_filter: CCommandFilter) -> Self {
        Self {
            include_build_squad: c_filter.include_build_squad,
            include_construct_entity: c_filter.include_construct_entity,
            include_build_global_upgrade: c_filter.include_build_global_upgrade,
            include_use_ability: c_filter.include_use_ability,
            include_use_battlegroup_ability: c_filter.include_use_battlegroup_ability,
            include_select_battlegroup: c_filter.include_select_battlegroup,
            include_select_battlegroup_ability: c_filter.include_select_battlegroup_ability,
            include_cancel_construction: c_filter.include_cancel_construction,
            include_cancel_production: c_filter.include_cancel_production,
            include_ai_takeover: c_filter.include_ai_takeover,
            include_unknown: c_filter.include_unknown,
        }
    }
}

#[no_mangle]
pub extern "C" fn parse_replay_with_filter(file_path: *const c_char, filter: *const CCommandFilter) -> *mut c_char {
    if file_path.is_null() {
        let error_result = ReplayData {
            success: false,
            error_message: Some("File path is null".to_string()),
            map_name: String::new(),
            map_filename: String::new(),
            duration_seconds: 0,
            duration_ticks: 0,
            game_version: None,
            timestamp: None,
            game_type: None,
            matchhistory_id: None,
            teams: vec![],
            winning_team: None,
            players: vec![],
            messages: vec![],
        };
        return serialize_replay_data(&error_result);
    }

    let c_str = unsafe { CStr::from_ptr(file_path) };
    let file_path_str = match c_str.to_str() {
        Ok(s) => s,
        Err(_) => {
            let error_result = ReplayData {
                success: false,
                error_message: Some("Invalid file path encoding".to_string()),
                map_name: String::new(),
                map_filename: String::new(),
                duration_seconds: 0,
                duration_ticks: 0,
                game_version: None,
                timestamp: None,
                game_type: None,
                matchhistory_id: None,
                teams: vec![],
                winning_team: None,
                players: vec![],
                messages: vec![],
            };
            return serialize_replay_data(&error_result);
        }
    };

    let command_filter = if filter.is_null() {
        CommandFilter::default()
    } else {
        let c_filter = unsafe { &*filter };
        c_filter.clone().into()
    };

    match parse_replay_with_filter_internal(file_path_str, &command_filter) {
        Ok(result) => serialize_replay_data(&result),
        Err(e) => {
            let error_result = ReplayData {
                success: false,
                error_message: Some(e),
                map_name: String::new(),
                map_filename: String::new(),
                duration_seconds: 0,
                duration_ticks: 0,
                game_version: None,
                timestamp: None,
                game_type: None,
                matchhistory_id: None,
                teams: vec![],
                winning_team: None,
                players: vec![],
                messages: vec![],
            };
            serialize_replay_data(&error_result)
        }
    }
}

#[no_mangle]
pub extern "C" fn free_string(s: *mut c_char) {
    if !s.is_null() {
        unsafe {
            let _ = CString::from_raw(s);
        }
    }
}


fn parse_replay_with_filter_internal(file_path: &str, command_filter: &CommandFilter) -> Result<ReplayData, String> {
    let data = std::fs::read(file_path)
        .map_err(|e| format!("Failed to read file: {}", e))?;
    
    let replay = vault::Replay::from_bytes(&data)
        .map_err(|e| format!("Failed to parse replay: {:?}", e))?;

    // Extract comprehensive match information
    let map = replay.map();
    let map_name = map.filename().to_string();
    let map_filename = if !map.localized_name_id().is_empty() {
        format!("id_{}", map.localized_name_id())
    } else {
        map_name.clone()
    };
    
    // Duration and timing information
    let duration_ticks = replay.length() as u32;
    let duration_seconds = (duration_ticks / TICKS_PER_SECOND) as u32;
    
    // Match metadata
    let game_version = Some(replay.version());
    let timestamp = Some(replay.timestamp().to_string());
    let game_type = Some(format!("{:?}", replay.game_type()));
    let matchhistory_id = replay.matchhistory_id().map(|id| id.to_string());
    
    // Extract teams and players with enhanced information
    let mut teams = Vec::new();
    let mut players_with_commands = Vec::new();
    
    let player_data = replay.players();
    let players_list: Vec<_> = player_data.iter().enumerate().collect();
    
    // Group players by their actual team ID
    let mut team_map: HashMap<u32, Vec<PlayerInfo>> = HashMap::new();
    
    for (idx, player) in &players_list {
        // Extract team ID (using debug output as vault library doesn't expose team ID directly)
        let team_debug = format!("{:?}", player.team());
        let team_id = extract_team_id_from_debug(&team_debug, *idx);
        
        // Extract faction information
        let faction = Some(format!("{:?}", player.faction()));
        let is_human = player.human();
        let steam_id = player.steam_id().map(|id| id.to_string());
        let profile_id = player.profile_id().map(|id| id.to_string());
        
        let player_info = PlayerInfo {
            player_id: *idx as u32,
            player_name: player.name().to_string(),
            team_id,
            faction: faction.clone(),
            is_human,
            steam_id: steam_id.clone(),
            profile_id: profile_id.clone(),
        };
        
        team_map.entry(team_id).or_insert_with(Vec::new).push(player_info);
        
        // Extract commands for this player using the provided filter
        let all_commands = extract_all_commands(&replay, *idx);
        let filtered_commands = all_commands.into_iter()
            .filter(|cmd| command_filter.should_include_command(&cmd.command_type))
            .collect();
        
        let player_with_commands = Player {
            player_id: *idx as u32,
            player_name: player.name().to_string(),
            team_id,
            faction: faction.clone(),
            is_human,
            steam_id: steam_id.clone(),
            profile_id: profile_id.clone(),
            battlegroup_id: None, // Add this field
            commands: extract_all_commands(&replay, *idx),
            build_commands: filtered_commands,
            chat_messages: extract_player_messages(&replay, *idx),
        };
        
        players_with_commands.push(player_with_commands);
    }
    
    // Create team structures
    for (team_id, player_list) in team_map {
        teams.push(Team {
            team_id,
            players: player_list,
        });
    }
    
    // Sort teams by ID for consistent output
    teams.sort_by_key(|t| t.team_id);
    
    // Extract match outcome (placeholder - vault library doesn't expose this directly)
    let winning_team = None; // TODO: Implement winning team detection
    
    // Extract global messages
    let messages = extract_game_messages(&replay);
    
    Ok(ReplayData {
        success: true,
        error_message: None,
        map_name,
        map_filename,
        duration_seconds,
        duration_ticks,
        game_version,
        timestamp,
        game_type,
        matchhistory_id,
        teams,
        winning_team,
        players: players_with_commands,
        messages,
    })
}

fn parse_replay_full_internal(file_path: &str) -> Result<ReplayData, String> {
    // Use default filter (build commands only) for backwards compatibility
    let default_filter = CommandFilter::default();
    parse_replay_with_filter_internal(file_path, &default_filter)
}




// Helper function to extract team ID from debug output
fn extract_team_id_from_debug(team_debug: &str, player_index: usize) -> u32 {
    if team_debug.contains("1") {
        1
    } else if team_debug.contains("2") {
        2
    } else if team_debug.contains("3") {
        3
    } else if team_debug.contains("4") {
        4
    } else {
        // Fallback: assign teams based on player position
        if player_index < 2 { 1 } else { 2 }
    }
}

// Extract all commands (not just build commands)
fn extract_all_commands(replay: &vault::Replay, player_index: usize) -> Vec<Command> {
    let mut commands = Vec::new();
    
    let players = replay.players();
    if player_index >= players.len() {
        return commands;
    }
    
    let player = &players[player_index];
    let all_commands = player.commands();
    
    for (i, command) in all_commands.iter().enumerate() {
        let command_debug = format!("{:#?}", command);
        let (command_type, details, pbgid) = parse_command_simple(&command_debug);
        
        // Extract real timestamp from tick field in debug output
        let timestamp = extract_timestamp_from_details(&details).unwrap_or((i as u32) * MILLISECONDS_PER_TICK);
        
        commands.push(Command {
            timestamp,
            command_type,
            details,
            pbgid,
            unit_name: None,    // Will be resolved in Go
            building_name: None, // Will be resolved in Go
        });
    }
    
    commands
}


// Extract player-specific chat messages
fn extract_player_messages(replay: &vault::Replay, player_index: usize) -> Vec<GameMessage> {
    let mut messages = Vec::new();
    
    let players = replay.players();
    if player_index >= players.len() {
        return messages;
    }
    
    let player = &players[player_index];
    let player_messages = player.messages();
    
    for (i, message) in player_messages.iter().enumerate() {
        let message_debug = format!("{:?}", message);
        let content = extract_message_content(&message_debug);
        let timestamp = extract_timestamp_from_details(&message_debug).unwrap_or((i as u32) * MILLISECONDS_PER_TICK);
        
        messages.push(GameMessage {
            timestamp,
            player_id: Some(player_index as u32),
            content,
            message_type: "chat".to_string(),
        });
    }
    
    messages
}

// Extract global game messages
fn extract_game_messages(replay: &vault::Replay) -> Vec<GameMessage> {
    let mut messages = Vec::new();
    
    // Collect all player messages first
    let players = replay.players();
    for (player_idx, player) in players.iter().enumerate() {
        let player_messages = player.messages();
        for (msg_idx, message) in player_messages.iter().enumerate() {
            let message_debug = format!("{:?}", message);
            let content = extract_message_content(&message_debug);
            let timestamp = extract_timestamp_from_details(&message_debug).unwrap_or((msg_idx as u32) * MILLISECONDS_PER_TICK + (player_idx as u32) * 1000);
            
            messages.push(GameMessage {
                timestamp,
                player_id: Some(player_idx as u32),
                content,
                message_type: "chat".to_string(),
            });
        }
    }
    
    
    // Sort by timestamp
    messages.sort_by_key(|m| m.timestamp);
    
    messages
}



// Simple command parsing that extracts command type and PBGID
fn parse_command_simple(command_debug: &str) -> (String, String, Option<String>) {
    let command_type;
    let details = command_debug.to_string();
    
    
    // Extract PBGID
    let pbgid = extract_pbgid(command_debug);
    
    // Determine command type from debug output
    if command_debug.contains("BuildSquad") {
        command_type = "build_squad".to_string();
    } else if command_debug.contains("ConstructEntity") || 
              command_debug.contains("PlaceAndConstructEntities") ||
              command_debug.contains("BuildStructure") ||
              command_debug.contains("SCMD_BuildStructure") {
        command_type = "construct_entity".to_string();
    } else if command_debug.contains("BuildGlobalUpgrade") ||
              command_debug.contains("TentativeUpgradePurchaseAll") ||
              command_debug.contains("SCMD_Upgrade") {
        command_type = "build_global_upgrade".to_string();
    } else if command_debug.contains("UseAbility") || command_debug.contains("SCMD_Ability") {
        command_type = "use_ability".to_string();
    } else if command_debug.contains("UseBattlegroupAbility") {
        command_type = "use_battlegroup_ability".to_string();
    } else if command_debug.contains("SelectBattlegroup") {
        command_type = "select_battlegroup".to_string();
    } else if command_debug.contains("SelectBattlegroupAbility") {
        command_type = "select_battlegroup_ability".to_string();
    } else if command_debug.contains("CancelConstruction") {
        command_type = "cancel_construction".to_string();
    } else if command_debug.contains("CancelProduction") || command_debug.contains("SCMD_CancelProduction") {
        command_type = "cancel_production".to_string();
    } else if command_debug.contains("AITakeover") {
        command_type = "ai_takeover".to_string();
    } else {
        command_type = "unknown".to_string();
    }
    
    
    (command_type, details, pbgid)
}

// Extract PBGID from command debug output
fn extract_pbgid(command_debug: &str) -> Option<String> {
    // Look for pbgid: pattern
    if let Some(start) = command_debug.find("pbgid:") {
        let substring = &command_debug[start + 6..];
        if let Some(end) = substring.find(',').or_else(|| substring.find('}')) {
            return Some(substring[..end].trim().to_string());
        }
    }
    
    // Look for Pbgid(value) pattern
    if let Some(start) = command_debug.find("Pbgid(") {
        let substring = &command_debug[start + 6..];
        if let Some(end) = substring.find(')') {
            return Some(substring[..end].trim().to_string());
        }
    }
    
    None
}



// Extract message content from message debug output
fn extract_message_content(message_debug: &str) -> String {
    // Try to extract the actual message content from debug output
    if let Some(start) = message_debug.find("content:") {
        let substring = &message_debug[start + 8..];
        if let Some(end) = substring.find(',').or_else(|| substring.find('}')) {
            let content = substring[..end].trim().trim_matches('"');
            return content.to_string();
        }
    }
    
    // Fallback to showing the debug output (truncated)
    if message_debug.len() > 50 {
        format!("Message: {}", &message_debug[..47])
    } else {
        format!("Message: {}", message_debug)
    }
}

// Configuration for command filtering
#[derive(Debug, Clone)]
pub struct CommandFilter {
    pub include_build_squad: bool,
    pub include_construct_entity: bool,
    pub include_build_global_upgrade: bool,
    pub include_use_ability: bool,
    pub include_use_battlegroup_ability: bool,
    pub include_select_battlegroup: bool,
    pub include_select_battlegroup_ability: bool,
    pub include_cancel_construction: bool,
    pub include_cancel_production: bool,
    pub include_ai_takeover: bool,
    pub include_unknown: bool,
}

impl Default for CommandFilter {
    fn default() -> Self {
        Self {
            include_build_squad: true,
            include_construct_entity: true,
            include_build_global_upgrade: true,
            include_use_ability: false,
            include_use_battlegroup_ability: false,
            include_select_battlegroup: true,
            include_select_battlegroup_ability: true,
            include_cancel_construction: false,
            include_cancel_production: false,
            include_ai_takeover: false,
            include_unknown: false,
        }
    }
}

impl CommandFilter {
    pub fn new_build_only() -> Self {
        Self::default()
    }
    
    pub fn new_all_commands() -> Self {
        Self {
            include_build_squad: true,
            include_construct_entity: true,
            include_build_global_upgrade: true,
            include_use_ability: true,
            include_use_battlegroup_ability: true,
            include_select_battlegroup: true,
            include_select_battlegroup_ability: true,
            include_cancel_construction: true,
            include_cancel_production: true,
            include_ai_takeover: true,
            include_unknown: true,
        }
    }
    
    pub fn new_combat_only() -> Self {
        Self {
            include_build_squad: false,
            include_construct_entity: false,
            include_build_global_upgrade: false,
            include_use_ability: true,
            include_use_battlegroup_ability: true,
            include_select_battlegroup: false,
            include_select_battlegroup_ability: false,
            include_cancel_construction: false,
            include_cancel_production: false,
            include_ai_takeover: false,
            include_unknown: false,
        }
    }
    
    fn should_include_command(&self, command_type: &str) -> bool {
        match command_type {
            "build_squad" => self.include_build_squad,
            "construct_entity" => self.include_construct_entity,
            "build_global_upgrade" => self.include_build_global_upgrade,
            "use_ability" => self.include_use_ability,
            "use_battlegroup_ability" => self.include_use_battlegroup_ability,
            "select_battlegroup" => self.include_select_battlegroup,
            "select_battlegroup_ability" => self.include_select_battlegroup_ability,
            "cancel_construction" => self.include_cancel_construction,
            "cancel_production" => self.include_cancel_production,
            "ai_takeover" => self.include_ai_takeover,
            "unknown" => self.include_unknown,
            _ => false,
        }
    }
}


// Extract real timestamp from command details by parsing the tick field
fn extract_timestamp_from_details(details: &str) -> Option<u32> {
    // Look for "tick: <number>" pattern in the debug output
    if let Some(start) = details.find("tick: ") {
        let substring = &details[start + 6..];
        if let Some(end) = substring.find(',').or_else(|| substring.find('}')) {
            let tick_str = substring[..end].trim();
            if let Ok(tick) = tick_str.parse::<u32>() {
                // Convert tick to milliseconds using the tick rate constant
                return Some(tick * MILLISECONDS_PER_TICK);
            }
        }
    }
    None
}

fn serialize_replay_data(result: &ReplayData) -> *mut c_char {
    match serde_json::to_string(result) {
        Ok(json) => match CString::new(json) {
            Ok(c_string) => c_string.into_raw(),
            Err(_) => ptr::null_mut(),
        },
        Err(_) => ptr::null_mut(),
    }
}