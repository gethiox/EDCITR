package commander_profile

import "fmt"

func (s CommanderMainShip) String() string {
	name := s.ShipName
	if name == "" {
		name = "No Name"
	}
	ident := s.ShipIdent
	if ident == "" {
		ident = "No Ident"
	}

	shipType := ShipTypes[s.ShipType]
	if shipType == "" {
		shipType = "Unknown Type"
	}

	return fmt.Sprintf("%s [%s] (%s)", name, ident, shipType)
}

var ShipTypes = map[string]string{
	"adder":                    "Adder",
	"typex_3":                  "Alliance Challenger",
	"typex":                    "Alliance Chieftain",
	"typex_2":                  "Alliance Crusader",
	"anaconda":                 "Anaconda",
	"asp":                      "Asp Explorer",
	"asp_scout":                "Asp Scout",
	"belugaliner":              "Beluga Liner",
	"explorer_nx":              "Caspian Explorer",
	"cobramkiii":               "Cobra Mk III",
	"cobramkiv":                "Cobra Mk IV",
	"cobramkv":                 "Cobra Mk V",
	"corsair":                  "Corsair",
	"diamondbackxl":            "Diamondback Explorer",
	"diamondback":              "Diamondback Scout",
	"dolphin":                  "Dolphin",
	"eagle":                    "Eagle",
	"federation_dropship_mkii": "Federal Assault Ship",
	"federation_corvette":      "Federal Corvette",
	"federation_dropship":      "Federal Dropship",
	"federation_gunship":       "Federal Gunship",
	"ferdelance":               "Fer-de-Lance",
	"hauler":                   "Hauler",
	"empire_trader":            "Imperial Clipper",
	"empire_courier":           "Imperial Courier",
	"cutter":                   "Imperial Cutter",
	"empire_eagle":             "Imperial Eagle",
	"independant_trader":       "Keelback",
	"krait_mkii":               "Krait Mk II",
	"krait_light":              "Krait Phantom",
	"mamba":                    "Mamba",
	"mandalay":                 "Mandalay",
	"orca":                     "Orca",
	"panthermkii":              "Panther Clipper Mk II",
	"python":                   "Python",
	"python_nx":                "Python Mk II",
	"sidewinder":               "Sidewinder",
	"type6":                    "Type-6 Transporter",
	"type7":                    "Type-7 Transporter",
	"type8":                    "Type-8 Transporter",
	"type9":                    "Type-9 Heavy",
	"type9_military":           "Type-10 Defender",
	"lakonminer":               "Type-11 Prospector",
	"viper":                    "Viper",
	"viper_mkiv":               "Viper Mk IV",
	"vulture":                  "Vulture",
}
