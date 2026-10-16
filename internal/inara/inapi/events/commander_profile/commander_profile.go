package commander_profile

import (
	"bytes"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gethiox/EDCITR/internal/inara/inapi"
	"github.com/gethiox/EDCITR/internal/inara/inapi/events"
)

const NoProfilePictureSuffix = "dummyprofile.jpg"

type CommanderRankPilot struct {
	RankName     string  `json:"rankName"`
	RankValue    string  `json:"rankValue"`    // number as string, starting from 0
	RankProgress float64 `json:"rankProgress"` // 0.0 - 1.0
}

type CommanderMainShip struct {
	ShipType  string `json:"shipType"`
	ShipName  string `json:"shipName"`
	ShipIdent string `json:"shipIdent"`
	ShipRole  string `json:"shipRole"`
}

type CommanderSquadron struct {
	SquadronID           uint64 `json:"squadronID"`
	SquadronName         string `json:"squadronName"`
	SquadronMembersCount uint64 `json:"squadronMembersCount"`
	SquadronMemberRank   string `json:"squadronMemberRank"`
	InaraURL             string `json:"inaraURL"`
}

type CommanderWing struct {
	WingID           uint64 `json:"wingID"`
	WingName         string `json:"wingName"`
	WingMembersCount uint64 `json:"wingMembersCount"`
	WingMemberRank   string `json:"wingMemberRank"`
	InaraURL         string `json:"inaraURL"`
}

type CommanderProfileInfo struct {
	UserID                  uint64               `json:"userID"`
	UserName                string               `json:"userName"`
	CommanderRanksPilot     []CommanderRankPilot `json:"commanderRanksPilot,omitempty"`
	PreferredAllegianceName string               `json:"preferredAllegianceName"`
	PreferredPowerName      string               `json:"preferredPowerName,omitempty"`
	CommanderMainShip       CommanderMainShip    `json:"commanderMainShip,omitempty"`
	CommanderSquadron       CommanderSquadron    `json:"commanderSquadron,omitempty"`
	CommanderWing           CommanderWing        `json:"commanderWing,omitempty"`
	PreferredGameRole       string               `json:"preferredGameRole,omitempty"`
	AvatarImageURL          string               `json:"avatarImageURL"`
	InaraURL                string               `json:"inaraURL"`
	CommanderName           string               `json:"commanderName,omitempty"`
	OtherNamesFound         []string             `json:"otherNamesFound,omitempty"`
}

type EventResponse struct {
	EventStatus     int                  `json:"eventStatus"` // 202 when multiple found, 200 when exact match found, 204 when no results found
	EventData       CommanderProfileInfo `json:"eventData,omitempty"`
	EventStatusText string               `json:"eventStatusText,omitempty"` // "Multiple results found.", "No results found."
}

type CommanderProfileResponse struct {
	Header struct {
		EventData struct {
			UserID   int    `json:"userID"`
			UserName string `json:"userName"`
		} `json:"eventData"`
		EventStatus int `json:"eventStatus"`
	} `json:"header"`
	Events []EventResponse `json:"events"`
}

func GetCommanderProfile(client *http.Client, header inapi.APIHeader, commanderNames []string) (CommanderProfileResponse, error) {
	type EventData struct {
		SearchName string `json:"searchName"`
	}

	now := time.Now()
	body := inapi_events.InaraAPIRequest{}
	body.Header = header

	body.Events = make([]inapi_events.InaraAPIEvent, 0, len(commanderNames))
	for _, commanderName := range commanderNames {
		body.Events = append(body.Events, inapi_events.InaraAPIEvent{
			EventName:      "getCommanderProfile",
			EventTimestamp: now,
			EventData:      EventData{SearchName: commanderName},
		})
	}

	bodyJson, err := json.Marshal(&body)
	if err != nil {
		return CommanderProfileResponse{}, fmt.Errorf("prepare json body failed: %w", err)
	}
	bodyJsonBuffer := bytes.NewBuffer(bodyJson)

	req, err := http.NewRequest(http.MethodPost, inapi_events.INAPI_URL, bodyJsonBuffer)
	if err != nil {
		return CommanderProfileResponse{}, fmt.Errorf("prepare request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if header.IsBeingDeveloped {
		req.Header.Set("User-Agent", fmt.Sprintf("%s/%s-dev", header.AppName, header.AppVersion))
	} else {
		req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", header.AppName, header.AppVersion))
	}

	resp, err := client.Do(req)
	if err != nil {
		return CommanderProfileResponse{}, fmt.Errorf("performing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CommanderProfileResponse{}, fmt.Errorf("response status code: %d, expected: %d", resp.StatusCode, http.StatusOK)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return CommanderProfileResponse{}, fmt.Errorf("reading response body failed: %w", err)
	}

	var respStruct CommanderProfileResponse

	err = json.Unmarshal(data, &respStruct)
	if err != nil {
		return CommanderProfileResponse{}, fmt.Errorf("parsing body json failed: %w", err)
	}

	return respStruct, nil
}
