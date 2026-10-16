package parser

import (
	"reflect"
	"time"
)

var EventRegistry = map[string]reflect.Type{
	// crucial for the app
	"Commander":         reflect.TypeFor[Commander](),
	"Continued":         reflect.TypeFor[Continued](),
	"Friends":           reflect.TypeFor[Friends](),
	"LoadGame":          reflect.TypeFor[LoadGame](),
	"ReceiveText":       reflect.TypeFor[ReceiveText](),
	"SendText":          reflect.TypeFor[SendText](),
	"WingInvite":        reflect.TypeFor[WingInvite](),
	"InvitedToSquadron": reflect.TypeFor[InvitedToSquadron](),
	"Shutdown":          reflect.TypeFor[Shutdown](),

	// used only for statistics and information
	"BuyAmmo":                       reflect.TypeFor[BuyAmmo](),
	"BuyDrones":                     reflect.TypeFor[BuyDrones](),
	"BuyExplorationData":            reflect.TypeFor[BuyExplorationData](),
	"BuySuit":                       reflect.TypeFor[BuySuit](),
	"BuyTradeData":                  reflect.TypeFor[BuyTradeData](),
	"BuyWeapon":                     reflect.TypeFor[BuyWeapon](),
	"ColonisationConstructionDepot": reflect.TypeFor[ColonisationConstructionDepot](),
	"CommunityGoalReward":           reflect.TypeFor[CommunityGoalReward](),
	"Died":                          reflect.TypeFor[Died](),
	"Docked":                        reflect.TypeFor[Docked](),
	"FSDJump":                       reflect.TypeFor[FSDJump](),
	"FSSSignalDiscovered":           reflect.TypeFor[FSSSignalDiscovered](),
	"FetchRemoteModule":             reflect.TypeFor[FetchRemoteModule](),
	"Fileheader":                    reflect.TypeFor[Fileheader](),
	"Location":                      reflect.TypeFor[Location](),
	"MarketBuy":                     reflect.TypeFor[MarketBuy](),
	"MarketSell":                    reflect.TypeFor[MarketSell](),
	"MissionCompleted":              reflect.TypeFor[MissionCompleted](),
	"ModuleBuy":                     reflect.TypeFor[ModuleBuy](),
	"ModuleBuyAndStore":             reflect.TypeFor[ModuleBuyAndStore](),
	"ModuleSell":                    reflect.TypeFor[ModuleSell](),
	"ModuleSellRemote":              reflect.TypeFor[ModuleSellRemote](),
	"MultiSellExplorationData":      reflect.TypeFor[MultiSellExplorationData](),
	"Music":                         reflect.TypeFor[Music](),
	"NpcCrewPaidWage":               reflect.TypeFor[NpcCrewPaidWage](),
	"PVPKill":                       reflect.TypeFor[PVPKill](),
	"PayBounties":                   reflect.TypeFor[PayBounties](),
	"PayFines":                      reflect.TypeFor[PayFines](),
	"RedeemVoucher":                 reflect.TypeFor[RedeemVoucher](),
	"RefuelAll":                     reflect.TypeFor[RefuelAll](),
	"RefuelPartial":                 reflect.TypeFor[RefuelPartial](),
	"Repair":                        reflect.TypeFor[Repair](),
	"RepairAll":                     reflect.TypeFor[RepairAll](),
	"Resupply":                      reflect.TypeFor[Resupply](),
	"Resurrect":                     reflect.TypeFor[Resurrect](),
	"SearchAndRescue":               reflect.TypeFor[SearchAndRescue](),
	"SellDrones":                    reflect.TypeFor[SellDrones](),
	"SellExplorationData":           reflect.TypeFor[SellExplorationData](),
	"SellMicroResources":            reflect.TypeFor[SellMicroResources](),
	"SellSuit":                      reflect.TypeFor[SellSuit](),
	"SellWeapon":                    reflect.TypeFor[SellWeapon](),
	"ShipyardBuy":                   reflect.TypeFor[ShipyardBuy](),
	"ShipyardRedeem":                reflect.TypeFor[ShipyardRedeem](),
	"ShipyardSell":                  reflect.TypeFor[ShipyardSell](),
	"ShipyardTransfer":              reflect.TypeFor[ShipyardTransfer](),
	"Missions":                      reflect.TypeFor[Missions](),

	// rest
	"Bounty":                   reflect.TypeFor[Bounty](),
	"CargoTransfer":            reflect.TypeFor[CargoTransfer](),
	"CarrierBankTransfer":      reflect.TypeFor[CarrierBankTransfer](),
	"ColonisationContribution": reflect.TypeFor[ColonisationContribution](),
	"FSDTarget":                reflect.TypeFor[FSDTarget](),
	"FactionKillBond":          reflect.TypeFor[FactionKillBond](),
	"MissionFailed":            reflect.TypeFor[MissionFailed](),
	"Screenshot":               reflect.TypeFor[Screenshot](),
	"ShipLocker":               reflect.TypeFor[ShipLocker](),
	"ShipRedeemed":             reflect.TypeFor[ShipRedeemed](),
	"SquadronStartup":          reflect.TypeFor[SquadronStartup](),
	"StartJump":                reflect.TypeFor[StartJump](),
	"Statistics":               reflect.TypeFor[Statistics](),
	"Loadout":                  reflect.TypeFor[Loadout](),
	"MissionAbandoned":         reflect.TypeFor[MissionAbandoned](),
	"MissionAccepted":          reflect.TypeFor[MissionAccepted](),
	"ShipTargeted":             reflect.TypeFor[ShipTargeted](),
	"Cargo":                    reflect.TypeFor[Cargo](),
	"StoredModules":            reflect.TypeFor[StoredModules](),
	"TechnologyBroker":         reflect.TypeFor[TechnologyBroker](),
	"Undocked":                 reflect.TypeFor[Undocked](),
}

type JournalEvent interface {
	IsEliteEvent() bool
	EventType() string
}

type BaseEvent struct {
	Timestamp time.Time `json:"timestamp,required,format:'2006-01-02T15:04:05Z'"`
	Event     string    `json:"event,required,unique"`
}

func (e *BaseEvent) IsEliteEvent() bool {
	return true
}

func (e *BaseEvent) EventType() string {
	return e.Event
}

type Continued struct {
	BaseEvent `json:",inline"`

	Part int `json:"Part,required"`
}

type ReceiveText struct {
	BaseEvent `json:",inline"`

	From             string `json:"From,required,unique"`
	Message          string `json:"Message,required"`
	MessageLocalised string `json:"Message_Localised,omitempty"`     // available only for messages from NPCs (Channel: "npc,required")
	FromLocalised    string `json:"From_Localised,unique,omitempty"` // available only for messages from NPCs (Channel: "npc,required")
	Channel          string `json:"Channel,required,unique"`
}

type SendText struct {
	BaseEvent `json:",inline"`

	To      string `json:"To,required,unique"`
	Message string `json:"Message,required"`
	Sent    bool   `json:"Sent,required"`
}

type Friends struct {
	BaseEvent `json:",inline"`

	Status string `json:"Status,required,unique"` // Requested, Declined, Added, Lost, Offline, Online
	Name   string `json:"Name,required,unique"`
}

type Commander struct {
	BaseEvent `json:",inline"`

	FID  string `json:"FID,required,unique"`
	Name string `json:"Name,required,unique"`
}

type LoadGame struct {
	BaseEvent `json:",inline"`

	FID           string   `json:"FID,unique,required"`
	Commander     string   `json:"Commander,unique,required"`
	Horizons      bool     `json:"Horizons,required"`
	Odyssey       bool     `json:"Odyssey,required"`
	Ship          string   `json:"Ship,unique,omitempty"`           // may be not present
	ShipLocalised string   `json:"Ship_Localised,unique,omitempty"` // may be not present
	ShipID        *int     `json:"ShipID,omitempty"`                // may be not present
	ShipName      string   `json:"ShipName,unique,omitempty"`       // may be not present
	ShipIdent     string   `json:"ShipIdent,unique,omitempty"`      // may be not present
	FuelLevel     *float64 `json:"FuelLevel,omitempty"`             // may be not present
	FuelCapacity  *float64 `json:"FuelCapacity,omitempty"`          // may be not present
	GameMode      string   `json:"GameMode,unique,omitempty"`       // may be not present TODO: when?
	Group         string   `json:"Group,unique,omitempty"`          // may be not present
	Credits       int      `json:"Credits,required"`
	Loan          int      `json:"Loan,required"`
	Language      string   `json:"language,unique,required"`
	Gameversion   string   `json:"gameversion,unique,required"`
	Build         string   `json:"build,unique,required"`
}

type Fileheader struct {
	BaseEvent `json:",inline"`

	Part        int    `json:"part,required"`
	Language    string `json:"language,unique,required"`
	Odyssey     bool   `json:"Odyssey,required"`
	Gameversion string `json:"gameversion,unique,required"`
	Build       string `json:"build,unique,required"`
}

