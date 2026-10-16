package message

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecoding(t *testing.T) {
	type TestCase struct {
		input    string
		expected DecodedMessage
	}

	for i, tc := range []TestCase{
		{
			input: `$CruiseLiner_SCPatrol01;`,
			expected: DecodedMessage{
				What:       "CruiseLiner",
				SubWhat:    "SCPatrol",
				Variant:    1,
				Parameters: map[string]string{},
			},
		},
		{
			input: `$Pirate_ReminderUnits01:#uints=12;`,
			expected: DecodedMessage{
				What:    "Pirate",
				SubWhat: "ReminderUnits",
				Variant: 1,
				Parameters: map[string]string{
					"uints": "12",
				},
			},
		},
		{
			input: `$Pirate_ThreatenTimer03:#seconds=15;`,
			expected: DecodedMessage{
				What:    "Pirate",
				SubWhat: "ThreatenTimer",
				Variant: 3,
				Parameters: map[string]string{
					"seconds": "15",
				},
			},
		},
		{
			input: `$Pirate_ThreatenSpecific01:#units=1:#CommodityName=$thargoidtitandrivecomponent_Name;;`,
			expected: DecodedMessage{
				What:    "Pirate",
				SubWhat: "ThreatenSpecific",
				Variant: 1,
				Parameters: map[string]string{
					"units":         "1",
					"CommodityName": "$thargoidtitandrivecomponent_Name",
				},
			},
		},
		{
			input: `$Pirate_Trader_OnEnemyShipDetection05;`,
			expected: DecodedMessage{
				What:       "Pirate",
				SubWhat:    "Trader_OnEnemyShipDetection",
				Variant:    5,
				Parameters: map[string]string{},
			},
		},
		{
			input: `$Pirate_HunterHostileSC_Relevant04;`,
			expected: DecodedMessage{
				What:       "Pirate",
				SubWhat:    "HunterHostileSC_Relevant",
				Variant:    4,
				Parameters: map[string]string{},
			},
		},
		{
			input: `$Mission_Collect_MessengerChat1:#CommodityName=$Coffee_Name;:#CommodityQuantity=72:#destinationStationName=Dewsnap Prospecting Station:#destinationStationSystemName=Sol:#reward=817628.000000:#startStationSystemName=Sol:#altDestinationStationName=Dewsnap Prospecting Station:#altDestinationStationSystemName=Sol:#CommodityValue=102600.000000:#missionGiverFactionContact=Kurt Preston:#missionGiverFactionContactTitle=$GovernmentRankContactDemocracy_6_Male;:#commanderName=Gethiox;`,
			expected: DecodedMessage{
				What:    "Mission",
				SubWhat: "Collect_MessengerChat",
				Variant: 1,
				Parameters: map[string]string{
					"CommodityName":                   "$Coffee_Name",
					"CommodityQuantity":               "72",
					"destinationStationName":          "Dewsnap Prospecting Station",
					"destinationStationSystemName":    "Sol",
					"reward":                          "817628.000000",
					"startStationSystemName":          "Sol",
					"altDestinationStationName":       "Dewsnap Prospecting Station",
					"altDestinationStationSystemName": "Sol",
					"CommodityValue":                  "102600.000000",
					"missionGiverFactionContact":      "Kurt Preston",
					"missionGiverFactionContactTitle": "$GovernmentRankContactDemocracy_6_Male",
					"commanderName":                   "Gethiox",
				},
			},
		},
		{
			// note: has duplicate key: "altDestinationStationSystemName"
			input: `$Mission_Assassinate_ContactChat1:#TargetType=$MissionUtil_FactionTag_PirateLord;:#AltTargetType=$MissionUtil_FactionTag_Informant;:#minorFaction=Workers of K Camelopardalis Resistance:#factionToKill=Ehlangai Family:#altFactionToKill=Workers of K Camelopardalis Resistance:#pilotName1=Jem Barnard:#pilotName2=Karl Hoefler:#pilotName3=greenrobot:#Reward=3066660.000000:#destinationStationName=Bulgarin Orbital:#destinationStationSystemName=Ehlangai:#startStationName=Hui Enterprise:#startSystemName=Catuntinigi:#startStationSystemName=Catuntinigi:#altDestinationStationName=Hui Enterprise:#altDestinationStationSystemName=Catuntinigi:#ContactSystem=LHS 2123:#NewDestinationStationSystemName=AC +67 1894:#altDestinationStationSystemName=Catuntinigi:#missionGiverFactionContact=Alec Mccarthy:#missionGiverFactionContactTitle=$GovernmentRankContactDemocracy_5_Male;:#commanderName=Gethiox;`,
			expected: DecodedMessage{
				What:    "Mission",
				SubWhat: "Assassinate_ContactChat",
				Variant: 1,
				Parameters: map[string]string{
					"TargetType":                      "$MissionUtil_FactionTag_PirateLord",
					"AltTargetType":                   "$MissionUtil_FactionTag_Informant",
					"minorFaction":                    "Workers of K Camelopardalis Resistance",
					"factionToKill":                   "Ehlangai Family",
					"altFactionToKill":                "Workers of K Camelopardalis Resistance",
					"pilotName1":                      "Jem Barnard",
					"pilotName2":                      "Karl Hoefler",
					"pilotName3":                      "greenrobot",
					"Reward":                          "3066660.000000",
					"destinationStationName":          "Bulgarin Orbital",
					"destinationStationSystemName":    "Ehlangai",
					"startStationName":                "Hui Enterprise",
					"startSystemName":                 "Catuntinigi",
					"startStationSystemName":          "Catuntinigi",
					"altDestinationStationName":       "Hui Enterprise",
					"altDestinationStationSystemName": "Catuntinigi",
					"ContactSystem":                   "LHS 2123",
					"NewDestinationStationSystemName": "AC +67 1894",
					//"altDestinationStationSystemName":"Catuntinigi",
					"missionGiverFactionContact":      "Alec Mccarthy",
					"missionGiverFactionContactTitle": "$GovernmentRankContactDemocracy_5_Male",
					"commanderName":                   "Gethiox",
				},
			},
		},
		{
			// note: underscore directly before variant number
			input: `$DistressCallAmbushedPilot_01;`,
			expected: DecodedMessage{
				What:       "DistressCallAmbushedPilot",
				SubWhat:    "",
				Variant:    1,
				Parameters: map[string]string{},
			},
		},
		{
			input: `$PowersSecurity_PassedStopAndSearchNoPower06:#myPowerName=Jerome Archer;`,
			expected: DecodedMessage{
				What:    "PowersSecurity",
				SubWhat: "PassedStopAndSearchNoPower",
				Variant: 6,
				Parameters: map[string]string{
					"myPowerName": "Jerome Archer",
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%d_%s", i, tc.expected.OriginalIdentifier()), func(t *testing.T) {
			decoded := DecodeMessage(tc.input)
			assert.Equal(t, tc.expected, decoded)
		})
	}
}
