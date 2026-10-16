package commander_profile

import (
	"reflect"
	"strconv"
)

type (
	RankCombat       uint
	RankCQC          uint
	RankExobiologist uint
	RankExplorer     uint
	RankMercenary    uint
	RankTrader       uint
	RankFederalNavy  uint
	RankImperialNavy uint
)

func ConvertToRankNameAndValue(rank CommanderRankPilot) (string, string) {
	i, err := strconv.Atoi(rank.RankValue)
	if err != nil {
		panic(err)
	}
	switch rank.RankName {
	case "combat":
		return "Combat", RankCombatToString[RankCombat(i)]
	case "cqc":
		return "CQC", RankCQCToString[RankCQC(i)]
	case "exobiologist":
		return "Exobiologist", RankExobiologistToString[RankExobiologist(i)]
	case "exploration":
		return "Explorer", RankExplorerToString[RankExplorer(i)]
	case "soldier":
		return "Mercenary", RankMercenaryToString[RankMercenary(i)]
	case "trade":
		return "Trader", RankTraderToString[RankTrader(i)]
	case "federation":
		return "Federal Navy", RankFederalNavyToString[RankFederalNavy(i)]
	case "empire":
		return "Imperial Navy", RankImperialNavyToString[RankImperialNavy(i)]
	default:
		panic("unknown rank name")
	}
}

func ConvertToInt(rank CommanderRankPilot) (int, int) {
	i, err := strconv.Atoi(rank.RankValue)
	if err != nil {
		panic(err)
	}
	switch rank.RankName {
	case "combat":
		return 1, i
	case "trade":
		return 2, i
	case "exploration":
		return 3, i
	case "cqc":
		return 4, i
	case "soldier":
		return 5, i
	case "exobiologist":
		return 6, i
	case "federation":
		return 12, 0
	case "empire":
		return 13, 0
	case "alliance":
		return 14, 0
	case "independent":
		return 15, 0
	default:
		panic("unknown rank name")
	}
}

func RankToString(rank any) string {
	switch rank.(type) {
	case RankCombat:
		return "Combat"
	case RankCQC:
		return "CQC"
	case RankExobiologist:
		return "Exobiologist"
	case RankExplorer:
		return "Explorer"
	case RankMercenary:
		return "Mercenary"
	case RankTrader:
		return "Trader"
	case RankFederalNavy:
		return "Federal Navy"
	case RankImperialNavy:
		return "Imperial Navy"
	default:
		panic("unknown rank type")
	}
}

var RankIntToName = map[int]string{
	0: "combat",
	1: "cqc",
	2: "exobiologist",
	3: "exploration",
	4: "soldier",
	5: "trade",
	6: "federation",
	7: "empire",
}

var RankNameToType = map[string]reflect.Type{
	"combat":       reflect.TypeFor[RankCombat](),
	"cqc":          reflect.TypeFor[RankCQC](),
	"exobiologist": reflect.TypeFor[RankExobiologist](),
	"exploration":  reflect.TypeFor[RankExplorer](),
	"soldier":      reflect.TypeFor[RankMercenary](),
	"trade":        reflect.TypeFor[RankTrader](),
	"federation":   reflect.TypeFor[RankFederalNavy](),
	"empire":       reflect.TypeFor[RankImperialNavy](),
}

const (
	CombatHarmless RankCombat = iota
	CombatMostlyHarmless
	CombatNovice
	CombatCompetent
	CombatExpert
	CombatMaster
	CombatDangerous
	CombatDeadly
	CombatElite
	CombatEliteI
	CombatEliteII
	CombatEliteIII
	CombatEliteIV
	CombatEliteV
)

const (
	CQCHelpless         RankCQC = iota // 0
	CQCMostlyHelpless                  // 5,000
	CQCAmateur                         // 10,000
	CQCSemiProfessional                // 40,000
	CQCProfessional                    // 100,000
	CQCChampion                        // 285,000
	CQCHero                            // 685,000
	CQCLegend                          // 1,685,000
	CQCElite                           // 3,685,000
	CQCEliteI                          // 7,370,000
	CQCEliteII                         // 11,055,000
	CQCEliteIII                        // 14,740,000
	CQCEliteIV                         // 18,425,000
	CQCEliteV                          // 22,110,000
)