type FSSSignalDiscovered struct {
	BaseEvent `json:",inline"`

	SystemAddress int64  `json:"SystemAddress,required"`
	SignalName    string `json:"SignalName,required,unique"`
	SignalType    string `json:"SignalType,required,unique"`

	ThreatLevel              *int     `json:"ThreatLevel,omitempty"`                      // may be not present
	TimeRemaining            *float64 `json:"TimeRemaining,omitempty"`                    // may be not present
	IsStation                *bool    `json:"IsStation,omitempty"`                        // may be not present
	SignalNameLocalised      string   `json:"SignalName_Localised,omitempty,unique"`      // may be not present
	USSType                  string   `json:"USSType,omitempty,unique"`                   // may be not present
	USSTypeLocalised         string   `json:"USSType_Localised,omitempty,unique"`         // may be not present
	SpawningState            string   `json:"SpawningState,omitempty,unique"`             // may be not present
	SpawningStateLocalised   string   `json:"SpawningState_Localised,omitempty,unique"`   // may be not present
	SpawningFaction          string   `json:"SpawningFaction,omitempty,unique"`           // may be not present
	SpawningFactionLocalised string   `json:"SpawningFaction_Localised,omitempty,unique"` // may be not present
}

type Music struct {
	BaseEvent `json:",inline"`

	MusicTrack string `json:"MusicTrack,unique,required"` // "NoTrack" when silence
}

type Shutdown struct {
	BaseEvent `json:",inline"`
}

type FSDJump struct {
	BaseEvent `json:",inline"`

	Taxi                          bool      `json:"Taxi,required"`
	Multicrew                     bool      `json:"Multicrew,required"`
	StarSystem                    string    `json:"StarSystem,required,unique"`
	SystemAddress                 int64     `json:"SystemAddress,required"`
	StarPos                       []float64 `json:"StarPos,required"`
	SystemAllegiance              string    `json:"SystemAllegiance,required,unique"`
	SystemEconomy                 string    `json:"SystemEconomy,required,unique"`
	SystemEconomyLocalised        string    `json:"SystemEconomy_Localised,required,unique"`
	SystemSecondEconomy           string    `json:"SystemSecondEconomy,required,unique"`
	SystemSecondEconomyLocalised  string    `json:"SystemSecondEconomy_Localised,required,unique"`
	SystemGovernment              string    `json:"SystemGovernment,required,unique"`
	SystemGovernmentLocalised     string    `json:"SystemGovernment_Localised,required,unique"`
	SystemSecurity                string    `json:"SystemSecurity,required,unique"`
	SystemSecurityLocalised       string    `json:"SystemSecurity_Localised,required,unique"`
	Population                    int       `json:"Population,required"`
	Body                          string    `json:"Body,required,unique"`
	BodyID                        int       `json:"BodyID,required"`
	BodyType                      string    `json:"BodyType,required,unique"`
	ControllingPower              string    `json:"ControllingPower,omitempty,unique"`       // may be not present
	Powers                        []string  `json:"Powers,omitempty"`                        // may be not present
	PowerplayState                string    `json:"PowerplayState,omitempty,unique"`         // may be not present
	PowerplayStateControlProgress *float64  `json:"PowerplayStateControlProgress,omitempty"` // may be not present
	PowerplayStateReinforcement   *int      `json:"PowerplayStateReinforcement,omitempty"`   // may be not present
	PowerplayStateUndermining     *int      `json:"PowerplayStateUndermining,omitempty"`     // may be not present
	JumpDist                      float64   `json:"JumpDist,required"`
	FuelUsed                      float64   `json:"FuelUsed,required"`
	FuelLevel                     float64   `json:"FuelLevel,required"`

	// TODO:
	//Factions                      []struct {
	//	Name               string  `json:"Name"`
	//	FactionState       string  `json:"FactionState"`
	//	Government         string  `json:"Government"`
	//	Influence          float64 `json:"Influence"`
	//	Allegiance         string  `json:"Allegiance"`
	//	Happiness          string  `json:"Happiness"`
	//	HappinessLocalised string  `json:"Happiness_Localised"`
	//	MyReputation       float64 `json:"MyReputation"`
	//	PendingStates      []struct {
	//		State string `json:"State"`
	//		Trend int    `json:"Trend"`
	//	} `json:"PendingStates,omitempty"`
	//	ActiveStates []struct {
	//		State string `json:"State"`
	//	} `json:"ActiveStates,omitempty"`
	//	RecoveringStates []struct {
	//		State string `json:"State"`
	//		Trend int    `json:"Trend"`
	//	} `json:"RecoveringStates,omitempty"`
	//} `json:"Factions"`
	//SystemFaction struct {
	//	Name string `json:"Name"`
	//} `json:"SystemFaction"`
	//Conflicts []struct {
	//	WarType  string `json:"WarType"`
	//	Status   string `json:"Status"`
	//	Faction1 struct {
	//		Name    string `json:"Name"`
	//		Stake   string `json:"Stake"`
	//		WonDays int    `json:"WonDays"`
	//	} `json:"Faction1"`
	//	Faction2 struct {
	//		Name    string `json:"Name"`
	//		Stake   string `json:"Stake"`
	//		WonDays int    `json:"WonDays"`
	//	} `json:"Faction2"`
	//} `json:"Conflicts"`
}

type NpcCrewPaidWage struct {
	BaseEvent `json:",inline"`

	NpcCrewName string `json:"NpcCrewName,unique,required"`
	NpcCrewId   int    `json:"NpcCrewId,required"`
	Amount      int    `json:"Amount,required"`
}

type StartJump struct {
	BaseEvent `json:",inline"`

	JumpType      string `json:"JumpType,unique,required"`
	Taxi          bool   `json:"Taxi,required"`
	StarSystem    string `json:"StarSystem,unique,omitempty"` // may be not present (example: "JumpType": "Supercruise")
	SystemAddress *int64 `json:"SystemAddress,omitempty"`     // may be not present (example: "JumpType": "Supercruise")
	StarClass     string `json:"StarClass,unique,omitempty"`  // may be not present (example: "JumpType": "Supercruise")
}

type SquadronStartup struct {
	BaseEvent `json:",inline"`

	SquadronID      *int   `json:"SquadronID,omitempty"` // TODO: not present when not in squadron?
	SquadronName    string `json:"SquadronName,required,unique"`
	CurrentRank     int    `json:"CurrentRank,required"`
	CurrentRankName string `json:"CurrentRankName,unique,omitempty"` // TODO: not present when not in squadron?
}

type Died struct {
	BaseEvent `json:",inline"`
	// when killed by npc, it can be just a name directly, or in a format:
	// "$LUASC_Scenario_Warzone_NPC_SpecOps_D;" - in this case, "Localised" versions should be used.
	// When killed by a player, there is always "Cmdr " prefix
	KillerName          string `json:"KillerName,omitempty,unique"`           // may be not present
	KillerShip          string `json:"KillerShip,omitempty,unique"`           // may be not present
	KillerRank          string `json:"KillerRank,omitempty,unique"`           // may be not present
	KillerNameLocalised string `json:"KillerName_Localised,omitempty,unique"` // may be not present
}

type FSDTarget struct {
	BaseEvent `json:",inline"`

	Name                  string `json:"Name,required,unique"`
	SystemAddress         int64  `json:"SystemAddress,required"`
	StarClass             string `json:"StarClass,required,unique"`
	RemainingJumpsInRoute *int   `json:"RemainingJumpsInRoute,omitempty"` // may be not present
}

type Docked struct {
	BaseEvent `json:",inline"`

	StationName    string `json:"StationName,required,unique"`
	StationType    string `json:"StationType,required,unique"`
	Taxi           bool   `json:"Taxi,required"`
	Multicrew      bool   `json:"Multicrew,required"`
	StarSystem     string `json:"StarSystem,required,unique"`
	SystemAddress  int64  `json:"SystemAddress,required"`
	MarketID       int64  `json:"MarketID,required"`
	StationFaction struct {
		Name         string `json:"Name,required,unique"`
		FactionState string `json:"FactionState,omitempty,unique"` // may be not present
	} `json:"StationFaction,required"`
	StationGovernment          string   `json:"StationGovernment,required,unique"`
	StationGovernmentLocalised string   `json:"StationGovernment_Localised,required,unique"`
	StationAllegiance          string   `json:"StationAllegiance,omitempty,unique"` // may be not present
	StationServices            []string `json:"StationServices,required"`
	StationEconomy             string   `json:"StationEconomy,required,unique"`
	StationEconomyLocalised    string   `json:"StationEconomy_Localised,required,unique"`
	StationEconomies           []struct {
		Name          string  `json:"Name,required,unique"`
		NameLocalised string  `json:"Name_Localised,required,unique"`
		Proportion    float64 `json:"Proportion,required"`
	} `json:"StationEconomies,required"`
	DistFromStarLS float64 `json:"DistFromStarLS,required"`
	LandingPads    struct {
		Small  int `json:"Small,required"`
		Medium int `json:"Medium,required"`
		Large  int `json:"Large,required"`
	} `json:"LandingPads,required"`
	Wanted     *bool `json:"Wanted,omitempty"`     // may be not present
	ActiveFine *bool `json:"ActiveFine,omitempty"` // may be not present
}

type Undocked struct {
	BaseEvent `json:",inline"`

	StationName string `json:"StationName,required,unique"`
	StationType string `json:"StationType,required,unique"`
	MarketID    int64  `json:"MarketID,required"`
	Taxi        bool   `json:"Taxi,required"`
	Multicrew   bool   `json:"Multicrew,required"`
}

type MarketBuy struct {
	BaseEvent `json:",inline"`

	MarketID  int64  `json:"MarketID,required"`
	Type      string `json:"Type,required,unique"`
	Count     int    `json:"Count,required"`
	BuyPrice  int    `json:"BuyPrice,required"`
	TotalCost int    `json:"TotalCost,required"`
}

