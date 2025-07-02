use std::env;
use std::fs;

// Tick rate constant: each tick represents 0.125 seconds (8 ticks per second)
const SECONDS_PER_TICK: f64 = 0.125;
const TICKS_PER_SECOND: u32 = 8;

fn main() {
    let args: Vec<String> = env::args().collect();
    
    if args.len() != 2 {
        eprintln!("Usage: {} <replay_file.rec>", args[0]);
        std::process::exit(1);
    }
    
    let file_path = &args[1];
    
    println!("=== VAULT LIBRARY RAW DEBUG OUTPUT ===");
    println!("Reading file: {}", file_path);
    
    // Read the replay file
    let data = match fs::read(file_path) {
        Ok(data) => data,
        Err(e) => {
            eprintln!("Failed to read file: {}", e);
            std::process::exit(1);
        }
    };
    
    println!("File size: {} bytes", data.len());
    
    // Parse with vault
    let replay = match vault::Replay::from_bytes(&data) {
        Ok(replay) => replay,
        Err(e) => {
            eprintln!("Failed to parse replay: {:?}", e);
            std::process::exit(1);
        }
    };
    
    println!("\n=== BASIC REPLAY INFO ===");
    println!("Version: {}", replay.version());
    println!("Length (ticks): {}", replay.length());
    println!("Duration (seconds): {}", replay.length() / TICKS_PER_SECOND as usize);
    println!("Duration (exact): {:.2} seconds", replay.length() as f64 * SECONDS_PER_TICK);
    println!("Timestamp: {}", replay.timestamp());
    println!("Game type: {:?}", replay.game_type());
    if let Some(id) = replay.matchhistory_id() {
        println!("Match history ID: {}", id);
    }
    
    // Map info
    let map = replay.map();
    println!("\n=== MAP INFO ===");
    println!("Map filename: {}", map.filename());
    println!("Map localized name ID: {}", map.localized_name_id());
    println!("Map debug: {:#?}", map);
    
    // Players
    let players = replay.players();
    println!("\n=== PLAYERS ({}) ===", players.len());
    
    for (i, player) in players.iter().enumerate() {
        println!("\n--- Player {} ---", i);
        println!("Name: {}", player.name());
        println!("Human: {}", player.human());
        println!("Faction: {:?}", player.faction());
        println!("Team: {:?}", player.team());
        
        if let Some(steam_id) = player.steam_id() {
            println!("Steam ID: {}", steam_id);
        }
        
        if let Some(profile_id) = player.profile_id() {
            println!("Profile ID: {}", profile_id);
        }
        
        if let Some(battlegroup) = player.battlegroup() {
            println!("Battlegroup: {:?}", battlegroup);
        }
        
        // Commands (first 5)
        let commands = player.commands();
        println!("Total commands: {}", commands.len());
        
        if !commands.is_empty() {
            println!("First 5 commands (raw debug):");
            for (j, command) in commands.iter().take(5).enumerate() {
                println!("  Command {}: {:#?}", j, command);
            }
        }
        
        // Messages (first 3)
        let messages = player.messages();
        println!("Total messages: {}", messages.len());
        
        if !messages.is_empty() {
            println!("First 3 messages (raw debug):");
            for (j, message) in messages.iter().take(3).enumerate() {
                println!("  Message {}: {:#?}", j, message);
            }
        }
    }
    
    println!("\n=== FULL REPLAY DEBUG (TRUNCATED) ===");
    let debug_output = format!("{:#?}", replay);
    println!("{}", debug_output);
    // let truncated_size = 8000;
    // let truncated = if debug_output.len() > truncated_size {
    //     format!("{}...\n[TRUNCATED - {} more characters]", &debug_output[..truncated_size], debug_output.len() - truncated_size)
    // } else {
    //     debug_output
    // };
    // println!("{}", truncated);
}