const (
	ExobiologistDirectionless       RankExobiologist = iota // 0
	ExobiologistMostlyDirectionless                         // 22,500,000
	ExobiologistCompiler                                    // 83,475,000
	ExobiologistCollector                                   // 210,560,000
	ExobiologistCataloguer                                  // 532,800,000
	ExobiologistTaxonomist                                  // 1,144,000,000
	ExobiologistEcologist                                   // 2,262,600,000
	ExobiologistGeneticist                                  // 3,996,000,000
	ExobiologistElite                                       // 8,425,000,000
	ExobiologistEliteI                                      // 12,969,000,000
	ExobiologistEliteII                                     // 17,425,000,000
	ExobiologistEliteIII                                    // 21,925,000,000
	ExobiologistEliteIV                                     // 26,070,000,000
	ExobiologistEliteV                                      // 30,553,600,000
)

const (
	ExplorerAimless       RankExplorer = iota // 0
	ExplorerMostlyAimless                     // 40,000
	ExplorerScout                             // 270,000
	ExplorerSurveyor                          // 1,140,000
	ExplorerTrailblazer                       // 4,200,000
	ExplorerPathfinder                        // 10,000,000
	ExplorerRanger                            // 35,000,000
	ExplorerPioneer                           // 116,000,000
	ExplorerElite                             // 320,000,000
	ExplorerEliteI                            // 640,000,000
	ExplorerEliteII                           // 945,000,000
	ExplorerEliteIII                          // 1,255,400,000
	ExplorerEliteIV                           // 1,570,000,000
	ExplorerEliteV                            // 1,884,000,000
)

const (
	MercenaryDefenceless       RankMercenary = iota // 0	Dominator Suit Body Suit Livery (Bronze)
	MercenaryMostlyDefenceless                      // 10,000,000	Dominator Suit Torso, Arms, and Legs Livery (Bronze)
	MercenaryRookie                                 // 30,000,000	Dominator Suit Helmet Livery (Bronze)
	MercenarySoldier                                // 60,000,000	Dominator Suit Body Suit Livery (Silver)
	MercenaryGunslinger                             // 125,000,000	Dominator Suit Torso, Arms, and Legs Livery (Silver)
	MercenaryWarrior                                // 350,000,000	Dominator Suit Helmet Livery (Silver)
	MercenaryGladiator                              // 520,000,000	Dominator Suit Body Suit Livery (Gold)
	MercenaryDeadeye                                // 888,000,000	Dominator Suit Torso, Arms, and Legs Livery (Gold)
	MercenaryElite                                  // 1,550,000,000	Dominator Suit Helmet Livery (Gold)
	MercenaryEliteI                                 // 2,605,000,000
	MercenaryEliteII                                // 3,537,000,000
	MercenaryEliteIII                               // 5,364,000,000
	MercenaryEliteIV                                // 6,734,000,000
	MercenaryEliteV                                 // 8,118,000,000
)

const (
	TraderPenniless       RankTrader = iota // 0	"Penniless" Ship & Suit Decals
	TraderMostlyPenniless                   // 5,000	"Mostly Penniless" Ship & Suit Decals
	TraderPeddler                           // 100,000	"Peddler" Ship & Suit Decals
	TraderDealer                            // 800,000	"Dealer" Ship & Suit Decals
	TraderMerchant                          // 3,700,000	"Merchant" Ship & Suit Decals
	TraderBroker                            // 30,000,000	"Broker" Ship & Suit Decals
	TraderEntrepreneur                      // 140,000,000	"Entrepreneur" Ship & Suit Decals
	TraderTycoon                            // 390,000,000	"Tycoon" Ship & Suit Decals
	TraderElite                             // 1,000,000,000	"Elite" (Trader) Ship & Suit Decals
	TraderEliteI                            // ~2,000,000,000	"Elite I" (Trader) Ship & Suit Decal
	TraderEliteII                           // ~3,000,000,000	"Elite II" (Trader) Ship & Suit Decal
	TraderEliteIII                          // ~4,000,000,000	"Elite III" (Trader) Ship & Suit Decal
	TraderEliteIV                           // ~5,000,000,000 (~ 4,820,000,000 promotion to Elite IV) "Elite IV" (Trader) Ship & Suit Decal
	TraderEliteV                            // ~6,000,000,000 (~5,802,300,000 promotion to Elite V) "Elite V" (Trader) Ship & Suit Decal
)