type MarketSell struct {
	BaseEvent `json:",inline"`

	MarketID      int64  `json:"MarketID,required"`
	Type          string `json:"Type,required,unique"`
	TypeLocalised string `json:"Type_Localised,omitempty,unique"` // may be not present
	Count         int    `json:"Count,required"`
	SellPrice     int    `json:"SellPrice,required"`
	TotalSale     int    `json:"TotalSale,required"`
	AvgPricePaid  int    `json:"AvgPricePaid,required"`
	StolenGoods   bool   `json:"StolenGoods,omitempty"` // may be not present
	BlackMarket   bool   `json:"BlackMarket,omitempty"` // may be not present
}

type Repair struct {
	BaseEvent `json:",inline"`

	Items []string `json:"Items,required"`
	Cost  int      `json:"Cost,required"`
}

type RepairAll struct {
	BaseEvent `json:",inline"`

	Cost int `json:"Cost,required"`
}

type RefuelAll struct {
	BaseEvent `json:",inline"`

	Cost   int     `json:"Cost,required"`
	Amount float64 `json:"Amount,required"`
}

type RefuelPartial struct {
	BaseEvent `json:",inline"`

	Cost   int     `json:"Cost,required"`
	Amount float64 `json:"Amount,required"`
}

type MissionCompleted struct {
	BaseEvent `json:",inline"`

	Faction       string `json:"Faction,required"`
	Name          string `json:"Name,required"`
	LocalisedName string `json:"LocalisedName,required"`
	MissionID     uint64 `json:"MissionID,required"`

	//Commodity          string `json:"Commodity,omitempty"`
	//CommodityLocalised string `json:"Commodity_Localised,omitempty"`
	//Count              int    `json:"Count,omitempty"`
	Reward *int `json:"Reward,omitempty"` // may be not present (only when reward type is credits)
	//MaterialsReward    []struct {
	//	Name              string `json:"Name"`
	//	NameLocalised     string `json:"Name_Localised,omitempty"`
	//	Category          string `json:"Category"`
	//	CategoryLocalised string `json:"Category_Localised"`
	//	Count             int    `json:"Count"`
	//} `json:"MaterialsReward,omitempty"`
	//FactionEffects []struct {
	//	Faction string `json:"Faction"`
	//	Effects []struct {
	//		Effect          string `json:"Effect"`
	//		EffectLocalised string `json:"Effect_Localised"`
	//		Trend           string `json:"Trend"`
	//	} `json:"Effects"`
	//	Influence []struct {
	//		SystemAddress int64  `json:"SystemAddress"`
	//		Trend         string `json:"Trend"`
	//		Influence     string `json:"Influence"`
	//	} `json:"Influence"`
	//	ReputationTrend string `json:"ReputationTrend"`
	//	Reputation      string `json:"Reputation"`
	//} `json:"FactionEffects"`
	//Donation              string `json:"Donation,omitempty"`
	//Donated               int    `json:"Donated,omitempty"`
	//TargetType            string `json:"TargetType,omitempty"`
	//TargetTypeLocalised   string `json:"TargetType_Localised,omitempty"`
	//TargetFaction         string `json:"TargetFaction,omitempty"`
	//KillCount             int    `json:"KillCount,omitempty"`
	//DestinationSystem     string `json:"DestinationSystem,omitempty"`
	//DestinationSettlement string `json:"DestinationSettlement,omitempty"`
	//Target                string `json:"Target,omitempty"`
	//TargetLocalised       string `json:"Target_Localised,omitempty"`
}

type MissionFailed struct {
	BaseEvent `json:",inline"`

	Name          string `json:"Name,required"`
	LocalisedName string `json:"LocalisedName,required"`
	MissionID     uint64 `json:"MissionID,required"`
}

type ModuleBuy struct {
	BaseEvent `json:",inline"`

	Slot string `json:"Slot,required"`
	//SellItem            string `json:"SellItem,omitempty"`
	//SellItemLocalised   string `json:"SellItem_Localised,omitempty"`
	//SellPrice           int    `json:"SellPrice,omitempty"`
	BuyItem          string `json:"BuyItem,required,unique"`
	BuyItemLocalised string `json:"BuyItem_Localised,required,unique"`
	MarketID         int64  `json:"MarketID,required"`
	BuyPrice         int    `json:"BuyPrice,required"`
	Ship             string `json:"Ship,required,unique"`
	ShipID           int    `json:"ShipID,required"`
	//StoredItem          string `json:"StoredItem,omitempty"`
	//StoredItemLocalised string `json:"StoredItem_Localised,omitempty"`
}

type ModuleSell struct {
	BaseEvent `json:",inline"`

	MarketID          int64  `json:"MarketID,required"`
	Slot              string `json:"Slot,required,unique"`
	SellItem          string `json:"SellItem,required,unique"`
	SellItemLocalised string `json:"SellItem_Localised,required,unique"`
	SellPrice         int    `json:"SellPrice,required"`
	Ship              string `json:"Ship,required"`
	ShipID            int    `json:"ShipID,required"`
}

type ModuleSellRemote struct {
	BaseEvent `json:",inline"`

	StorageSlot       int    `json:"StorageSlot,required"`
	SellItem          string `json:"SellItem,required,unique"`
	SellItemLocalised string `json:"SellItem_Localised,required,unique"`
	ServerId          int    `json:"ServerId,required"`
	SellPrice         int    `json:"SellPrice,required"`
	Ship              string `json:"Ship,required,unique"`
	ShipID            int    `json:"ShipID,required"`
}

type PayBounties struct {
	BaseEvent `json:",inline"`

	Amount           int     `json:"Amount,required"`
	AllFines         bool    `json:"AllFines,required"`
	ShipID           int64   `json:"ShipID,required"`
	BrokerPercentage float64 `json:"BrokerPercentage,required"`
	Faction          string  `json:"Faction,omitempty,unique"` // may be not present
}

type PayFines struct {
	BaseEvent `json:",inline"`

	Amount           int      `json:"Amount,required"`
	AllFines         bool     `json:"AllFines,required"`
	ShipID           int      `json:"ShipID,required"`
	BrokerPercentage *float64 `json:"BrokerPercentage,omitempty"` // may be not present
	Faction          string   `json:"Faction,omitempty,unique"`   // may be not present
}

type BuyAmmo struct {
	BaseEvent `json:",inline"`

	Cost int `json:"Cost,required"`
}

type BuyDrones struct {
	BaseEvent `json:",inline"`

	Type      string `json:"Type,required,unique"`
	Count     int    `json:"Count,required"`
	BuyPrice  int    `json:"BuyPrice,required"`
	TotalCost int    `json:"TotalCost,required"`
}

type ShipyardSell struct {
	BaseEvent `json:",inline"`

	ShipType          string `json:"ShipType,required,unique"`
	ShipTypeLocalised string `json:"ShipType_Localised,required,unique"`
	SellShipID        int    `json:"SellShipID,required"`
	ShipPrice         int    `json:"ShipPrice,required"`
	MarketID          int64  `json:"MarketID,required"`
}

type SellDrones struct {
	BaseEvent `json:",inline"`

	Type      string `json:"Type,required,unique"`
	Count     int    `json:"Count,required"`
	SellPrice int    `json:"SellPrice,required"`
	TotalSale int    `json:"TotalSale,required"`
}

type RedeemVoucher struct {
	BaseEvent `json:",inline"`

	Type   string `json:"Type,required,unique"`
	Amount int    `json:"Amount,required"`
	//Factions []struct {
	//	Faction string `json:"Faction,required"`
	//	Amount  int    `json:"Amount,required"`
	//} `json:"Factions,omitempty"`
	BrokerPercentage *float64 `json:"BrokerPercentage,omitempty"` // may be not present
	Faction          string   `json:"Faction,omitempty,unique"`   // may be not present
}

type SearchAndRescue struct {
	BaseEvent `json:",inline"`

	MarketID      int64  `json:"MarketID,required"`
	Name          string `json:"Name,required,unique"`
	NameLocalised string `json:"Name_Localised,required,unique"`
	Count         int    `json:"Count,required"`
	Reward        int    `json:"Reward,required"`
}

type CommunityGoalReward struct {
	BaseEvent `json:",inline"`

	CGID   int    `json:"CGID,required"`
	Name   string `json:"Name,required,unique"`
	System string `json:"System,required,unique"`
	Reward int    `json:"Reward,required"`
}

type Resurrect struct {
	BaseEvent `json:",inline"`

	Option   string `json:"Option,required,unique"`
	Cost     int    `json:"Cost,required"`
	Bankrupt bool   `json:"Bankrupt,required"`
}

type FetchRemoteModule struct {
	BaseEvent `json:",inline"`

	StorageSlot         int    `json:"StorageSlot,required"`
	StoredItem          string `json:"StoredItem,required,unique"`
	StoredItemLocalised string `json:"StoredItem_Localised,required,unique"`
	ServerId            int    `json:"ServerId,required"`
	TransferCost        int    `json:"TransferCost,required"`
	TransferTime        int    `json:"TransferTime,required"`
	Ship                string `json:"Ship,required,unique"`
	ShipID              int    `json:"ShipID,required"`
}

type SellExplorationData struct {
	BaseEvent `json:",inline"`

	Systems []string `json:"Systems,required"`
	//Discovered    []interface{} `json:"Discovered,required"` // ?
	BaseValue     int `json:"BaseValue,required"`
	Bonus         int `json:"Bonus,required"`
	TotalEarnings int `json:"TotalEarnings,required"`
}

type SellWeapon struct {
	BaseEvent `json:",inline"`

	Name          string   `json:"Name,required,unique"`
	NameLocalised string   `json:"Name_Localised,required,unique"`
	Class         int      `json:"Class,required"`
	WeaponMods    []string `json:"WeaponMods,required"`
	Price         int      `json:"Price,required"`
	SuitModuleID  int64    `json:"SuitModuleID,required"`
}

