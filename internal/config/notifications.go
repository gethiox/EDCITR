package config

type NotificationSelection struct {
	WhoForUser         string
	DescriptionForUser string

	BaseLevel Level
	Who       string
	MsgTypes  []MsgType
}

var LocalNotificationsAvailable = []NotificationSelection{ // who, msgType
	{
		WhoForUser: "Escort", DescriptionForUser: "Escort", Who: "Escort", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "EscortJump", DescriptionForUser: "EscortJump", Value: "EscortJump", Level: LevelUseless},
		},
	},
	{
		WhoForUser: "COMMS", DescriptionForUser: "COMMS", Who: "COMMS", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "entered", DescriptionForUser: "entered", Value: "entered", Level: LevelUseless},
		},
	},
	{
		WhoForUser: "DockingPadBlockHostileComms", DescriptionForUser: "DockingPadBlockHostileComms", Who: "DockingPadBlockHostileComms", BaseLevel: LevelCritical,
		MsgTypes: []MsgType{}, // station starts firing
	},
	{
		WhoForUser: "DockingDoorBlockWarningComms", DescriptionForUser: "DockingDoorBlockWarningComms", Who: "DockingDoorBlockWarningComms", BaseLevel: LevelWarning,
		MsgTypes: []MsgType{},
	},
	{
		WhoForUser: "DockingTresspassHostileComms", DescriptionForUser: "DockingTresspassHostileComms", Who: "DockingTresspassHostileComms", BaseLevel: LevelWarning,
		MsgTypes: []MsgType{},
	},

	{
		WhoForUser: "STATION", DescriptionForUser: "STATION", Who: "STATION", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "NoFireZone_entered", DescriptionForUser: "NoFireZone_entered", Value: "NoFireZone_entered", Level: LevelUseless},
			{ValueForUser: "NoFireZone_entered_deployed", DescriptionForUser: "NoFireZone_entered_deployed", Value: "NoFireZone_entered_deployed", Level: LevelWarning}, // asks to stow hardpoints
			{ValueForUser: "NoFireZone_exited", DescriptionForUser: "NoFireZone_exited", Value: "NoFireZone_exited", Level: LevelUseless},
			{ValueForUser: "docking_granted", DescriptionForUser: "docking_granted", Value: "docking_granted", Level: LevelGood},
			{ValueForUser: "docking_denied", DescriptionForUser: "docking_denied", Value: "docking_denied", Level: LevelCritical},
			{ValueForUser: "docking_cancelled", DescriptionForUser: "docking_cancelled", Value: "docking_cancelled", Level: LevelInteresting},
			{ValueForUser: "docking_denied_toolarge", DescriptionForUser: "docking_denied_toolarge", Value: "docking_denied_toolarge", Level: LevelCritical},
			{ValueForUser: "docking_denied_accessRestricted", DescriptionForUser: "docking_denied_accessRestricted", Value: "docking_denied_accessRestricted", Level: LevelCritical},
			{ValueForUser: "docking_timeexpired", DescriptionForUser: "docking_timeexpired", Value: "docking_timeexpired", Level: LevelWarning},
			{ValueForUser: "docking_launchtimerexpire", DescriptionForUser: "docking_launchtimerexpire", Value: "docking_launchtimerexpire", Level: LevelWarning},
		},
	},
	{
		WhoForUser: "StationAggressorResponseMessage", DescriptionForUser: "StationAggressorResponseMessage", Who: "StationAggressorResponseMessage", BaseLevel: LevelCritical,
		MsgTypes: []MsgType{},
	},
	{
		WhoForUser: "DockingFailed", DescriptionForUser: "DockingFailed", Who: "DockingFailed", BaseLevel: LevelWarning,
		MsgTypes: []MsgType{
			{ValueForUser: "Distance", DescriptionForUser: "Distance", Value: "Distance", Level: LevelWarning}, // has some parameters
			{ValueForUser: "Hostile", DescriptionForUser: "Hostile", Value: "Hostile", Level: LevelWarning},
			{ValueForUser: "Offences", DescriptionForUser: "Offences", Value: "Offences", Level: LevelWarning}, // has some parameters
		},
	},
	{
		WhoForUser: "DockingChatter", DescriptionForUser: "DockingChatter", Who: "DockingChatter", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "Neutral", DescriptionForUser: "Neutral", Value: "Neutral", Level: LevelUseless},
			{ValueForUser: "Cordial", DescriptionForUser: "Cordial", Value: "Cordial", Level: LevelGood},
			{ValueForUser: "Allied", DescriptionForUser: "Allied", Value: "Allied", Level: LevelGood},
			{ValueForUser: "Friendly", DescriptionForUser: "Friendly", Value: "Friendly", Level: LevelGood},
			{ValueForUser: "Unfriendly", DescriptionForUser: "Unfriendly", Value: "Unfriendly", Level: LevelWarning},
		},
	},
	{
		WhoForUser: "DockingPadBlockWarningComms", DescriptionForUser: "DockingPadBlockWarningComms", Who: "DockingPadBlockWarningComms", BaseLevel: LevelWarning,
		MsgTypes: []MsgType{},
	},
	{
		WhoForUser: "DockingOffenceCleared", DescriptionForUser: "DockingOffenceCleared", Who: "DockingOffenceCleared", BaseLevel: LevelGood,
		MsgTypes: []MsgType{},
	},
	{
		WhoForUser: "DockingTresspassWarningComms", DescriptionForUser: "DockingTresspassWarningComms", Who: "DockingTresspassWarningComms", BaseLevel: LevelWarning,
		MsgTypes: []MsgType{},
	},
	{
		WhoForUser: "HumanoidEmote", DescriptionForUser: "HumanoidEmote", Who: "HumanoidEmote", BaseLevel: LevelInteresting,
		MsgTypes: []MsgType{
			{ValueForUser: "DefaultMessage", DescriptionForUser: "DefaultMessage", Value: "DefaultMessage", Level: LevelInteresting}, // has some parameters
		},
	},
}
