package elite_status

import (
	"io"
	"net/http"
	"time"

	"github.com/gethiox/EDCITR/internal/jsonutil"
)

const EndpointProductionServer = "https://ed-server-status.orerve.net/"

type ServerStatus struct {
	EndpointData    EndpointFormat
	Timestamp       time.Time
	RequestDuration time.Duration
}

type StatusCode int

const (
	CodeGood        StatusCode = 1
	CodeMaintenance StatusCode = 0
	CodeDown        StatusCode = -1
)

type EndpointFormat struct {
	Code    StatusCode `json:"code,required"`
	Status  string     `json:"status,required"`  // example: 'Good', 'Maintenance', 'Down'
	Message string     `json:"message,required"` // Usually empty, example: 'Elite is currently updating and will be back online soon.'
	Product string     `json:"product,required"` // example: 'FORC-FDEV-DO-1000'
}

type EliteStatus struct {
	endpoint string
	client   *http.Client
}

func GetProductionClient(timeout time.Duration) EliteStatus {
	return EliteStatus{
		endpoint: EndpointProductionServer,
		client:   &http.Client{Timeout: timeout},
	}
}

func (s *EliteStatus) GetEDServerStatus() (ServerStatus, error) {
	var status ServerStatus

	req, err := http.NewRequest(http.MethodGet, s.endpoint, nil)
	if err != nil {
		status.Timestamp = time.Now()
		return status, err
	}

	req.Header.Set("User-Agent", "EDCITR/0.0.0") // TODO

	start := time.Now()
	resp, err := s.client.Do(req)
	end := time.Now()

	status.Timestamp = start
	status.RequestDuration = end.Sub(start)

	if err != nil {
		return status, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return status, err
	}

	err = jsonutil.UnmarshalStrictV3(data, &status.EndpointData)
	if err != nil {
		return status, err
	}

	return status, nil
}