type BuyWeapon struct {
	BaseEvent `json:",inline"`

	Name          string   `json:"Name,required,unique"`
	NameLocalised string   `json:"Name_Localised,required,unique"`
	Class         int      `json:"Class,required"`
	Price         int      `json:"Price,required"`
	SuitModuleID  int64    `json:"SuitModuleID,required"`
	WeaponMods    []string `json:"WeaponMods,required"`
}

type ModuleBuyAndStore struct {
	BaseEvent `json:",inline"`

	BuyItem          string `json:"BuyItem,required,unique"`
	BuyItemLocalised string `json:"BuyItem_Localised,required,unique"`
	MarketID         int64  `json:"MarketID,required"`
	BuyPrice         int    `json:"BuyPrice,required"`
	Ship             string `json:"Ship,required,unique"`
	ShipID           int    `json:"ShipID,required"`
}

type TechnologyBroker struct {
	BaseEvent `json:",inline"`

	BrokerType    string `json:"BrokerType,required"`
	MarketID      int64  `json:"MarketID,required"`
	ItemsUnlocked []struct {
		Name          string `json:"Name,required,unique"`
		NameLocalised string `json:"Name_Localised,required,unique"`
	} `json:"ItemsUnlocked,required"`
	Commodities []struct {
		Name          string `json:"Name,required,unique"`
		NameLocalised string `json:"Name_Localised,required,unique"`
		Count         int    `json:"Count,required"`
	} `json:"Commodities,required"`
	Materials []struct {
		Name          string `json:"Name,required,unique"`
		Count         int    `json:"Count,required"`
		Category      string `json:"Category,required,unique"`
		NameLocalised string `json:"Name_Localised,omitempty,unique"` // may be not present
	} `json:"Materials,required"`
}

type BuyTradeData struct {
	BaseEvent `json:",inline"`

	System string `json:"System,required,unique"`
	Cost   int    `json:"Cost,required"`
}

type Resupply struct {
	BaseEvent `json:",inline"`
}

type SellSuit struct {
	BaseEvent `json:",inline"`

	SuitID        int64    `json:"SuitID,required"`
	SuitMods      []string `json:"SuitMods,required"`
	Name          string   `json:"Name,required,unique"`
	NameLocalised string   `json:"Name_Localised,required,unique"`
	Price         int      `json:"Price,required"`
}

type ShipyardRedeem struct {
	BaseEvent `json:",inline"`

	ShipType          string `json:"ShipType,required,unique"`
	ShipTypeLocalised string `json:"ShipType_Localised,omitempty,unique"` // may be not present
	BundleID          int    `json:"BundleID,required"`
	MarketID          int64  `json:"MarketID,required"`
}

type ShipRedeemed struct {
	BaseEvent `json:",inline"`

	ShipType          string `json:"ShipType,required,unique"`
	ShipTypeLocalised string `json:"ShipType_Localised,omitempty,unique"` // may be not present
	NewShipID         int    `json:"NewShipID,required"`
}

type BuySuit struct {
	BaseEvent `json:",inline"`

	Name          string   `json:"Name,required,unique"`
	NameLocalised string   `json:"Name_Localised,required,unique"`
	Price         int      `json:"Price,required"`
	SuitID        int64    `json:"SuitID,required"`
	SuitMods      []string `json:"SuitMods,required"`
}

type ShipyardBuy struct {
	BaseEvent `json:",inline"`

	ShipType          string `json:"ShipType,required,unique"`
	ShipTypeLocalised string `json:"ShipType_Localised,omitempty,unique"` // may be not present
	ShipPrice         int    `json:"ShipPrice,required"`
	StoreOldShip      string `json:"StoreOldShip,required,unique"`
	StoreShipID       int    `json:"StoreShipID,required"`
	MarketID          int64  `json:"MarketID,required"`
}

type Screenshot struct {
	BaseEvent `json:",inline"`

	Filename  string   `json:"Filename,required"`
	Width     int      `json:"Width,required"`
	Height    int      `json:"Height,required"`
	System    *string  `json:"System,omitempty"`    // may be not present
	Body      *string  `json:"Body,omitempty"`      // may be not present
	Latitude  *float64 `json:"Latitude,omitempty"`  // may be not present
	Longitude *float64 `json:"Longitude,omitempty"` // may be not present
	Heading   *int     `json:"Heading,omitempty"`   // may be not present
}

type CargoTransfer struct {
	BaseEvent `json:",inline"`

	Transfers []struct {
		Type          string `json:"Type,required,unique"`
		TypeLocalised string `json:"Type_Localised,omitempty,unique"` // may be not present
		Count         int    `json:"Count,required"`
		Direction     string `json:"Direction,required,unique"`
	} `json:"Transfers,required"`
}

type ShipLocker struct {
	BaseEvent `json:",inline"`

	Items []struct {
		Name          string  `json:"Name,required,unique"`
		OwnerID       int     `json:"OwnerID,required"`
		MissionID     *uint64 `json:"MissionID,omitempty"` // may be not present
		Count         int     `json:"Count,required"`
		NameLocalised string  `json:"Name_Localised,omitempty,unique"` // may be not present
	} `json:"Items,omitempty"` // may be not present
	Components []struct {
		Name          string `json:"Name,required,unique"`
		OwnerID       int    `json:"OwnerID,required"`
		Count         int    `json:"Count,required"`
		NameLocalised string `json:"Name_Localised,omitempty,unique"` // may be not present
	} `json:"Components,omitempty"` // may be not present
	Consumables []struct {
		Name          string `json:"Name,required,unique"`
		NameLocalised string `json:"Name_Localised,required,unique"`
		OwnerID       int    `json:"OwnerID,required"`
		Count         int    `json:"Count,required"`
	} `json:"Consumables,omitempty"` // may be not present
	Data []struct {
		Name          string  `json:"Name,required,unique"`
		OwnerID       int     `json:"OwnerID,required"`
		MissionID     *uint64 `json:"MissionID,omitempty"` // may be not present
		Count         int     `json:"Count,required"`
		NameLocalised string  `json:"Name_Localised,omitempty,unique"` // may be not present
	} `json:"Data,omitempty"` // may be not present
}