const (
	FederalNavyNone                RankFederalNavy = iota // None
	FederalNavyRecruit                                    // None
	FederalNavyCadet                                      // None
	FederalNavyMidshipman                                 // Federal Dropship, The Sarge invitation
	FederalNavyPettyOfficer                               // Sol
	FederalNavyChiefPettyOfficer                          // Federal Assault Ship, Vega, Beta Hydri
	FederalNavyWarrantOfficer                             // PLX 695
	FederalNavyEnsign                                     // Federal Gunship, Ross 128
	FederalNavyLieutenant                                 // Exbeur
	FederalNavyLieutenantCommander                        // None
	FederalNavyPostCommander                              // Hors
	FederalNavyPostCaptain                                // None
	FederalNavyRearAdmiral                                // Federal Corvette
	FederalNavyViceAdmiral                                // None
	FederalNavyAdmiral                                    // None
)

const (
	ImperialNavyNone     RankImperialNavy = iota // None
	ImperialNavyOutsider                         // Hera Tani invitation
	ImperialNavySerf                             // None
	ImperialNavyMaster                           // Imperial Courier
	ImperialNavySquire                           // Achenar
	ImperialNavyKnight                           // None
	ImperialNavyLord                             // None
	ImperialNavyBaron                            // Imperial Clipper, Summerland
	ImperialNavyViscount                         // None
	ImperialNavyCount                            // None
	ImperialNavyEarl                             // Facece
	ImperialNavyMarquis                          // None
	ImperialNavyDuke                             // Imperial Cutter
	ImperialNavyPrince                           // None
	ImperialNavyKing                             // None
)

var RankCombatToString = map[RankCombat]string{
	CombatHarmless:       "Harmless",
	CombatMostlyHarmless: "Mostly Harmless",
	CombatNovice:         "Novice",
	CombatCompetent:      "Competent",
	CombatExpert:         "Expert",
	CombatMaster:         "Master",
	CombatDangerous:      "Dangerous",
	CombatDeadly:         "Deadly",
	CombatElite:          "Elite",
	CombatEliteI:         "Elite I",
	CombatEliteII:        "Elite II",
	CombatEliteIII:       "Elite III",
	CombatEliteIV:        "Elite IV",
	CombatEliteV:         "Elite V",
}

var RankCQCToString = map[RankCQC]string{
	CQCHelpless:         "Helpless",
	CQCMostlyHelpless:   "Mostly Helpless",
	CQCAmateur:          "Amateur",
	CQCSemiProfessional: "Semi Professional",
	CQCProfessional:     "Professional",
	CQCChampion:         "Champion",
	CQCHero:             "Hero",
	CQCLegend:           "Legend",
	CQCElite:            "Elite",
	CQCEliteI:           "Elite I",
	CQCEliteII:          "Elite II",
	CQCEliteIII:         "Elite III",
	CQCEliteIV:          "Elite IV",
	CQCEliteV:           "Elite V",
}

var RankExobiologistToString = map[RankExobiologist]string{
	ExobiologistDirectionless:       "Directionless",
	ExobiologistMostlyDirectionless: "Mostly Directionless",
	ExobiologistCompiler:            "Compiler",
	ExobiologistCollector:           "Collector",
	ExobiologistCataloguer:          "Cataloguer",
	ExobiologistTaxonomist:          "Taxonomist",
	ExobiologistEcologist:           "Ecologist",
	ExobiologistGeneticist:          "Geneticist",
	ExobiologistElite:               "Elite",
	ExobiologistEliteI:              "Elite I",
	ExobiologistEliteII:             "Elite II",
	ExobiologistEliteIII:            "Elite III",
	ExobiologistEliteIV:             "Elite IV",
	ExobiologistEliteV:              "Elite V",
}

