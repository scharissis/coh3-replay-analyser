package commands

import "github.com/scharissis/coh3-replay-analyser/vault"

// FilterConfig provides an easy way for developers to configure command filtering
type FilterConfig struct {
	preset *FilterPreset
	custom []CommandType
}

// NewFilterConfig creates a new filter configuration
func NewFilterConfig() *FilterConfig {
	return &FilterConfig{}
}

// WithPreset sets a predefined filter preset
func (c *FilterConfig) WithPreset(preset FilterPreset) *FilterConfig {
	c.preset = &preset
	return c
}

// WithCustom sets custom command types to include
func (c *FilterConfig) WithCustom(commandTypes ...CommandType) *FilterConfig {
	c.custom = commandTypes
	return c
}

// WithCategory includes all commands from a specific category
func (c *FilterConfig) WithCategory(category CommandCategory) *FilterConfig {
	c.custom = append(c.custom, GetCommandsByCategory(category)...)
	return c
}

// WithProperty includes commands matching a property filter
func (c *FilterConfig) WithProperty(filter func(CommandDefinition) bool) *FilterConfig {
	c.custom = append(c.custom, GetCommandsByProperty(filter)...)
	return c
}

// ToVaultFilter converts the configuration to a vault.CommandFilter
func (c *FilterConfig) ToVaultFilter() vault.CommandFilter {
	// Start with all false
	filter := vault.CommandFilter{}
	
	// Determine which commands to include
	var includeTypes []CommandType
	if c.preset != nil {
		includeTypes = c.preset.Include
	} else if len(c.custom) > 0 {
		includeTypes = c.custom
	} else {
		// Default to build commands
		includeTypes = BuildOnlyPreset.Include
	}
	
	// Set the appropriate flags
	for _, cmdType := range includeTypes {
		switch cmdType {
		case BuildSquad:
			filter.IncludeBuildSquad = true
		case ConstructEntity:
			filter.IncludeConstructEntity = true
		case BuildGlobalUpgrade:
			filter.IncludeBuildGlobalUpgrade = true
		case UseAbility:
			filter.IncludeUseAbility = true
		case UseBattlegroupAbility:
			filter.IncludeUseBattlegroupAbility = true
		case SelectBattlegroup:
			filter.IncludeSelectBattlegroup = true
		case SelectBattlegroupAbility:
			filter.IncludeSelectBattlegroupAbility = true
		case CancelConstruction:
			filter.IncludeCancelConstruction = true
		case CancelProduction:
			filter.IncludeCancelProduction = true
		case AITakeover:
			filter.IncludeAITakeover = true
		case Unknown:
			filter.IncludeUnknown = true
		}
	}
	
	return filter
}

// Quick access functions for common configurations

// BuildCommands returns a filter for build-related commands only
func BuildCommands() vault.CommandFilter {
	return NewFilterConfig().WithPreset(BuildOnlyPreset).ToVaultFilter()
}

// CombatCommands returns a filter for combat-related commands only
func CombatCommands() vault.CommandFilter {
	return NewFilterConfig().WithPreset(CombatOnlyPreset).ToVaultFilter()
}

// AllCommands returns a filter that includes all command types
func AllCommands() vault.CommandFilter {
	return NewFilterConfig().WithPreset(AllCommandsPreset).ToVaultFilter()
}

// EconomicCommands returns a filter for economy-affecting commands
func EconomicCommands() vault.CommandFilter {
	return NewFilterConfig().WithPreset(EconomicPreset).ToVaultFilter()
}

// OnlySquadBuilding returns a filter for squad building commands only
func OnlySquadBuilding() vault.CommandFilter {
	return NewFilterConfig().WithCustom(BuildSquad).ToVaultFilter()
}

// OnlyBuildings returns a filter for building construction commands only
func OnlyBuildings() vault.CommandFilter {
	return NewFilterConfig().WithCustom(ConstructEntity).ToVaultFilter()
}

// OnlyAbilities returns a filter for ability usage commands only
func OnlyAbilities() vault.CommandFilter {
	return NewFilterConfig().WithCustom(UseAbility, UseBattlegroupAbility).ToVaultFilter()
}

// BattlegroupRelated returns a filter for battlegroup-related commands
func BattlegroupRelated() vault.CommandFilter {
	return NewFilterConfig().WithCustom(
		SelectBattlegroup,
		SelectBattlegroupAbility,
		UseBattlegroupAbility,
	).ToVaultFilter()
}

// Example usage for developers:
/*

// Use predefined presets
filter := commands.BuildCommands()
filter := commands.CombatCommands()
filter := commands.AllCommands()

// Create custom filters easily
filter := commands.NewFilterConfig().
	WithCustom(commands.BuildSquad, commands.ConstructEntity).
	ToVaultFilter()

// Filter by category
filter := commands.NewFilterConfig().
	WithCategory(commands.CategoryBuild).
	ToVaultFilter()

// Filter by properties
filter := commands.NewFilterConfig().
	WithProperty(func(def commands.CommandDefinition) bool {
		return def.IsCombat || def.IsEconomic
	}).
	ToVaultFilter()

// Use the filter with vault parsing
data, err := vault.ParseReplayWithFilter(replayFile, dataDir, filter)

*/