type Statistics struct {
	BaseEvent `json:",inline"`

	BankAccount struct {
		CurrentWealth          int64 `json:"Current_Wealth,required"`
		SpentOnShips           int   `json:"Spent_On_Ships,required"`
		SpentOnOutfitting      int64 `json:"Spent_On_Outfitting,required"`
		SpentOnRepairs         int   `json:"Spent_On_Repairs,required"`
		SpentOnFuel            int   `json:"Spent_On_Fuel,required"`
		SpentOnAmmoConsumables int   `json:"Spent_On_Ammo_Consumables,required"`
		InsuranceClaims        int   `json:"Insurance_Claims,required"`
		SpentOnInsurance       int   `json:"Spent_On_Insurance,required"`
		OwnedShipCount         int   `json:"Owned_Ship_Count,required"`
		SpentOnSuits           int   `json:"Spent_On_Suits,omitempty"`
		SpentOnWeapons         int   `json:"Spent_On_Weapons,omitempty"`
		SpentOnSuitConsumables int   `json:"Spent_On_Suit_Consumables,omitempty"`
		SuitsOwned             int   `json:"Suits_Owned,required"`
		WeaponsOwned           int   `json:"Weapons_Owned,required"`
		SpentOnPremiumStock    int   `json:"Spent_On_Premium_Stock,required"`
		PremiumStockBought     int   `json:"Premium_Stock_Bought,required"`
	} `json:"Bank_Account,required"`
	Combat struct {
		BountiesClaimed          int   `json:"Bounties_Claimed,required"`
		BountyHuntingProfit      int   `json:"Bounty_Hunting_Profit,required"`
		CombatBonds              int   `json:"Combat_Bonds,required"`
		CombatBondProfits        int64 `json:"Combat_Bond_Profits,required"`
		Assassinations           int   `json:"Assassinations,required"`
		AssassinationProfits     int   `json:"Assassination_Profits,required"`
		HighestSingleReward      int   `json:"Highest_Single_Reward,required"`
		SkimmersKilled           int   `json:"Skimmers_Killed,required"`
		OnFootCombatBonds        int   `json:"OnFoot_Combat_Bonds,required"`
		OnFootCombatBondsProfits int   `json:"OnFoot_Combat_Bonds_Profits,required"`
		OnFootVehiclesDestroyed  int   `json:"OnFoot_Vehicles_Destroyed,required"`
		OnFootShipsDestroyed     int   `json:"OnFoot_Ships_Destroyed,required"`
		DropshipsTaken           int   `json:"Dropships_Taken,required"`
		DropshipsBooked          int   `json:"Dropships_Booked,required"`
		DropshipsCancelled       int   `json:"Dropships_Cancelled,required"`
		ConflictZoneHigh         int   `json:"ConflictZone_High,required"`
		ConflictZoneMedium       int   `json:"ConflictZone_Medium,required"`
		ConflictZoneLow          int   `json:"ConflictZone_Low,required"`
		ConflictZoneTotal        int   `json:"ConflictZone_Total,required"`
		ConflictZoneHighWins     int   `json:"ConflictZone_High_Wins,required"`
		ConflictZoneMediumWins   int   `json:"ConflictZone_Medium_Wins,required"`
		ConflictZoneLowWins      int   `json:"ConflictZone_Low_Wins,required"`
		ConflictZoneTotalWins    int   `json:"ConflictZone_Total_Wins,required"`
		SettlementDefended       int   `json:"Settlement_Defended,required"`
		SettlementConquered      int   `json:"Settlement_Conquered,required"`
		OnFootSkimmersKilled     int   `json:"OnFoot_Skimmers_Killed,required"`
		OnFootScavsKilled        int   `json:"OnFoot_Scavs_Killed,required"`
	} `json:"Combat,required"`
	Crime struct {
		Notoriety                int `json:"Notoriety,required"`
		Fines                    int `json:"Fines,required"`
		TotalFines               int `json:"Total_Fines,required"`
		BountiesReceived         int `json:"Bounties_Received,required"`
		TotalBounties            int `json:"Total_Bounties,required"`
		HighestBounty            int `json:"Highest_Bounty,required"`
		MalwareUploaded          int `json:"Malware_Uploaded,required"`
		SettlementsStateShutdown int `json:"Settlements_State_Shutdown,required"`
		ProductionSabotage       int `json:"Production_Sabotage,required"`
		ProductionTheft          int `json:"Production_Theft,required"`
		TotalMurders             int `json:"Total_Murders,required"`
		CitizensMurdered         int `json:"Citizens_Murdered,required"`
		OmnipolMurdered          int `json:"Omnipol_Murdered,required"`
		GuardsMurdered           int `json:"Guards_Murdered,required"`
		DataStolen               int `json:"Data_Stolen,required"`
		GoodsStolen              int `json:"Goods_Stolen,required"`
		SampleStolen             int `json:"Sample_Stolen,required"`
		TotalStolen              int `json:"Total_Stolen,required"`
		TurretsDestroyed         int `json:"Turrets_Destroyed,required"`
		TurretsOverloaded        int `json:"Turrets_Overloaded,required"`
		TurretsTotal             int `json:"Turrets_Total,required"`
		ValueStolenStateChange   int `json:"Value_Stolen_StateChange,required"`
		ProfilesCloned           int `json:"Profiles_Cloned,required"`
	} `json:"Crime,required"`
	Smuggling struct {
		BlackMarketsTradedWith   int     `json:"Black_Markets_Traded_With,required"`
		BlackMarketsProfits      int     `json:"Black_Markets_Profits,required"`
		ResourcesSmuggled        int     `json:"Resources_Smuggled,required"`
		AverageProfit            float64 `json:"Average_Profit,required"`
		HighestSingleTransaction int     `json:"Highest_Single_Transaction,required"`
	} `json:"Smuggling,required"`
	Trading struct {
		MarketsTradedWith        int     `json:"Markets_Traded_With,required"`
		MarketProfits            int64   `json:"Market_Profits,required"`
		ResourcesTraded          int     `json:"Resources_Traded,required"`
		AverageProfit            float64 `json:"Average_Profit,required"`
		HighestSingleTransaction int     `json:"Highest_Single_Transaction,required"`
		DataSold                 int     `json:"Data_Sold,required"`
		GoodsSold                int     `json:"Goods_Sold,required"`
		AssetsSold               int     `json:"Assets_Sold,required"`
	} `json:"Trading,required"`
	Mining struct {
		MiningProfits      int `json:"Mining_Profits,required"`
		QuantityMined      int `json:"Quantity_Mined,required"`
		MaterialsCollected int `json:"Materials_Collected,required"`
	} `json:"Mining,required"`
	Exploration struct {
		SystemsVisited            int     `json:"Systems_Visited,required"`
		ExplorationProfits        int     `json:"Exploration_Profits,required"`
		PlanetsScannedToLevel2    int     `json:"Planets_Scanned_To_Level_2,required"`
		PlanetsScannedToLevel3    int     `json:"Planets_Scanned_To_Level_3,required"`
		EfficientScans            int     `json:"Efficient_Scans,required"`
		HighestPayout             int     `json:"Highest_Payout,required"`
		TotalHyperspaceDistance   int     `json:"Total_Hyperspace_Distance,required"`
		TotalHyperspaceJumps      int     `json:"Total_Hyperspace_Jumps,required"`
		GreatestDistanceFromStart float64 `json:"Greatest_Distance_From_Start,required"`
		TimePlayed                int     `json:"Time_Played,required"`
		OnFootDistanceTravelled   int     `json:"OnFoot_Distance_Travelled,required"`
		ShuttleJourneys           int     `json:"Shuttle_Journeys,required"`
		ShuttleDistanceTravelled  float64 `json:"Shuttle_Distance_Travelled,required"`
		SpentOnShuttles           int     `json:"Spent_On_Shuttles,required"`
		FirstFootfalls            int     `json:"First_Footfalls,required"`
		PlanetFootfalls           int     `json:"Planet_Footfalls,required"`
		SettlementsVisited        int     `json:"Settlements_Visited,required"`
	} `json:"Exploration,required"`
	Passengers struct {
		PassengersMissionsBulk      int `json:"Passengers_Missions_Bulk,required"`
		PassengersMissionsVIP       int `json:"Passengers_Missions_VIP,required"`
		PassengersMissionsDelivered int `json:"Passengers_Missions_Delivered,required"`
		PassengersMissionsEjected   int `json:"Passengers_Missions_Ejected,required"`
	} `json:"Passengers,required"`
	SearchAndRescue struct {
		SearchRescueTraded        int `json:"SearchRescue_Traded,required"`
		SearchRescueProfit        int `json:"SearchRescue_Profit,required"`
		SearchRescueCount         int `json:"SearchRescue_Count,required"`
		SalvageLegalPOI           int `json:"Salvage_Legal_POI,required"`
		SalvageLegalSettlements   int `json:"Salvage_Legal_Settlements,required"`
		SalvageIllegalPOI         int `json:"Salvage_Illegal_POI,required"`
		SalvageIllegalSettlements int `json:"Salvage_Illegal_Settlements,required"`
		MaglocksOpened            int `json:"Maglocks_Opened,required"`
		PanelsOpened              int `json:"Panels_Opened,required"`
		SettlementsStateFireOut   int `json:"Settlements_State_FireOut,required"`
		SettlementsStateReboot    int `json:"Settlements_State_Reboot,required"`
	} `json:"Search_And_Rescue,required"`
	Squadron *struct {
		SquadronBankCreditsDeposited                                   int `json:"Squadron_Bank_Credits_Deposited,required"`
		SquadronBankCreditsWithdrawn                                   int `json:"Squadron_Bank_Credits_Withdrawn,required"`
		SquadronBankCommoditiesDepositedNum                            int `json:"Squadron_Bank_Commodities_Deposited_Num,required"`
		SquadronBankCommoditiesDepositedValue                          int `json:"Squadron_Bank_Commodities_Deposited_Value,required"`
		SquadronBankCommoditiesWithdrawnNum                            int `json:"Squadron_Bank_Commodities_Withdrawn_Num,required"`
		SquadronBankCommoditiesWithdrawnValue                          int `json:"Squadron_Bank_Commodities_Withdrawn_Value,required"`
		SquadronBankPersonalAssetsDepositedNum                         int `json:"Squadron_Bank_PersonalAssets_Deposited_Num,required"`
		SquadronBankPersonalAssetsDepositedValue                       int `json:"Squadron_Bank_PersonalAssets_Deposited_Value,required"`
		SquadronBankPersonalAssetsWithdrawnNum                         int `json:"Squadron_Bank_PersonalAssets_Withdrawn_Num,required"`
		SquadronBankPersonalAssetsWithdrawnValue                       int `json:"Squadron_Bank_PersonalAssets_Withdrawn_Value,required"`
		SquadronBankShipsDepositedNum                                  int `json:"Squadron_Bank_Ships_Deposited_Num,required"`
		SquadronBankShipsDepositedValue                                int `json:"Squadron_Bank_Ships_Deposited_Value,required"`
		SquadronLeaderboardAegisHighestcontribution                    int `json:"Squadron_Leaderboard_aegis_highestcontribution,required"`
		SquadronLeaderboardBgsHighestcontribution                      int `json:"Squadron_Leaderboard_bgs_highestcontribution,required"`
		SquadronLeaderboardBountyHighestcontribution                   int `json:"Squadron_Leaderboard_bounty_highestcontribution,required"`
		SquadronLeaderboardColonisationContributionHighestcontribution int `json:"Squadron_Leaderboard_colonisation_contribution_highestcontribution,required"`
		SquadronLeaderboardCombatHighestcontribution                   int `json:"Squadron_Leaderboard_combat_highestcontribution,required"`
		SquadronLeaderboardCqcHighestcontribution                      int `json:"Squadron_Leaderboard_cqc_highestcontribution,required"`
		SquadronLeaderboardExplorationHighestcontribution              int `json:"Squadron_Leaderboard_exploration_highestcontribution,required"`
		SquadronLeaderboardMiningHighestcontribution                   int `json:"Squadron_Leaderboard_mining_highestcontribution,required"`
		SquadronLeaderboardPowerplayHighestcontribution                int `json:"Squadron_Leaderboard_powerplay_highestcontribution,required"`
		SquadronLeaderboardTradeHighestcontribution                    int `json:"Squadron_Leaderboard_trade_highestcontribution,required"`
		SquadronLeaderboardTradeIllicitHighestcontribution             int `json:"Squadron_Leaderboard_trade_illicit_highestcontribution,required"`
		SquadronLeaderboardPodiums                                     int `json:"Squadron_Leaderboard_podiums,required"`
	} `json:"Squadron,omitempty"` // TODO: not preset when not in squadron?
	TgEncounters *struct {
		TgEncounterKilled             int    `json:"TG_ENCOUNTER_KILLED,required"`
		TgEncounterTotal              *int   `json:"TG_ENCOUNTER_TOTAL,omitempty"`                // may be not present (?)
		TgEncounterTotalLastSystem    string `json:"TG_ENCOUNTER_TOTAL_LAST_SYSTEM,omitempty"`    // may be not present (?)
		TgEncounterTotalLastTimestamp string `json:"TG_ENCOUNTER_TOTAL_LAST_TIMESTAMP,omitempty"` // may be not present (?)
		TgEncounterTotalLastShip      string `json:"TG_ENCOUNTER_TOTAL_LAST_SHIP,omitempty"`      // may be not present (?)
	} `json:"TG_ENCOUNTERS,omitempty"` // TODO: not present in older versions?
	Crafting struct {
		CountOfUsedEngineers  int `json:"Count_Of_Used_Engineers,required"`
		RecipesGenerated      int `json:"Recipes_Generated,required"`
		RecipesGeneratedRank1 int `json:"Recipes_Generated_Rank_1,required"`
		RecipesGeneratedRank2 int `json:"Recipes_Generated_Rank_2,required"`
		RecipesGeneratedRank3 int `json:"Recipes_Generated_Rank_3,required"`
		RecipesGeneratedRank4 int `json:"Recipes_Generated_Rank_4,required"`
		RecipesGeneratedRank5 int `json:"Recipes_Generated_Rank_5,required"`
		SuitModsApplied       int `json:"Suit_Mods_Applied,required"`
		WeaponModsApplied     int `json:"Weapon_Mods_Applied,required"`
		SuitsUpgraded         int `json:"Suits_Upgraded,required"`
		WeaponsUpgraded       int `json:"Weapons_Upgraded,required"`
		SuitsUpgradedFull     int `json:"Suits_Upgraded_Full,required"`
		WeaponsUpgradedFull   int `json:"Weapons_Upgraded_Full,required"`
		SuitModsAppliedFull   int `json:"Suit_Mods_Applied_Full,required"`
		WeaponModsAppliedFull int `json:"Weapon_Mods_Applied_Full,required"`
	} `json:"Crafting,required"`
	Crew struct {
		NpcCrewTotalWages int64 `json:"NpcCrew_TotalWages,required"`
		NpcCrewHired      int   `json:"NpcCrew_Hired,required"`
		NpcCrewFired      int   `json:"NpcCrew_Fired,required"`
		NpcCrewDied       int   `json:"NpcCrew_Died,required"`
	} `json:"Crew,required"`
	Multicrew struct {
		MulticrewTimeTotal        int `json:"Multicrew_Time_Total,required"`
		MulticrewGunnerTimeTotal  int `json:"Multicrew_Gunner_Time_Total,required"`
		MulticrewFighterTimeTotal int `json:"Multicrew_Fighter_Time_Total,required"`
		MulticrewCreditsTotal     int `json:"Multicrew_Credits_Total,required"`
		MulticrewFinesTotal       int `json:"Multicrew_Fines_Total,required"`
	} `json:"Multicrew,required"`
	MaterialTraderStats *struct {
		TradesCompleted        int  `json:"Trades_Completed,required"`
		MaterialsTraded        int  `json:"Materials_Traded,required"`
		EncodedMaterialsTraded *int `json:"Encoded_Materials_Traded,omitempty"` // todo: not present in older event version?
		RawMaterialsTraded     *int `json:"Raw_Materials_Traded,omitempty"`     // todo: not present in older event version?
		Grade1MaterialsTraded  *int `json:"Grade_1_Materials_Traded,omitempty"` // todo: not present in older event version?
		Grade2MaterialsTraded  *int `json:"Grade_2_Materials_Traded,omitempty"` // todo: not present in older event version?
		Grade3MaterialsTraded  *int `json:"Grade_3_Materials_Traded,omitempty"` // todo: not present in older event version?
		Grade4MaterialsTraded  *int `json:"Grade_4_Materials_Traded,omitempty"` // todo: not present in older event version?
		Grade5MaterialsTraded  *int `json:"Grade_5_Materials_Traded,omitempty"` // todo: not present in older event version?
		AssetsTradedIn         int  `json:"Assets_Traded_In,required"`
		AssetsTradedOut        int  `json:"Assets_Traded_Out,required"`
	} `json:"Material_Trader_Stats,required"`
	Cqc *struct {
		CQCCreditsEarned int     `json:"CQC_Credits_Earned,required"`
		CQCTimePlayed    int     `json:"CQC_Time_Played,required"`
		CqcKd            float64 `json:"CQC_KD,required"`
		CQCKills         int     `json:"CQC_Kills,required"`
		CqcWl            int     `json:"CQC_WL,required"`
	} `json:"CQC,omitempty"` // TODO: not present in some situations?
	Fleetcarrier *struct {
		FleetcarrierExportTotal       int     `json:"FLEETCARRIER_EXPORT_TOTAL,required"`
		FleetcarrierImportTotal       int     `json:"FLEETCARRIER_IMPORT_TOTAL,required"`
		FleetcarrierTradeprofitTotal  int     `json:"FLEETCARRIER_TRADEPROFIT_TOTAL,required"`
		FleetcarrierTradespendTotal   int     `json:"FLEETCARRIER_TRADESPEND_TOTAL,required"`
		FleetcarrierStolenprofitTotal int     `json:"FLEETCARRIER_STOLENPROFIT_TOTAL,required"`
		FleetcarrierStolenspendTotal  int     `json:"FLEETCARRIER_STOLENSPEND_TOTAL,required"`
		FleetcarrierDistanceTravelled float64 `json:"FLEETCARRIER_DISTANCE_TRAVELLED,required"`
		FleetcarrierTotalJumps        int     `json:"FLEETCARRIER_TOTAL_JUMPS,required"`
		FleetcarrierShipyardSold      int     `json:"FLEETCARRIER_SHIPYARD_SOLD,required"`
		FleetcarrierShipyardProfit    int     `json:"FLEETCARRIER_SHIPYARD_PROFIT,required"`
		FleetcarrierOutfittingSold    int     `json:"FLEETCARRIER_OUTFITTING_SOLD,required"`
		FleetcarrierOutfittingProfit  int     `json:"FLEETCARRIER_OUTFITTING_PROFIT,required"`
		FleetcarrierRearmTotal        int     `json:"FLEETCARRIER_REARM_TOTAL,required"`
		FleetcarrierRefuelTotal       int     `json:"FLEETCARRIER_REFUEL_TOTAL,required"`
		FleetcarrierRefuelProfit      int     `json:"FLEETCARRIER_REFUEL_PROFIT,required"`
		FleetcarrierRepairsTotal      int     `json:"FLEETCARRIER_REPAIRS_TOTAL,required"`
		FleetcarrierVouchersRedeemed  int     `json:"FLEETCARRIER_VOUCHERS_REDEEMED,required"`
		FleetcarrierVouchersProfit    int     `json:"FLEETCARRIER_VOUCHERS_PROFIT,required"`
	} `json:"FLEETCARRIER,omitempty"` // TODO: not present when not having a FC?
	Exobiology struct {
		OrganicGenusEncountered   int `json:"Organic_Genus_Encountered,required"`
		OrganicSpeciesEncountered int `json:"Organic_Species_Encountered,required"`
		OrganicVariantEncountered int `json:"Organic_Variant_Encountered,required"`
		OrganicDataProfits        int `json:"Organic_Data_Profits,required"`
		OrganicData               int `json:"Organic_Data,required"`
		FirstLoggedProfits        int `json:"First_Logged_Profits,required"`
		FirstLogged               int `json:"First_Logged,required"`
		OrganicSystems            int `json:"Organic_Systems,required"`
		OrganicPlanets            int `json:"Organic_Planets,required"`
		OrganicGenus              int `json:"Organic_Genus,required"`
		OrganicSpecies            int `json:"Organic_Species,required"`
	} `json:"Exobiology,required"`
}

