use std::collections::HashMap;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UnitData {
    pub name: String,
    pub faction: Option<String>,
    pub category: String,
    pub description: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BuildingData {
    pub name: String,
    pub faction: Option<String>,
    pub category: String,
    pub description: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AbilityData {
    pub name: String,
    pub description: Option<String>,
}

pub struct GameDataLookup {
    units: HashMap<u32, UnitData>,
    buildings: HashMap<u32, BuildingData>,
    abilities: HashMap<u32, AbilityData>,
}

impl GameDataLookup {
    pub fn new() -> Self {
        let mut lookup = Self {
            units: HashMap::new(),
            buildings: HashMap::new(),
            abilities: HashMap::new(),
        };
        
        // Try to load from JSON files first, fallback to hardcoded data
        if lookup.load_from_json().is_err() {
            lookup.initialize_hardcoded_data();
        }
        lookup
    }
    
    fn load_from_json(&mut self) -> Result<(), Box<dyn std::error::Error>> {
        // Try to load units from JSON file
        if let Ok(units_json) = std::fs::read_to_string("data/units.json") {
            let units_map: HashMap<String, UnitData> = serde_json::from_str(&units_json)?;
            for (key, value) in units_map {
                if let Ok(pbgid) = key.parse::<u32>() {
                    self.units.insert(pbgid, value);
                }
            }
        }
        
        // Try to load buildings from JSON file (if it exists)
        if let Ok(buildings_json) = std::fs::read_to_string("data/buildings.json") {
            let buildings_map: HashMap<String, BuildingData> = serde_json::from_str(&buildings_json)?;
            for (key, value) in buildings_map {
                if let Ok(pbgid) = key.parse::<u32>() {
                    self.buildings.insert(pbgid, value);
                }
            }
        }
        
        // Try to load abilities from JSON file (if it exists)
        if let Ok(abilities_json) = std::fs::read_to_string("data/abilities.json") {
            let abilities_map: HashMap<String, AbilityData> = serde_json::from_str(&abilities_json)?;
            for (key, value) in abilities_map {
                if let Ok(pbgid) = key.parse::<u32>() {
                    self.abilities.insert(pbgid, value);
                }
            }
        }
        
        Ok(())
    }
    
    pub fn get_unit(&self, pbgid: u32) -> Option<&UnitData> {
        self.units.get(&pbgid)
    }
    
    pub fn get_building(&self, pbgid: u32) -> Option<&BuildingData> {
        self.buildings.get(&pbgid)
    }
    
    pub fn get_ability(&self, pbgid: u32) -> Option<&AbilityData> {
        self.abilities.get(&pbgid)
    }
    
    pub fn get_friendly_name(&self, pbgid: u32) -> Option<String> {
        if let Some(unit) = self.get_unit(pbgid) {
            return Some(unit.name.clone());
        }
        if let Some(building) = self.get_building(pbgid) {
            return Some(building.name.clone());
        }
        if let Some(ability) = self.get_ability(pbgid) {
            return Some(ability.name.clone());
        }
        None
    }
    
    fn initialize_hardcoded_data(&mut self) {
        // Based on the replay data we observed, add common PBGIDs
        // These are hardcoded mappings that we'll eventually replace with downloaded data
        
        // Wehrmacht units (observed PBGIDs: 198355, 198340, 198347, 198342, etc.)
        self.units.insert(198355, UnitData {
            name: "Grenadier Squad".to_string(),
            faction: Some("Wehrmacht".to_string()),
            category: "Infantry".to_string(),  
            description: Some("Basic infantry squad".to_string()),
        });
        
        self.units.insert(198340, UnitData {
            name: "Pioneer Squad".to_string(),
            faction: Some("Wehrmacht".to_string()),
            category: "Engineer".to_string(),
            description: Some("Engineer unit for construction and repair".to_string()),
        });
        
        self.units.insert(198347, UnitData {
            name: "MG42 Machine Gun Team".to_string(),
            faction: Some("Wehrmacht".to_string()),
            category: "Support".to_string(),
            description: Some("Heavy machine gun team".to_string()),
        });
        
        self.units.insert(198342, UnitData {
            name: "Mortar Team".to_string(),
            faction: Some("Wehrmacht".to_string()),
            category: "Support".to_string(),
            description: Some("Indirect fire support".to_string()),
        });
        
        self.units.insert(198341, UnitData {
            name: "Sniper".to_string(),
            faction: Some("Wehrmacht".to_string()),
            category: "Infantry".to_string(),
            description: Some("Long-range precision infantry".to_string()),
        });
        
        self.units.insert(198357, UnitData {
            name: "Assault Grenadier Squad".to_string(),
            faction: Some("Wehrmacht".to_string()),
            category: "Infantry".to_string(),
            description: Some("Close-combat infantry squad".to_string()),
        });
        
        self.units.insert(198410, UnitData {
            name: "Panzer IV".to_string(),
            faction: Some("Wehrmacht".to_string()),
            category: "Vehicle".to_string(),
            description: Some("Medium tank".to_string()),
        });
        
        // Afrika Korps units (observed PBGIDs: 2033664, 2075940)
        self.units.insert(2033664, UnitData {
            name: "Panzergrenadier Squad".to_string(),
            faction: Some("Afrika Korps".to_string()),
            category: "Infantry".to_string(),
            description: Some("Elite mechanized infantry".to_string()),
        });
        
        self.units.insert(2075940, UnitData {
            name: "8 Rad Armored Car".to_string(),
            faction: Some("Afrika Korps".to_string()),
            category: "Vehicle".to_string(),
            description: Some("Light armored reconnaissance vehicle".to_string()),
        });
        
        // British units (observed PBGIDs: 203604, 203611, 203787, 203607, etc.)
        self.units.insert(203604, UnitData {
            name: "Section".to_string(),
            faction: Some("British".to_string()),
            category: "Infantry".to_string(),
            description: Some("Basic infantry squad".to_string()),
        });
        
        self.units.insert(203611, UnitData {
            name: "Royal Engineer Section".to_string(),
            faction: Some("British".to_string()),
            category: "Engineer".to_string(),
            description: Some("Engineer unit for construction and repair".to_string()),
        });
        
        self.units.insert(203787, UnitData {
            name: "Vickers Machine Gun Team".to_string(),
            faction: Some("British".to_string()),
            category: "Support".to_string(),
            description: Some("Heavy machine gun team".to_string()),
        });
        
        self.units.insert(203607, UnitData {
            name: "Commando Section".to_string(),
            faction: Some("British".to_string()),
            category: "Infantry".to_string(),
            description: Some("Elite infiltration infantry".to_string()),
        });
        
        self.units.insert(203610, UnitData {
            name: "Sniper".to_string(),
            faction: Some("British".to_string()),
            category: "Infantry".to_string(),
            description: Some("Long-range precision infantry".to_string()),
        });
        
        self.units.insert(203790, UnitData {
            name: "17-pounder Anti-tank Gun".to_string(),
            faction: Some("British".to_string()),
            category: "Support".to_string(),
            description: Some("Heavy anti-tank gun".to_string()),
        });
        
        self.units.insert(203788, UnitData {
            name: "Churchill Tank".to_string(),
            faction: Some("British".to_string()),
            category: "Vehicle".to_string(),
            description: Some("Heavy tank".to_string()),
        });
        
        self.units.insert(203789, UnitData {
            name: "Crusader AA Tank".to_string(),
            faction: Some("British".to_string()),
            category: "Vehicle".to_string(),
            description: Some("Anti-aircraft tank".to_string()),
        });
        
        self.units.insert(224382, UnitData {
            name: "Sherman Firefly".to_string(),
            faction: Some("British".to_string()),
            category: "Vehicle".to_string(),
            description: Some("Tank destroyer variant of Sherman".to_string()),
        });
        
        // American units (observed PBGIDs: 168613, 137121, 137122, etc.)
        self.units.insert(168613, UnitData {
            name: "Rifleman Squad".to_string(),
            faction: Some("US Forces".to_string()),
            category: "Infantry".to_string(),
            description: Some("Basic infantry squad".to_string()),
        });
        
        self.units.insert(137121, UnitData {
            name: "Engineer Squad".to_string(),
            faction: Some("US Forces".to_string()),
            category: "Engineer".to_string(),
            description: Some("Engineer unit for construction and repair".to_string()),
        });
        
        self.units.insert(137122, UnitData {
            name: "Assault Engineer Squad".to_string(),
            faction: Some("US Forces".to_string()),
            category: "Engineer".to_string(),
            description: Some("Combat engineer squad".to_string()),
        });
        
        self.units.insert(168619, UnitData {
            name: "Bazooka Team".to_string(),
            faction: Some("US Forces".to_string()),
            category: "Support".to_string(),
            description: Some("Anti-tank rocket team".to_string()),
        });
        
        self.units.insert(170304, UnitData {
            name: ".30 Cal Machine Gun Team".to_string(),
            faction: Some("US Forces".to_string()),
            category: "Support".to_string(),
            description: Some("Heavy machine gun team".to_string()),
        });
        
        self.units.insert(170321, UnitData {
            name: "Sherman Tank".to_string(),
            faction: Some("US Forces".to_string()),
            category: "Vehicle".to_string(),
            description: Some("Medium tank".to_string()),
        });
        
        self.units.insert(170305, UnitData {
            name: "M8 Greyhound".to_string(),
            faction: Some("US Forces".to_string()),
            category: "Vehicle".to_string(),
            description: Some("Light armored car".to_string()),
        });
        
        self.units.insert(170315, UnitData {
            name: "M3 Halftrack".to_string(),
            faction: Some("US Forces".to_string()),
            category: "Vehicle".to_string(),
            description: Some("Armored personnel carrier".to_string()),
        });
    }
}

impl Default for GameDataLookup {
    fn default() -> Self {
        Self::new()
    }
}