package lookup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DataResolver handles PBGID to friendly name resolution using coh3-data files
type DataResolver struct {
	locstrings     map[string]string
	sbpsData       map[string]interface{}
	ebpsData       map[string]interface{}
	battlegroupMap map[uint32]string
	upgradeMap     map[uint32]string
	dataDir        string
}

// UnitInfo represents resolved unit information
type UnitInfo struct {
	Name        string `json:"name"`
	Faction     string `json:"faction"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// NewDataResolver creates a new resolver instance
func NewDataResolver(dataDir string) (*DataResolver, error) {
	resolver := &DataResolver{
		dataDir:        dataDir,
		locstrings:     make(map[string]string),
		sbpsData:       make(map[string]interface{}),
		ebpsData:       make(map[string]interface{}),
		battlegroupMap: make(map[uint32]string),
		upgradeMap:     make(map[uint32]string),
	}

	if err := resolver.loadData(); err != nil {
		return nil, fmt.Errorf("failed to load data: %w", err)
	}

	return resolver, nil
}

// loadData loads all necessary data files
func (r *DataResolver) loadData() error {
	// Load English localization strings
	locPath := filepath.Join(r.dataDir, "locales", "en-locstring.json")
	if err := r.loadJSONFile(locPath, &r.locstrings); err != nil {
		// Fallback to generic locstring.json if en-locstring.json doesn't exist
		fallbackPath := filepath.Join(r.dataDir, "locstring.json")
		if err2 := r.loadJSONFile(fallbackPath, &r.locstrings); err2 != nil {
			return fmt.Errorf("failed to load locstrings from both %s and %s: %w, %w", locPath, fallbackPath, err, err2)
		}
	}

	// Load squad blueprints (units)
	sbpsPath := filepath.Join(r.dataDir, "sbps.json")
	if err := r.loadJSONFile(sbpsPath, &r.sbpsData); err != nil {
		return fmt.Errorf("failed to load sbps: %w", err)
	}

	// Load entity blueprints (buildings)
	ebpsPath := filepath.Join(r.dataDir, "ebps.json")
	if err := r.loadJSONFile(ebpsPath, &r.ebpsData); err != nil {
		return fmt.Errorf("failed to load ebps: %w", err)
	}

	// Load battlegroup mappings
	r.loadBattlegroupMappings()
	
	// Load upgrade mappings
	r.loadUpgradeMappings()

	return nil
}

// loadJSONFile loads a JSON file into the provided interface
func (r *DataResolver) loadJSONFile(path string, dest interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// ResolvePBGID resolves a PBGID to friendly unit information
func (r *DataResolver) ResolvePBGID(pbgid uint32) (*UnitInfo, error) {
	pbgidStr := strconv.FormatUint(uint64(pbgid), 10)

	// First try to find in squad blueprints (units)
	if unitInfo := r.findInSBPS(pbgidStr); unitInfo != nil {
		return unitInfo, nil
	}

	// Then try entity blueprints (buildings)
	if buildingInfo := r.findInEBPS(pbgidStr); buildingInfo != nil {
		return buildingInfo, nil
	}

	return nil, fmt.Errorf("PBGID %d not found in data files", pbgid)
}

// findInSBPS searches for a PBGID in squad blueprint data
func (r *DataResolver) findInSBPS(pbgid string) *UnitInfo {
	races, ok := r.sbpsData["races"].(map[string]interface{})
	if !ok {
		return nil
	}

	factionMap := map[string]string{
		"afrika_korps":   "Afrika Korps",
		"american":       "US Forces",
		"british":        "British",
		"british_africa": "British",
		"german":         "Wehrmacht",
		"common":         "Common",
	}

	for factionKey, factionData := range races {
		faction, ok := factionData.(map[string]interface{})
		if !ok {
			continue
		}

		factionName := factionMap[factionKey]
		if factionName == "" {
			factionName = strings.Title(strings.ReplaceAll(factionKey, "_", " "))
		}

		// Search through all categories (infantry, vehicles, aircraft, etc.)
		for categoryKey, categoryData := range faction {
			category, ok := categoryData.(map[string]interface{})
			if !ok {
				continue
			}

			// Search through all units in this category
			for unitKey, unitData := range category {
				unit, ok := unitData.(map[string]interface{})
				if !ok {
					continue
				}

				// Check if this unit has the PBGID we're looking for
				if unitPBGID, exists := unit["pbgid"]; exists {
					if fmt.Sprintf("%.0f", unitPBGID) == pbgid {
						return r.extractUnitInfoFromSBPS(unitKey, unit, factionName, categoryKey)
					}
				}
			}
		}
	}

	return nil
}

// findInEBPS searches for a PBGID in entity blueprint data
func (r *DataResolver) findInEBPS(pbgid string) *UnitInfo {
	races, ok := r.ebpsData["races"].(map[string]interface{})
	if !ok {
		return nil
	}

	factionMap := map[string]string{
		"afrika_korps": "Afrika Korps",
		"american":     "US Forces",
		"british":      "British",
		"german":       "Wehrmacht",
	}

	for factionKey, factionData := range races {
		faction, ok := factionData.(map[string]interface{})
		if !ok {
			continue
		}

		factionName := factionMap[factionKey]
		if factionName == "" {
			factionName = strings.Title(strings.ReplaceAll(factionKey, "_", " "))
		}

		for entityKey, entityData := range faction {
			entity, ok := entityData.(map[string]interface{})
			if !ok {
				continue
			}

			// Check if this entity has the PBGID we're looking for
			if entityPBGID, exists := entity["pbgid"]; exists {
				if fmt.Sprintf("%.0f", entityPBGID) == pbgid {
					return r.extractUnitInfo(entityKey, entity, factionName, "Building")
				}
			}
		}
	}

	return nil
}

// extractUnitInfoFromSBPS extracts unit information from SBPS data structure
func (r *DataResolver) extractUnitInfoFromSBPS(key string, data map[string]interface{}, faction, category string) *UnitInfo {
	info := &UnitInfo{
		Faction:  faction,
		Category: strings.Title(category),
	}

	// Navigate through the complex SBPS structure to find UI info
	if extensions, ok := data["extensions"].([]interface{}); ok {
		for _, ext := range extensions {
			if extMap, ok := ext.(map[string]interface{}); ok {
				if squadexts, ok := extMap["squadexts"].(map[string]interface{}); ok {
					// Look for race_list with UI info
					if raceList, ok := squadexts["race_list"].([]interface{}); ok {
						for _, race := range raceList {
							if raceData, ok := race.(map[string]interface{}); ok {
								if raceInfo, ok := raceData["race_data"].(map[string]interface{}); ok {
									if infoData, ok := raceInfo["info"].(map[string]interface{}); ok {
										// Extract screen name (localization ID)
										if screenName, ok := infoData["screen_name"].(map[string]interface{}); ok {
											if locstring, ok := screenName["locstring"].(map[string]interface{}); ok {
												if nameID, ok := locstring["value"].(string); ok {
													if localizedName, found := r.locstrings[nameID]; found {
														info.Name = localizedName
													}
												}
											}
										}

										// Extract help text (description)
										if helpText, ok := infoData["help_text"].(map[string]interface{}); ok {
											if locstring, ok := helpText["locstring"].(map[string]interface{}); ok {
												if descID, ok := locstring["value"].(string); ok && descID != "0" {
													if localizedDesc, found := r.locstrings[descID]; found {
														info.Description = localizedDesc
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Fallback to key-based name if no localized name found
	if info.Name == "" {
		info.Name = strings.Title(strings.ReplaceAll(key, "_", " "))
	}

	// Set more specific category based on the data category
	switch category {
	case "vehicles":
		info.Category = "Vehicle"
	case "infantry":
		info.Category = "Infantry"
	case "aircraft":
		info.Category = "Aircraft"
	case "emplacements", "team_weapons":
		info.Category = "Support"
	default:
		info.Category = "Unit"
	}

	return info
}

// extractUnitInfo extracts unit information from the data structure (legacy function)
func (r *DataResolver) extractUnitInfo(key string, data map[string]interface{}, faction, defaultCategory string) *UnitInfo {
	info := &UnitInfo{
		Faction:  faction,
		Category: defaultCategory,
	}

	// Try to get localized name
	if uiInfo, ok := data["ui_info"].(map[string]interface{}); ok {
		if screenNameID, exists := uiInfo["screen_name_id"]; exists {
			if nameID := fmt.Sprintf("%.0f", screenNameID); nameID != "" {
				if localizedName, found := r.locstrings[nameID]; found {
					info.Name = localizedName
				}
			}
		}

		if helpTextID, exists := uiInfo["help_text_id"]; exists {
			if descID := fmt.Sprintf("%.0f", helpTextID); descID != "" {
				if localizedDesc, found := r.locstrings[descID]; found {
					info.Description = localizedDesc
				}
			}
		}
	}

	// Fallback to key-based name if no localized name found
	if info.Name == "" {
		info.Name = strings.Title(strings.ReplaceAll(key, "_", " "))
	}

	// Determine category from key patterns
	keyLower := strings.ToLower(key)
	if strings.Contains(keyLower, "vehicle") || strings.Contains(keyLower, "tank") || strings.Contains(keyLower, "halftrack") {
		info.Category = "Vehicle"
	} else if strings.Contains(keyLower, "engineer") || strings.Contains(keyLower, "pioneer") {
		info.Category = "Engineer"
	} else if strings.Contains(keyLower, "gun") || strings.Contains(keyLower, "mortar") || strings.Contains(keyLower, "mg") {
		info.Category = "Support"
	}

	return info
}

// GetFriendlyName returns just the friendly name for a PBGID
func (r *DataResolver) GetFriendlyName(pbgid uint32) string {
	if info, err := r.ResolvePBGID(pbgid); err == nil {
		return info.Name
	}
	return ""
}

// GetBattlegroupName returns the battlegroup name for a PBGID
func (r *DataResolver) GetBattlegroupName(pbgid uint32) string {
	if name, exists := r.battlegroupMap[pbgid]; exists {
		return name
	}
	return ""
}

// GetUpgradeName returns the upgrade name for a PBGID
func (r *DataResolver) GetUpgradeName(pbgid uint32) string {
	if name, exists := r.upgradeMap[pbgid]; exists {
		return name
	}
	return ""
}

// loadBattlegroupMappings initializes battlegroup PBGID to name mappings
func (r *DataResolver) loadBattlegroupMappings() {
	// Afrika Korps battlegroups (from battlegroup.json)
	r.battlegroupMap[2075338] = "Armored Support"
	r.battlegroupMap[2074237] = "Italian Combined Arms"
	r.battlegroupMap[2072429] = "Italian Infantry"
	r.battlegroupMap[2164392] = "Panzerjäger Kommand"
	r.battlegroupMap[2151628] = "Subterfuge"

	// US Forces battlegroups (from battlegroup.json)
	r.battlegroupMap[199102] = "Airborne"
	r.battlegroupMap[199103] = "Armored"
	r.battlegroupMap[199104] = "Infantry"
	r.battlegroupMap[201151] = "Special Operations"

	// British battlegroups (from battlegroup.json)
	r.battlegroupMap[2164585] = "Special Weapons"
	r.battlegroupMap[2031369] = "Australian Defense"
	r.battlegroupMap[222365] = "British Air and Sea"
	r.battlegroupMap[202334] = "British Armored"
	r.battlegroupMap[2164115] = "Canadian Shock"
	r.battlegroupMap[201661] = "Indian Artillery"

	// Wehrmacht battlegroups (from battlegroup.json)
	r.battlegroupMap[199106] = "Breakthrough"
	r.battlegroupMap[2033170] = "Coastal"
	r.battlegroupMap[200769] = "Defense"
	r.battlegroupMap[199091] = "Luftwaffe"
	r.battlegroupMap[199105] = "Mechanized"
	r.battlegroupMap[2163770] = "Terror"

	// However, the SelectBattlegroup commands in replays use different PBGIDs than the main battlegroup
	// definitions. These are the actual PBGIDs seen in SelectBattlegroup commands and player battlegroup assignments:
	
	// US Forces battlegroups (actual replay PBGIDs)
	r.battlegroupMap[196934] = "Armored"
	
	// Wehrmacht battlegroups (actual replay PBGIDs)  
	r.battlegroupMap[198405] = "Wehrmacht Battlegroup 1"
	r.battlegroupMap[197799] = "Wehrmacht Battlegroup 2"
	
	// Afrika Korps battlegroups (actual replay PBGIDs)
	r.battlegroupMap[2164378] = "Afrika Korps Battlegroup"
	
	// British battlegroups (actual replay PBGIDs)
	r.battlegroupMap[2164107] = "British Battlegroup 1"
	r.battlegroupMap[2031370] = "British Battlegroup 2"
	
	// US Forces battlegroup selections (actual PBGIDs from replays)
	r.battlegroupMap[196934] = "Armored (US)" // Americans - Armored
	
	// Wehrmacht battlegroup selections (actual PBGIDs from replays)  
	r.battlegroupMap[198405] = "Unknown Wehrmacht BG 1" // Wehrmacht - Unknown
	r.battlegroupMap[197799] = "Unknown Wehrmacht BG 2" // Wehrmacht - Unknown
	
	// Afrika Korps battlegroup selections (actual PBGIDs from replays)
	r.battlegroupMap[2164378] = "Unknown Afrika Korps BG" // AfrikaKorps - Unknown
	
	// British battlegroup selections (actual PBGIDs from replays)
	r.battlegroupMap[2164107] = "Unknown British BG 1" // British - Unknown
	r.battlegroupMap[2031370] = "Unknown British BG 2" // British - Unknown
	
	// TODO: Map these to the correct battlegroup names by analyzing the battlegroup abilities used
}

// loadUpgradeMappings initializes upgrade PBGID to name mappings
func (r *DataResolver) loadUpgradeMappings() {
	// These are actual upgrade PBGIDs found in replay files with their real names
	// based on the upgrade.json data structure
	
	// Afrika Korps upgrades
	r.upgradeMap[2072101] = "T1 Unit Unlock (Afrika Korps)"
	r.upgradeMap[2072102] = "T2 Unit Unlock (Afrika Korps)"
	r.upgradeMap[2108279] = "Armored Assault Tactics (Afrika Korps)"
	r.upgradeMap[2084237] = "Vehicle Survivability Self-Repair (Afrika Korps)"
	r.upgradeMap[2084216] = "Operational Blitzkrieg (Afrika Korps)"
	r.upgradeMap[2084214] = "Smoke Survivability (Afrika Korps)"
	
	// British upgrades
	r.upgradeMap[197637] = "Bishop Squad Unlock (British)"
	r.upgradeMap[197636] = "Stuart Squad Unlock (British)"
	r.upgradeMap[197635] = "Rifle Grenade Tommy (British)"
	r.upgradeMap[197638] = "17-pounder Squad Unlock (British)"
	r.upgradeMap[2072354] = "Grant Tank Unlock (British)"
	
	// British Africa upgrades
	r.upgradeMap[2082737] = "Training Center Infantry (British Africa)"
	r.upgradeMap[2082738] = "Training Center Team Weapons (British Africa)"
	
	// Wehrmacht upgrades
	r.upgradeMap[170742] = "Medical Station (Wehrmacht)"
	r.upgradeMap[2081888] = "Panzer Kompanie Veterancy (Wehrmacht)"
	r.upgradeMap[2081886] = "Panzergrenadier Kompanie Veterancy (Wehrmacht)"
	r.upgradeMap[2089293] = "Side Skirts Global (Wehrmacht)"
	r.upgradeMap[2140327] = "Medical Bunker Defense (Wehrmacht)"
	r.upgradeMap[201588] = "Advanced Mechanical Assault Tactics (Wehrmacht)"
	r.upgradeMap[205683] = "Repair Bunker Defense (Wehrmacht)"
	
	// Note: Many upgrade commands appear as "Unknown" with action types:
	// - PCMD_TentativeUpgradePurchaseAll: Tentative upgrade purchase (UI interaction)
	// - SCMD_Upgrade: Actual upgrade application to specific entities
	// Only BuildGlobalUpgrade commands have clear PBGID mappings
}