type ColonisationConstructionDepot struct {
	BaseEvent `json:",inline"`

	MarketID             int64   `json:"MarketID,required"`
	ConstructionProgress float64 `json:"ConstructionProgress,required"`
	ConstructionComplete bool    `json:"ConstructionComplete,required"`
	ConstructionFailed   bool    `json:"ConstructionFailed,required"`
	ResourcesRequired    []struct {
		Name           string `json:"Name,required,unique"`
		NameLocalised  string `json:"Name_Localised,required,unique"`
		RequiredAmount int    `json:"RequiredAmount,required"`
		ProvidedAmount int    `json:"ProvidedAmount,required"`
		Payment        int    `json:"Payment,required"`
	} `json:"ResourcesRequired,required"`
}

type ColonisationContribution struct {
	BaseEvent `json:",inline"`

	MarketID      int64 `json:"MarketID,required"`
	Contributions []struct {
		Name          string `json:"Name,required,unique"`
		NameLocalised string `json:"Name_Localised,required,unique"`
		Amount        int    `json:"Amount,required"`
	} `json:"Contributions,required"`
}

type SellMicroResources struct {
	BaseEvent `json:",inline"`

	TotalCount     int `json:"TotalCount,required"`
	MicroResources []struct {
		Name          string `json:"Name,required"`
		NameLocalised string `json:"Name_Localised,omitempty"` // may be not present
		Category      string `json:"Category,required"`
		Count         int    `json:"Count,required"`
	} `json:"MicroResources,required"`
	Price    int   `json:"Price,required"`
	MarketID int64 `json:"MarketID,required"`
}

