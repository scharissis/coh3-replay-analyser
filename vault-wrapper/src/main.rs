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
    
    println!("=== SEARCHING FOR UPGRADE COMMANDS ===");
    println!("Reading file: {}", file_path);
    
    // Read the replay file
    let data = match fs::read(file_path) {
        Ok(data) => data,
        Err(e) => {
            eprintln!("Failed to read file: {}", e);
            std::process::exit(1);
        }
    };
    
    // Parse with vault
    let replay = match vault::Replay::from_bytes(&data) {
        Ok(replay) => replay,
        Err(e) => {
            eprintln!("Failed to parse replay: {:?}", e);
            std::process::exit(1);
        }
    };
    
    // Players
    let players = replay.players();
    println!("\n=== SEARCHING FOR UPGRADE COMMANDS ===");
    
    for (i, player) in players.iter().enumerate() {
        println!("\n--- Player {} ({}) ---", i, player.name());
        
        let commands = player.commands();
        let mut upgrade_count = 0;
        
        for (j, command) in commands.iter().enumerate() {
            let command_debug = format!("{:#?}", command);
            
            // Check if this is an upgrade command
            if command_debug.contains("BuildGlobalUpgrade") || 
               command_debug.contains("TentativeUpgradePurchaseAll") ||
               command_debug.contains("SCMD_Upgrade") {
                upgrade_count += 1;
                println!("\n  === UPGRADE COMMAND {} ===", upgrade_count);
                println!("  Command index: {}", j);
                println!("  Raw debug output:");
                println!("{}", command_debug);
                println!("  === END UPGRADE COMMAND ===");
                
                // Extract and highlight the PBGID
                if let Some(pbgid) = extract_pbgid_from_debug(&command_debug) {
                    println!("  *** EXTRACTED PBGID: {} ***", pbgid);
                }
            }
        }
        
        println!("Found {} upgrade commands for player {}", upgrade_count, i);
    }
}

fn extract_pbgid_from_debug(command_debug: &str) -> Option<String> {
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