var RankExplorerToString = map[RankExplorer]string{
	ExplorerAimless:       "Aimless",
	ExplorerMostlyAimless: "Mostly Aimless",
	ExplorerScout:         "Scout",
	ExplorerSurveyor:      "Surveyor",
	ExplorerTrailblazer:   "Trailblazer",
	ExplorerPathfinder:    "Pathfinder",
	ExplorerRanger:        "Ranger",
	ExplorerPioneer:       "Pioneer",
	ExplorerElite:         "Elite",
	ExplorerEliteI:        "Elite I",
	ExplorerEliteII:       "Elite II",
	ExplorerEliteIII:      "Elite III",
	ExplorerEliteIV:       "Elite IV",
	ExplorerEliteV:        "Elite V",
}

var RankMercenaryToString = map[RankMercenary]string{
	MercenaryDefenceless:       "Defenceless",
	MercenaryMostlyDefenceless: "Mostly Defenceless",
	MercenaryRookie:            "Rookie",
	MercenarySoldier:           "Soldier",
	MercenaryGunslinger:        "Gunslinger",
	MercenaryWarrior:           "Warrior",
	MercenaryGladiator:         "Gladiator",
	MercenaryDeadeye:           "Deadeye",
	MercenaryElite:             "Elite",
	MercenaryEliteI:            "Elite I",
	MercenaryEliteII:           "Elite II",
	MercenaryEliteIII:          "Elite III",
	MercenaryEliteIV:           "Elite IV",
	MercenaryEliteV:            "Elite V",
}

var RankTraderToString = map[RankTrader]string{
	TraderPenniless:       "Penniless",
	TraderMostlyPenniless: "Mostly Penniless",
	TraderPeddler:         "Peddler",
	TraderDealer:          "Dealer",
	TraderMerchant:        "Merchant",
	TraderBroker:          "Broker",
	TraderEntrepreneur:    "Entrepreneur",
	TraderTycoon:          "Tycoon",
	TraderElite:           "Elite",
	TraderEliteI:          "Elite I",
	TraderEliteII:         "Elite II",
	TraderEliteIII:        "Elite III",
	TraderEliteIV:         "Elite IV",
	TraderEliteV:          "Elite V",
}

var RankFederalNavyToString = map[RankFederalNavy]string{
	FederalNavyNone: "None",

	FederalNavyRecruit:             "Recruit",
	FederalNavyCadet:               "Cadet",
	FederalNavyMidshipman:          "Midshipman",
	FederalNavyPettyOfficer:        "Petty Officer",
	FederalNavyChiefPettyOfficer:   "Chief Petty Officer",
	FederalNavyWarrantOfficer:      "Warrant Officer",
	FederalNavyEnsign:              "Ensign",
	FederalNavyLieutenant:          "Lieutenant",
	FederalNavyLieutenantCommander: "Lieutenant Commander",
	FederalNavyPostCommander:       "Post Commander",
	FederalNavyPostCaptain:         "Post Captain",
	FederalNavyRearAdmiral:         "Rear Admiral",
	FederalNavyViceAdmiral:         "Vice Admiral",
	FederalNavyAdmiral:             "Admiral",
}

var RankImperialNavyToString = map[RankImperialNavy]string{
	ImperialNavyNone: "None",

	ImperialNavyOutsider: "Outsider",
	ImperialNavySerf:     "Serf",
	ImperialNavyMaster:   "Master",
	ImperialNavySquire:   "Squire",
	ImperialNavyKnight:   "Knight",
	ImperialNavyLord:     "Lord",
	ImperialNavyBaron:    "Baron",
	ImperialNavyViscount: "Viscount",
	ImperialNavyCount:    "Count",
	ImperialNavyEarl:     "Earl",
	ImperialNavyMarquis:  "Marquis",
	ImperialNavyDuke:     "Duke",
	ImperialNavyPrince:   "Prince",
	ImperialNavyKing:     "King",
}