type MultiSellExplorationData struct {
	BaseEvent `json:",inline"`

	Discovered []struct {
		SystemName string `json:"SystemName,required,unique"`
		NumBodies  int    `json:"NumBodies,required"`
	} `json:"Discovered,required"`
	BaseValue     int `json:"BaseValue,required"`
	Bonus         int `json:"Bonus,required"`
	TotalEarnings int `json:"TotalEarnings,required"`
}

type BuyExplorationData struct {
	BaseEvent `json:",inline"`

	System string `json:"System,required,unique"`
	Cost   int    `json:"Cost,required"`
}

type CarrierBankTransfer struct {
	BaseEvent `json:",inline"`

	CarrierID      int64  `json:"CarrierID,required"`
	CarrierType    string `json:"CarrierType,required,unique"`
	Deposit        *int   `json:"Deposit,omitempty"` // may be not present
	PlayerBalance  int64  `json:"PlayerBalance,required"`
	CarrierBalance int    `json:"CarrierBalance,required"`
	Withdraw       *int   `json:"Withdraw,omitempty"` // may be not present
}

type PVPKill struct {
	BaseEvent `json:",inline"`

	Victim     string `json:"Victim,required,unique"`
	CombatRank int    `json:"CombatRank,required"`
}

type StoredModules struct {
	BaseEvent `json:",inline"`

	MarketID    int64  `json:"MarketID,required"`
	StationName string `json:"StationName,required,unique"`
	StarSystem  string `json:"StarSystem,required,unique"`
	Items       []struct {
		Name                  string   `json:"Name,required,unique"`
		NameLocalised         string   `json:"Name_Localised,required,unique"`
		StorageSlot           int      `json:"StorageSlot,required"`
		StarSystem            string   `json:"StarSystem,required,unique"`
		MarketID              int64    `json:"MarketID,required"`
		TransferCost          int      `json:"TransferCost,required"`
		TransferTime          int      `json:"TransferTime,required"`
		BuyPrice              int      `json:"BuyPrice,required"`
		Hot                   bool     `json:"Hot,required"`
		EngineerModifications string   `json:"EngineerModifications,omitempty,unique"` // may be not present
		Level                 *int     `json:"Level,omitempty"`                        // may be not present
		Quality               *float64 `json:"Quality,omitempty"`                      // may be not present
	} `json:"Items"`
}

type Bounty struct {
	BaseEvent `json:",inline"`

	Rewards []struct {
		Faction string `json:"Faction,require,uniqued"`
		Reward  int    `json:"Reward,required"`
	} `json:"Rewards,required"`
	PilotName              string `json:"PilotName,required,unique"`
	PilotNameLocalised     string `json:"PilotName_Localised,required,unique"`
	Target                 string `json:"Target,required,unique"`
	TargetLocalised        string `json:"Target_Localised,omitempty,unique"` // may be not present
	TotalReward            int    `json:"TotalReward,required"`
	VictimFaction          string `json:"VictimFaction,required,unique"`
	VictimFactionLocalised string `json:"VictimFaction_Localised,omitempty,unique"` // may be not present
}

type FactionKillBond struct {
	BaseEvent `json:",inline"`

	Reward                   int    `json:"Reward,required"`
	AwardingFaction          string `json:"AwardingFaction,required,unique"`
	AwardingFactionLocalised string `json:"AwardingFaction_Localised,omitempty,unique"` // may be not present
	VictimFaction            string `json:"VictimFaction,required,unique"`
	VictimFactionLocalised   string `json:"VictimFaction_Localised,omitempty,unique"` // may be not present
}

type Location struct {
	BaseEvent `json:",inline"`

	Docked                       bool      `json:"Docked,required"`
	StarSystem                   string    `json:"StarSystem,required,unique"`
	SystemAddress                int64     `json:"SystemAddress,required"`
	StarPos                      []float64 `json:"StarPos,required"`
	SystemAllegiance             string    `json:"SystemAllegiance,required,unique"`
	SystemEconomy                string    `json:"SystemEconomy,required,unique"`
	SystemEconomyLocalised       string    `json:"SystemEconomy_Localised,required,unique"`
	SystemSecondEconomy          string    `json:"SystemSecondEconomy,required,unique"`
	SystemSecondEconomyLocalised string    `json:"SystemSecondEconomy_Localised,required,unique"`
	SystemGovernment             string    `json:"SystemGovernment,required,unique"`
	SystemGovernmentLocalised    string    `json:"SystemGovernment_Localised,required,unique"`
	SystemSecurity               string    `json:"SystemSecurity,required,unique"`
	SystemSecurityLocalised      string    `json:"SystemSecurity_Localised,required,unique"`
	Population                   int       `json:"Population,required"`
	Body                         string    `json:"Body,required,unique"`
	BodyID                       int       `json:"BodyID,required"`
	BodyType                     string    `json:"BodyType,required,unique"`

	DistFromStarLS *float64 `json:"DistFromStarLS,omitempty"`     // may be not present
	StationName    string   `json:"StationName,omitempty,unique"` // may be not present
	StationType    string   `json:"StationType,omitempty,unique"` // may be not present
	MarketID       *int64   `json:"MarketID,omitempty"`           // may be not present
	Taxi           bool     `json:"Taxi,omitempty"`               // may be not present
	Multicrew      bool     `json:"Multicrew,omitempty"`          // may be not present
	OnFoot         bool     `json:"OnFoot,omitempty"`             // may be not present
	Latitude       *float64 `json:"Latitude,omitempty"`           // may be not present
	Longitude      *float64 `json:"Longitude,omitempty"`          // may be not present

	// TODO: rest of fields that might be or might not be present
	//StationFaction struct {
	//	Name string `json:"Name"`
	//} `json:"StationFaction,omitempty"`
	//StationGovernment          string   `json:"StationGovernment,omitempty"`
	//StationGovernmentLocalised string   `json:"StationGovernment_Localised,omitempty"`
	//StationServices            []string `json:"StationServices,omitempty"`
	//StationEconomy             string   `json:"StationEconomy,omitempty"`
	//StationEconomyLocalised    string   `json:"StationEconomy_Localised,omitempty"`
	//StationEconomies           []struct {
	//	Name          string  `json:"Name"`
	//	NameLocalised string  `json:"Name_Localised"`
	//	Proportion    float64 `json:"Proportion"`
	//} `json:"StationEconomies,omitempty"`
	//Factions                     []struct {
	//	Name               string  `json:"Name"`
	//	FactionState       string  `json:"FactionState"`
	//	Government         string  `json:"Government"`
	//	Influence          float64 `json:"Influence"`
	//	Allegiance         string  `json:"Allegiance"`
	//	Happiness          string  `json:"Happiness"`
	//	HappinessLocalised string  `json:"Happiness_Localised"`
	//	MyReputation       float64 `json:"MyReputation"`
	//	SquadronFaction    bool    `json:"SquadronFaction,omitempty"`
	//	PendingStates      []struct {
	//		State string `json:"State"`
	//		Trend int    `json:"Trend"`
	//	} `json:"PendingStates,omitempty"`
	//	ActiveStates []struct {
	//		State string `json:"State"`
	//	} `json:"ActiveStates,omitempty"`
	//} `json:"Factions,omitempty"`
	//SystemFaction struct {
	//	Name string `json:"Name"`
	//} `json:"SystemFaction,omitempty"`
	//ControllingPower              string   `json:"ControllingPower,omitempty"`
	//Powers                        []string `json:"Powers,omitempty"`
	//PowerplayState                string   `json:"PowerplayState,omitempty"`
	//PowerplayStateControlProgress float64  `json:"PowerplayStateControlProgress,omitempty"`
	//PowerplayStateReinforcement   int      `json:"PowerplayStateReinforcement,omitempty"`
	//PowerplayStateUndermining     int      `json:"PowerplayStateUndermining,omitempty"`
	//Conflicts                     []struct {
	//	WarType  string `json:"WarType"`
	//	Status   string `json:"Status"`
	//	Faction1 struct {
	//		Name    string `json:"Name"`
	//		Stake   string `json:"Stake"`
	//		WonDays int    `json:"WonDays"`
	//	} `json:"Faction1"`
	//	Faction2 struct {
	//		Name    string `json:"Name"`
	//		Stake   string `json:"Stake"`
	//		WonDays int    `json:"WonDays"`
	//	} `json:"Faction2"`
	//} `json:"Conflicts,omitempty"`
	//StationAllegiance         string `json:"StationAllegiance,omitempty"`
	//PowerplayConflictProgress []struct {
	//	Power            string  `json:"Power"`
	//	ConflictProgress float64 `json:"ConflictProgress"`
	//} `json:"PowerplayConflictProgress,omitempty"`
	//ThargoidWar struct {
	//	CurrentState        string  `json:"CurrentState"`
	//	NextStateSuccess    string  `json:"NextStateSuccess"`
	//	NextStateFailure    string  `json:"NextStateFailure"`
	//	SuccessStateReached bool    `json:"SuccessStateReached"`
	//	WarProgress         float64 `json:"WarProgress"`
	//	RemainingPorts      int     `json:"RemainingPorts"`
	//} `json:"ThargoidWar,omitempty"`

}

type ShipyardTransfer struct {
	BaseEvent `json:",inline"`

	ShipType          string  `json:"ShipType,required"`
	ShipID            int     `json:"ShipID,required"`
	System            string  `json:"System,required,unique"`
	ShipMarketID      float64 `json:"ShipMarketID,required"`
	Distance          float64 `json:"Distance,required"`
	TransferPrice     int     `json:"TransferPrice,required"`
	TransferTime      int     `json:"TransferTime,required"`
	MarketID          int64   `json:"MarketID,required"`
	ShipTypeLocalised string  `json:"ShipType_Localised,omitempty,unique"` // may be not present
}

