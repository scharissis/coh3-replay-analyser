package commands

// CommandType represents the different types of commands that can be parsed from replays
type CommandType string

const (
	BuildSquad              CommandType = "build_squad"
	ConstructEntity         CommandType = "construct_entity"
	BuildGlobalUpgrade      CommandType = "build_global_upgrade"
	UseAbility              CommandType = "use_ability"
	UseBattlegroupAbility   CommandType = "use_battlegroup_ability"
	SelectBattlegroup       CommandType = "select_battlegroup"
	SelectBattlegroupAbility CommandType = "select_battlegroup_ability"
	CancelConstruction      CommandType = "cancel_construction"
	CancelProduction        CommandType = "cancel_production"
	AITakeover              CommandType = "ai_takeover"
	Unknown                 CommandType = "unknown"
)

// CommandCategory represents logical groupings of command types
type CommandCategory string

const (
	CategoryBuild   CommandCategory = "build"
	CategoryCombat  CommandCategory = "combat"
	CategoryControl CommandCategory = "control"
	CategoryCancel  CommandCategory = "cancel"
	CategoryOther   CommandCategory = "other"
)

// CommandDefinition defines properties and categorization for each command type
type CommandDefinition struct {
	Type        CommandType
	Category    CommandCategory
	Description string
	IsBuildable bool // Does this command create/build something?
	IsCombat    bool // Is this a combat-related action?
	IsEconomic  bool // Does this affect economy/resources?
}

// CommandDefinitions defines all known command types and their properties
// This is the single source of truth for command configuration
var CommandDefinitions = map[CommandType]CommandDefinition{
	BuildSquad: {
		Type:        BuildSquad,
		Category:    CategoryBuild,
		Description: "Build a squad/unit",
		IsBuildable: true,
		IsCombat:    false,
		IsEconomic:  true,
	},
	ConstructEntity: {
		Type:        ConstructEntity,
		Category:    CategoryBuild,
		Description: "Construct a building",
		IsBuildable: true,
		IsCombat:    false,
		IsEconomic:  true,
	},
	BuildGlobalUpgrade: {
		Type:        BuildGlobalUpgrade,
		Category:    CategoryBuild,
		Description: "Research a technology upgrade",
		IsBuildable: true,
		IsCombat:    false,
		IsEconomic:  true,
	},
	UseAbility: {
		Type:        UseAbility,
		Category:    CategoryCombat,
		Description: "Use a unit ability",
		IsBuildable: false,
		IsCombat:    true,
		IsEconomic:  false,
	},
	UseBattlegroupAbility: {
		Type:        UseBattlegroupAbility,
		Category:    CategoryCombat,
		Description: "Use a battlegroup ability",
		IsBuildable: false,
		IsCombat:    true,
		IsEconomic:  false,
	},
	SelectBattlegroup: {
		Type:        SelectBattlegroup,
		Category:    CategoryBuild,
		Description: "Select a battlegroup",
		IsBuildable: true,
		IsCombat:    false,
		IsEconomic:  true,
	},
	SelectBattlegroupAbility: {
		Type:        SelectBattlegroupAbility,
		Category:    CategoryBuild,
		Description: "Select a battlegroup ability",
		IsBuildable: true,
		IsCombat:    false,
		IsEconomic:  true,
	},
	CancelConstruction: {
		Type:        CancelConstruction,
		Category:    CategoryCancel,
		Description: "Cancel building construction",
		IsBuildable: false,
		IsCombat:    false,
		IsEconomic:  true,
	},
	CancelProduction: {
		Type:        CancelProduction,
		Category:    CategoryCancel,
		Description: "Cancel unit production",
		IsBuildable: false,
		IsCombat:    false,
		IsEconomic:  true,
	},
	AITakeover: {
		Type:        AITakeover,
		Category:    CategoryControl,
		Description: "AI takes control of player",
		IsBuildable: false,
		IsCombat:    false,
		IsEconomic:  false,
	},
	Unknown: {
		Type:        Unknown,
		Category:    CategoryOther,
		Description: "Unknown command type",
		IsBuildable: false,
		IsCombat:    false,
		IsEconomic:  false,
	},
}

// FilterPreset represents a named collection of command types to include
type FilterPreset struct {
	Name        string
	Description string
	Include     []CommandType
}

// Predefined filter presets for common use cases
var (
	BuildOnlyPreset = FilterPreset{
		Name:        "build",
		Description: "Units, buildings, upgrades, and battlegroup selections",
		Include: []CommandType{
			BuildSquad,
			ConstructEntity,
			BuildGlobalUpgrade,
			SelectBattlegroup,
			SelectBattlegroupAbility,
		},
	}

	CombatOnlyPreset = FilterPreset{
		Name:        "combat",
		Description: "Combat abilities and tactical actions",
		Include: []CommandType{
			UseAbility,
			UseBattlegroupAbility,
		},
	}

	EconomicPreset = FilterPreset{
		Name:        "economic",
		Description: "All economy-affecting commands",
		Include: GetCommandsByProperty(func(def CommandDefinition) bool {
			return def.IsEconomic
		}),
	}

	AllCommandsPreset = FilterPreset{
		Name:        "all",
		Description: "All command types",
		Include:     GetAllCommandTypes(),
	}
)

// Helper functions for developers to easily create custom filters

// GetAllCommandTypes returns all known command types
func GetAllCommandTypes() []CommandType {
	var types []CommandType
	for cmdType := range CommandDefinitions {
		types = append(types, cmdType)
	}
	return types
}

// GetCommandsByCategory returns all command types in a specific category
func GetCommandsByCategory(category CommandCategory) []CommandType {
	var types []CommandType
	for cmdType, def := range CommandDefinitions {
		if def.Category == category {
			types = append(types, cmdType)
		}
	}
	return types
}

// GetCommandsByProperty returns command types matching a custom property filter
func GetCommandsByProperty(filter func(CommandDefinition) bool) []CommandType {
	var types []CommandType
	for cmdType, def := range CommandDefinitions {
		if filter(def) {
			types = append(types, cmdType)
		}
	}
	return types
}

// CreateCustomFilter allows developers to easily create filters
func CreateCustomFilter(name, description string, commandTypes ...CommandType) FilterPreset {
	return FilterPreset{
		Name:        name,
		Description: description,
		Include:     commandTypes,
	}
}

// Examples of how developers can easily create custom filters:

// ExampleCustomFilters shows how developers can define their own filters
var ExampleCustomFilters = []FilterPreset{
	// Include only squad building and upgrades
	CreateCustomFilter("army_building", "Squad building and upgrades only",
		BuildSquad, BuildGlobalUpgrade),

	// Include all building/construction related commands
	CreateCustomFilter("construction", "All construction activities",
		ConstructEntity, CancelConstruction),

	// Include battlegroup-related commands only
	CreateCustomFilter("battlegroup", "Battlegroup selections and abilities",
		SelectBattlegroup, SelectBattlegroupAbility, UseBattlegroupAbility),

	// Custom filter using property-based selection
	{
		Name:        "combat_and_economic",
		Description: "Commands that affect combat or economy",
		Include: GetCommandsByProperty(func(def CommandDefinition) bool {
			return def.IsCombat || def.IsEconomic
		}),
	},
}