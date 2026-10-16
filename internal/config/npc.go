package config

type NPCMessageSelection struct {
	WhoForUser         string
	DescriptionForUser string

	BaseLevel Level
	Who       string
	MsgTypes  []MsgType
}

var NPCMessagesAvailable = []NPCMessageSelection{
	{
		WhoForUser: "Escort", DescriptionForUser: "Escort", Who: "Escort", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "AttackRun", DescriptionForUser: "AttackRun", Value: "AttackRun", Level: LevelUseless},    // "Target in my heads up, locking on"
			{ValueForUser: "EscortJump", DescriptionForUser: "EscortJump", Value: "EscortJump", Level: LevelUseless}, // "Damn, Looks like I'm done here."
			{ValueForUser: "NearDeath", DescriptionForUser: "NearDeath", Value: "NearDeath", Level: LevelUseless},
		},
	},
	{
		WhoForUser: "Wreckage", DescriptionForUser: "Wreckage", Who: "Wreckage", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "SearchandRescue_AccidentalDamage", DescriptionForUser: "SearchandRescue_AccidentalDamage", Value: "SearchandRescue_AccidentalDamage", Level: LevelUseless},
		},
	},
	{
		WhoForUser: "Propagandist", DescriptionForUser: "Propagandist", Who: "Propagandist", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "Patrol", DescriptionForUser: "Patrol", Value: "Patrol", Level: LevelUseless},
		},
	},
	{
		WhoForUser: "Protester", DescriptionForUser: "Protester", Who: "Protester", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "Patrol", DescriptionForUser: "Patrol", Value: "Patrol", Level: LevelUseless},
		},
	},
	{
		WhoForUser: "StreetVendorTeaCriticalDamage", DescriptionForUser: "StreetVendorTeaCriticalDamage", Who: "StreetVendorTeaCriticalDamage", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{}, // "No, it can't be!"
	},
	{
		WhoForUser: "MisfortuneCriticalDamage", DescriptionForUser: "MisfortuneCriticalDamage", Who: "MisfortuneCriticalDamage", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{}, // "I'm not so easily beaten" or "Don't scratch the paint - it cost me a fortune!"
	},
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	{
		WhoForUser: "BadKarmaCriticalDamage", DescriptionForUser: "BadKarmaCriticalDamage", Who: "BadKarmaCriticalDamage", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{}, // "I've taken too many hits"
	},
	{
		WhoForUser: "BountyHunter", DescriptionForUser: "BountyHunter", Who: "BountyHunter", BaseLevel: LevelWarning,
		MsgTypes: []MsgType{
			{ValueForUser: "Attack", DescriptionForUser: "Attack", Value: "Attack", Level: LevelWarning},
			{ValueForUser: "HunterHostileSC_Relevant", DescriptionForUser: "HunterHostileSC_Relevant", Value: "HunterHostileSC_Relevant", Level: LevelWarning},
			{ValueForUser: "StartInterdiction", DescriptionForUser: "StartInterdiction", Value: "StartInterdiction", Level: LevelWarning},
		},
	},
	{
		WhoForUser: "Commuter", DescriptionForUser: "Commuter", Who: "Commuter", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "AuthorityScan", DescriptionForUser: "AuthorityScan", Value: "AuthorityScan", Level: LevelUseless}, // "on what grounds are you scanning this vessel, officer?"
			{ValueForUser: "Passthrough", DescriptionForUser: "Passthrough", Value: "Passthrough", Level: LevelUseless},       // "See the stars in style with Sol Starways"
			{ValueForUser: "HostileScan", DescriptionForUser: "HostileScan", Value: "HostileScan", Level: LevelUseless},       // "This is not the trading ship yuou are looking for."
		},
	},
	{
		WhoForUser: "ConvoyWedding", DescriptionForUser: "ConvoyWedding", Who: "ConvoyWedding", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "Patrol", DescriptionForUser: "Patrol", Value: "Patrol", Level: LevelUseless},
		},
	},
	{
		WhoForUser: "CruiseLiner", DescriptionForUser: "CruiseLiner", Who: "CruiseLiner", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "SCPatrol", DescriptionForUser: "SCPatrol", Value: "SCPatrol", Level: LevelUseless},
		},
	},
	{
		WhoForUser: "DistressCallOutofFuelPilot", DescriptionForUser: "DistressCallOutofFuelPilot", Who: "DistressCallOutofFuelPilot", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "Raided", DescriptionForUser: "Raided", Value: "Raided", Level: LevelUseless},
		},
	},
	{
		WhoForUser: "Explorer", DescriptionForUser: "Explorer", Who: "Explorer", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "Passthrough", DescriptionForUser: "Passthrough", Value: "Passthrough", Level: LevelUseless},
		},
	},
	{
		WhoForUser: "Hitman", DescriptionForUser: "Hitman", Who: "Hitman", BaseLevel: LevelWarning,
		MsgTypes: []MsgType{
			{ValueForUser: "HunterHostileSC_Relevant", DescriptionForUser: "HunterHostileSC_Relevant", Value: "HunterHostileSC_Relevant", Level: LevelWarning}, // interdiction annoucment
		},
	},
	{
		WhoForUser: "Insurgent", DescriptionForUser: "Insurgent", Who: "Insurgent", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "NearDeath", DescriptionForUser: "NearDeath", Value: "NearDeath", Level: LevelUseless}, // "You can't hope to stop us. We are everywhere"
		},
	},
	{
		WhoForUser: "Miner", DescriptionForUser: "Miner", Who: "Miner", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "OnJumpNoAsteroids", DescriptionForUser: "OnJumpNoAsteroids", Value: "OnJumpNoAsteroids", Level: LevelUseless}, // "Location running dry, heading home"
		},
	},
	{
		WhoForUser: "MinerCriticalDamage", DescriptionForUser: "MinerCriticalDamage", Who: "MinerCriticalDamage", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{}, // "No, no, nooooooooo!"
	},
	{
		WhoForUser: "Military", DescriptionForUser: "Military", Who: "Military", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "AccidentalDamage", DescriptionForUser: "AccidentalDamage", Value: "AccidentalDamage", Level: LevelWarning}, // "You're shooting at the wrong person, Commander!"
			{ValueForUser: "NearDeath", DescriptionForUser: "NearDeath", Value: "NearDeath", Level: LevelUseless},
			{ValueForUser: "Passthrough", DescriptionForUser: "Passthrough", Value: "Passthrough", Level: LevelUseless}, // "Securing the perimeter"
			{ValueForUser: "UnderFire", DescriptionForUser: "UnderFire", Value: "UnderFire", Level: LevelUseless},       // may be also Thargoid lol
		},
	},
	{
		WhoForUser: "Mission", DescriptionForUser: "Mission", Who: "Mission", BaseLevel: LevelInteresting,
		MsgTypes: []MsgType{
			{ValueForUser: "BountyHunter_Interdiction", DescriptionForUser: "BountyHunter_Interdiction", Value: "BountyHunter_Interdiction", Level: LevelWarning},
			{ValueForUser: "BountyHunter_PostInterdiction", DescriptionForUser: "BountyHunter_PostInterdiction", Value: "BountyHunter_PostInterdiction", Level: LevelUseless},
			{ValueForUser: "Hitman_Interdiction", DescriptionForUser: "Hitman_Interdiction", Value: "Hitman_Interdiction", Level: LevelWarning},
			{ValueForUser: "Collect_MessengerChat", DescriptionForUser: "Collect_MessengerChat", Value: "Collect_MessengerChat", Level: LevelInteresting}, // has many mission-related parameters "Goods required! Guaranteed to pay more than the going rate! It's an opportunity you'd be foolish to ignore!"
		},
	},
	{
		WhoForUser: "OverwatchCriticalDamage", DescriptionForUser: "OverwatchCriticalDamage", Who: "OverwatchCriticalDamage", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{},
	},
	{
		WhoForUser: "OverwatchAttackRun", DescriptionForUser: "OverwatchAttackRun", Who: "OverwatchAttackRun", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{}, // "lining up for the shot"
	},
	{
		WhoForUser: "PassengerLiner", DescriptionForUser: "PassengerLiner", Who: "PassengerLiner", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "Passthrough", DescriptionForUser: "Passthrough", Value: "Passthrough", Level: LevelUseless}, // "Passengers are advised that..."
		},
	},
	{
		WhoForUser: "Pirate", DescriptionForUser: "Pirate", Who: "Pirate", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "Arrival", DescriptionForUser: "Arrival", Value: "Arrival", Level: LevelWarning},                                                    // "Judgement time!"
			{ValueForUser: "HunterHostileSC_Relevant", DescriptionForUser: "HunterHostileSC_Relevant", Value: "HunterHostileSC_Relevant", Level: LevelWarning}, // interdiction annoucment
			{ValueForUser: "LargeCargoIgnoreThreat", DescriptionForUser: "LargeCargoIgnoreThreat", Value: "LargeCargoIgnoreThreat", Level: LevelWarning},       // when npc sees a nice cargo
			{ValueForUser: "NotEnoughCargo", DescriptionForUser: "NotEnoughCargo", Value: "NotEnoughCargo", Level: LevelWarning},                               // when npc expected more to jetson
			{ValueForUser: "OnAuthorityDetection", DescriptionForUser: "OnAuthorityDetection", Value: "OnAuthorityDetection", Level: LevelUseless},             // "This will spice things up"
			{ValueForUser: "OnDeclarePiracyAttack", DescriptionForUser: "OnDeclarePiracyAttack", Value: "OnDeclarePiracyAttack", Level: LevelWarning},          // "You choose death then"
			{ValueForUser: "OnEnemyShipDetection", DescriptionForUser: "OnEnemyShipDetection", Value: "OnEnemyShipDetection", Level: LevelUseless},             // "Hostiles? The hits just keep on coming"
			{ValueForUser: "OnHullLowJumpAway", DescriptionForUser: "OnHullLowJumpAway", Value: "OnHullLowJumpAway", Level: LevelUseless},
			{ValueForUser: "OnNoCargoFound", DescriptionForUser: "OnNoCargoFound", Value: "OnNoCargoFound", Level: LevelUseless},                // "My children will go hungry tonight"
			{ValueForUser: "OnPiratedCargoTheft", DescriptionForUser: "OnPiratedCargoTheft", Value: "OnPiratedCargoTheft", Level: LevelWarning}, // "That's mine!"
			{ValueForUser: "OnSatisfactoryCargoDropped", DescriptionForUser: "OnSatisfactoryCargoDropped", Value: "OnSatisfactoryCargoDropped", Level: LevelGood},
			{ValueForUser: "OnStartScanCargo", DescriptionForUser: "OnStartScanCargo", Value: "OnStartScanCargo", Level: LevelWarning},
			{ValueForUser: "OnTargetFleeing", DescriptionForUser: "OnTargetFleeing", Value: "OnTargetFleeing", Level: LevelWarning},
			{ValueForUser: "OnUnsatisfactoryCargoDropped", DescriptionForUser: "OnUnsatisfactoryCargoDropped", Value: "OnUnsatisfactoryCargoDropped", Level: LevelWarning},
			{ValueForUser: "ReminderSpecific", DescriptionForUser: "ReminderSpecific", Value: "ReminderSpecific", Level: LevelWarning},
			{ValueForUser: "ReminderUnits", DescriptionForUser: "ReminderUnits", Value: "ReminderUnits", Level: LevelWarning},
			{ValueForUser: "ReminderValue", DescriptionForUser: "ReminderValue", Value: "ReminderValue", Level: LevelWarning},
			{ValueForUser: "Special_TrapText", DescriptionForUser: "Special_TrapText", Value: "Special_TrapText", Level: LevelWarning},    // "I can't belive that worked! Get'em boys!"
			{ValueForUser: "Special_Arrival", DescriptionForUser: "Special_Arrival", Value: "Special_Arrival", Level: LevelWarning},       // "Watch it pal, this is our haul! Just be on your way and don't try anything stupid"
			{ValueForUser: "StartInterdiction", DescriptionForUser: "StartInterdiction", Value: "StartInterdiction", Level: LevelWarning}, // interdiction
			{ValueForUser: "ThreatenUnits", DescriptionForUser: "ThreatenUnits", Value: "ThreatenUnits", Level: LevelWarning},
			{ValueForUser: "ThreatenSpecific", DescriptionForUser: "ThreatenSpecific", Value: "ThreatenSpecific", Level: LevelWarning},
			{ValueForUser: "ThreatenTimer", DescriptionForUser: "ThreatenTimer", Value: "ThreatenTimer", Level: LevelWarning},
			{ValueForUser: "ThreatenValue", DescriptionForUser: "ThreatenValue", Value: "ThreatenValue", Level: LevelWarning},
			{ValueForUser: "UnprovokedAttack", DescriptionForUser: "UnprovokedAttack", Value: "UnprovokedAttack", Level: LevelWarning}, // "today is my lucky day"
		},
	},
	{
		WhoForUser: "PirateLord", DescriptionForUser: "PirateLord", Who: "PirateLord", BaseLevel: LevelWarning,
		MsgTypes: []MsgType{
			{ValueForUser: "PatrolComs", DescriptionForUser: "PatrolComs", Value: "PatrolComs", Level: LevelUseless},
		},
	},
	{
		WhoForUser: "Police", DescriptionForUser: "Police", Who: "Police", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "ArrivedAtInvestigationLocation", DescriptionForUser: "ArrivedAtInvestigationLocation", Value: "ArrivedAtInvestigationLocation", Level: LevelInteresting},
			{ValueForUser: "ArriveInvestigate", DescriptionForUser: "ArriveInvestigate", Value: "ArriveInvestigate", Level: LevelInteresting},
			{ValueForUser: "ArriveReinforcement", DescriptionForUser: "ArriveReinforcement", Value: "ArriveReinforcement", Level: LevelInteresting},
			{ValueForUser: "Attack", DescriptionForUser: "Attack", Value: "Attack", Level: LevelWarning},          // "Crime detected. Deadly force authorised"
			{ValueForUser: "EndPatrol", DescriptionForUser: "EndPatrol", Value: "EndPatrol", Level: LevelUseless}, // "Sift's over guys, I'm heading back to base"
			{ValueForUser: "StartInterdiction", DescriptionForUser: "StartInterdiction", Value: "StartInterdiction", Level: LevelWarning},
			{ValueForUser: "StartPatrol", DescriptionForUser: "StartPatrol", Value: "StartPatrol", Level: LevelUseless},
			{ValueForUser: "ThankYouPassedStopAndSearch", DescriptionForUser: "ThankYouPassedStopAndSearch", Value: "ThankYouPassedStopAndSearch", Level: LevelGood}, // "You're clean"
		},
	},
	{
		WhoForUser: "Rescuer", DescriptionForUser: "Rescuer", Who: "Rescuer", BaseLevel: LevelInteresting,
		MsgTypes: []MsgType{
			{ValueForUser: "OnRescueDone", DescriptionForUser: "OnRescueDone", Value: "OnRescueDone", Level: LevelGood},
			{ValueForUser: "OnRescueStart", DescriptionForUser: "OnRescueStart", Value: "OnRescueStart", Level: LevelInteresting},
		},
	},
	{
		WhoForUser: "Smuggler", DescriptionForUser: "Smuggler", Who: "Smuggler", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "NearDeath", DescriptionForUser: "NearDeath", Value: "NearDeath", Level: LevelUseless},
		},
	},
	{
		WhoForUser: "Trader", DescriptionForUser: "Trader", Who: "Trader", BaseLevel: LevelUseless,
		MsgTypes: []MsgType{
			{ValueForUser: "OnEnemyShipDetection", DescriptionForUser: "OnEnemyShipDetection", Value: "OnEnemyShipDetection", Level: LevelUseless}, // "We don't have to do this!"
			{ValueForUser: "AttackRun", DescriptionForUser: "AttackRun", Value: "AttackRun", Level: LevelUseless},                                  // "You think I'm going to let you steam from me?"
		},
	},
}
