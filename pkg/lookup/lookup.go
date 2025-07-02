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
	locstrings map[string]string
	sbpsData   map[string]interface{}
	ebpsData   map[string]interface{}
	dataDir    string
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
		dataDir:    dataDir,
		locstrings: make(map[string]string),
		sbpsData:   make(map[string]interface{}),
		ebpsData:   make(map[string]interface{}),
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