type Loadout struct {
	BaseEvent `json:",inline"`

	Ship          string  `json:"Ship,required"`
	ShipID        int     `json:"ShipID,required"`
	ShipName      string  `json:"ShipName,required"`
	ShipIdent     string  `json:"ShipIdent,required"`
	HullValue     *int    `json:"HullValue,omitempty"`    // may be not present
	ModulesValue  *int    `json:"ModulesValue,omitempty"` // may be not present
	HullHealth    float64 `json:"HullHealth,required"`
	UnladenMass   float64 `json:"UnladenMass,required"`
	CargoCapacity int     `json:"CargoCapacity,required"`
	MaxJumpRange  float64 `json:"MaxJumpRange,required"`
	FuelCapacity  struct {
		Main    float64 `json:"Main,required"`
		Reserve float64 `json:"Reserve,required"`
	} `json:"FuelCapacity,required"`
	Rebuy   int `json:"Rebuy,required"`
	Modules []struct {
		Slot         string  `json:"Slot,required"`
		Item         string  `json:"Item,required"`
		On           bool    `json:"On,required"`
		Priority     int     `json:"Priority,required"`
		Health       float64 `json:"Health,required"`
		Value        *int    `json:"Value,omitempty"`        // may be not present
		AmmoInClip   *int    `json:"AmmoInClip,omitempty"`   // may be not present
		AmmoInHopper *int    `json:"AmmoInHopper,omitempty"` // may be not present
	} `json:"Modules,required"`
	Hot *bool `json:"Hot,omitempty"` // may be not present
}

type ShipTargeted struct {
	BaseEvent `json:",inline"`

	TargetLocked bool `json:"TargetLocked,required"`

	Ship               string   `json:"Ship,omitempty"`                // may be not present
	ShipLocalised      string   `json:"Ship_Localised,omitempty"`      // may be not present
	ScanStage          *int     `json:"ScanStage,omitempty"`           // may be not present
	PilotName          string   `json:"PilotName,omitempty"`           // may be not present
	PilotNameLocalised string   `json:"PilotName_Localised,omitempty"` // may be not present
	PilotRank          string   `json:"PilotRank,omitempty"`           // may be not present
	ShieldHealth       float64  `json:"ShieldHealth,omitempty"`        // may be not present
	HullHealth         float64  `json:"HullHealth,omitempty"`          // may be not present
	LegalStatus        string   `json:"LegalStatus,omitempty"`         // may be not present
	Faction            string   `json:"Faction,omitempty"`             // may be not present
	Bounty             *int     `json:"Bounty,omitempty"`              // may be not present
	Subsystem          string   `json:"Subsystem,omitempty"`           // may be not present
	SubsystemLocalised string   `json:"Subsystem_Localised,omitempty"` // may be not present
	SubsystemHealth    *float64 `json:"SubsystemHealth,omitempty"`     // may be not present
	SquadronID         string   `json:"SquadronID,omitempty"`          // may be not present
	Power              string   `json:"Power,omitempty"`               // may be not present
}

type Cargo struct {
	BaseEvent `json:",inline"`

	Vessel    string `json:"Vessel,required"`
	Count     int    `json:"Count,required"`
	Inventory []struct {
		Name          string `json:"Name,required"`
		NameLocalised string `json:"Name_Localised,required"`
		Count         int    `json:"Count,required"`
		Stolen        int    `json:"Stolen,required"`
	} `json:"Inventory,omitempty"`
}

type MissionAbandoned struct {
	BaseEvent `json:",inline"`

	Name          string `json:"Name"`
	LocalisedName string `json:"LocalisedName"`
	MissionID     uint64 `json:"MissionID"`
}

type MissionAccepted struct {
	BaseEvent `json:",inline"`

	Faction               string     `json:"Faction"`
	Name                  string     `json:"Name"`
	LocalisedName         string     `json:"LocalisedName"`
	Commodity             string     `json:"Commodity,omitempty"`                            // may not be present
	CommodityLocalised    string     `json:"Commodity_Localised,omitempty"`                  // may not be present
	Count                 *int       `json:"Count,omitempty"`                                // may not be present
	DestinationSystem     string     `json:"DestinationSystem,omitempty"`                    // may not be present
	DestinationSettlement string     `json:"DestinationSettlement,omitempty"`                // may not be present
	Target                string     `json:"Target,omitempty"`                               // may not be present
	Expiry                *time.Time `json:"Expiry,omitempty,format:'2006-01-02T15:04:05Z'"` // may not be present
	Wing                  bool       `json:"Wing"`
	Influence             string     `json:"Influence"`
	Reputation            string     `json:"Reputation"`
	Reward                *int       `json:"Reward,omitempty"` // may not be present
	MissionID             uint64     `json:"MissionID"`
	Donation              string     `json:"Donation,omitempty"`             // may not be present
	TargetType            string     `json:"TargetType,omitempty"`           // may not be present
	TargetTypeLocalised   string     `json:"TargetType_Localised,omitempty"` // may not be present
	TargetFaction         string     `json:"TargetFaction,omitempty"`        // may not be present
	KillCount             *int       `json:"KillCount,omitempty"`            // may not be present
	DestinationStation    string     `json:"DestinationStation,omitempty"`   // may not be present
	TargetLocalised       string     `json:"Target_Localised,omitempty"`     // may not be present
}

type Scan struct {
	BaseEvent `json:",inline"`

	ScanType string `json:"ScanType"`
	BodyName string `json:"BodyName"`
	BodyID   int    `json:"BodyID"`
	Parents  []struct {
		Null int `json:"Null"`
	} `json:"Parents,omitempty"`
	StarSystem            string  `json:"StarSystem"`
	SystemAddress         int64   `json:"SystemAddress"`
	DistanceFromArrivalLS float64 `json:"DistanceFromArrivalLS"`
	StarType              string  `json:"StarType,omitempty"`
	Subclass              int     `json:"Subclass,omitempty"`
	StellarMass           float64 `json:"StellarMass,omitempty"`
	Radius                float64 `json:"Radius,omitempty"`
	AbsoluteMagnitude     float64 `json:"AbsoluteMagnitude,omitempty"`
	AgeMY                 int     `json:"Age_MY,omitempty"`
	SurfaceTemperature    float64 `json:"SurfaceTemperature,omitempty"`
	Luminosity            string  `json:"Luminosity,omitempty"`
	SemiMajorAxis         float64 `json:"SemiMajorAxis,omitempty"`
	Eccentricity          float64 `json:"Eccentricity,omitempty"`
	OrbitalInclination    float64 `json:"OrbitalInclination,omitempty"`
	Periapsis             float64 `json:"Periapsis,omitempty"`
	OrbitalPeriod         float64 `json:"OrbitalPeriod,omitempty"`
	AscendingNode         float64 `json:"AscendingNode,omitempty"`
	MeanAnomaly           float64 `json:"MeanAnomaly,omitempty"`
	RotationPeriod        float64 `json:"RotationPeriod,omitempty"`
	AxialTilt             float64 `json:"AxialTilt,omitempty"`
	WasDiscovered         bool    `json:"WasDiscovered"`
	WasMapped             bool    `json:"WasMapped"`
	Rings                 []struct {
		Name      string `json:"Name"`
		RingClass string `json:"RingClass"`
		MassMT    int64  `json:"MassMT"`
		InnerRad  int    `json:"InnerRad"`
		OuterRad  int    `json:"OuterRad"`
	} `json:"Rings,omitempty"`
	TidalLock       bool    `json:"TidalLock,omitempty"`
	TerraformState  string  `json:"TerraformState,omitempty"`
	PlanetClass     string  `json:"PlanetClass,omitempty"`
	Atmosphere      string  `json:"Atmosphere,omitempty"`
	AtmosphereType  string  `json:"AtmosphereType,omitempty"`
	Volcanism       string  `json:"Volcanism,omitempty"`
	MassEM          float64 `json:"MassEM,omitempty"`
	SurfaceGravity  float64 `json:"SurfaceGravity,omitempty"`
	SurfacePressure float64 `json:"SurfacePressure,omitempty"`
	Landable        bool    `json:"Landable,omitempty"`
	Materials       []struct {
		Name    string  `json:"Name"`
		Percent float64 `json:"Percent"`
	} `json:"Materials,omitempty"`
	Composition struct {
		Ice   float64 `json:"Ice"`
		Rock  float64 `json:"Rock"`
		Metal float64 `json:"Metal"`
	} `json:"Composition,omitempty"`
	AtmosphereComposition []struct {
		Name    string  `json:"Name"`
		Percent float64 `json:"Percent"`
	} `json:"AtmosphereComposition,omitempty"`
	ReserveLevel  string `json:"ReserveLevel,omitempty"`
	WasFootfalled bool   `json:"WasFootfalled,omitempty"`
}

type MissionInfo struct {
	MissionID        uint64 `json:"MissionID,required"`
	Name             string `json:"Name,required"`
	PassengerMission bool   `json:"PassengerMission,required"`
	Expires          int    `json:"Expires,required"`
}

type Missions struct {
	BaseEvent `json:",inline"`

	Active []MissionInfo `json:"Active,required"`
	Failed []MissionInfo `json:"Failed,required"`
}

type WingInvite struct {
	BaseEvent `json:",inline"`

	Name string `json:"Name,required"`
}

type InvitedToSquadron struct {
	BaseEvent `json:",inline"`

	SquadronName string `json:"SquadronName,required"`
}
