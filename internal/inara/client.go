package inara

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gethiox/EDCITR/internal/inara/inapi"
	"github.com/gethiox/EDCITR/internal/inara/inapi/events/commander_profile"
	"github.com/gethiox/EDCITR/internal/version"
)

var HeaderNotSetError = errors.New("header not set")

func GetClient(timeout time.Duration) *Client {
	return &Client{
		HttpClient: &http.Client{
			Timeout: timeout,
		},
		headerSet: false,
		readyChan: make(chan struct{}),
		closed:    false,
	}
}

type Client struct {
	HttpClient *http.Client
	header     inapi.APIHeader
	headerSet  bool
	readyChan  chan struct{}
	closed     bool
}

func (c *Client) Close() error {
	if !c.closed {
		c.closed = true
		if !c.headerSet {
			close(c.readyChan)
		}
	}
	return nil
}

func (c *Client) ReadyChan() chan struct{} {
	return c.readyChan
}

func (c *Client) Ready() bool {
	select {
	case <-c.readyChan:
		return true
	default:
		return false
	}
}

func (c *Client) GetCommanderProfile(cmdrList []string) (commander_profile.CommanderProfileResponse, error) {
	if !c.headerSet {
		return commander_profile.CommanderProfileResponse{}, HeaderNotSetError
	}
	return commander_profile.GetCommanderProfile(c.HttpClient, c.header, cmdrList)
}

func (c *Client) SetHeader(header inapi.APIHeader) error {
	switch {
	case c.closed:
		return errors.New("client is closed")
	case header.AppName == "":
		return errors.New("AppName is required")
	case header.AppVersion == "":
		return errors.New("AppVersion is required")
	case header.APIkey == "":
		return errors.New("APIKey is required")
	case header.CommanderName == "":
		return errors.New("CommanderName is required")
	default:
		c.header = header
		c.headerSet = true
		close(c.readyChan)
		return nil
	}
}

func (c *Client) TestClient() error {
	if !c.headerSet {
		return HeaderNotSetError
	}

	resp, err := commander_profile.GetCommanderProfile(c.HttpClient, c.header, []string{c.header.CommanderName})
	if err != nil {
		return err
	}

	if len(resp.Events) != 1 {
		return fmt.Errorf("expected 1 event, got %d", len(resp.Events))
	}

	if resp.Events[0].EventStatus != 200 {
		return fmt.Errorf("expected 200 event, got %d", resp.Events[0].EventStatus)
	}

	return nil
}

func NewHeader(appVersion version.Version, appName, apiKey, cmdrName, frontierID string) inapi.APIHeader {
	return inapi.APIHeader{
		AppName:             appName,
		AppVersion:          appVersion.FormatMMP(),
		IsBeingDeveloped:    appVersion.IsDev(),
		APIkey:              apiKey,
		CommanderName:       cmdrName,
		CommanderFrontierID: frontierID,
	}
}
