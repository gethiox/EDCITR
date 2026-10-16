package inapi

type APIHeader struct {
	// Name of the application used
	AppName string `json:"appName"`
	// Version of the application (like: '1.2.4', please keep this standard versioning format as there may be a version control in effect)
	AppVersion string `json:"appVersion"`

	// If you are developing the application and/or testing new updates,
	// please set this as 'true'. This setting will skip the updates for global events like setCommunityGoal,
	// which may otherwise cause the problems with your test data. Please note that newly accepted applications
	// in development has this set by default on server side, but don't rely on it and set it on your side, too.
	IsBeingDeveloped bool `json:"isBeingDeveloped"`

	// User's personal API key (set by user) or generic application key (for general read-only events)
	APIkey string `json:"APIkey"`

	// In-game CMDR name of user (not set by user, get this from journals or cAPI to ensure it is a correct in-game name to avoid future problems).
	// It is recommended to be always set when no generic API key is used (otherwise some events may not work).
	CommanderName string `json:"commanderName"`

	// Commander's unique Frontier ID (is provided by journals since 3.3) in a format: 'F123456' or '123456'.
	// When not known, set nothing.
	CommanderFrontierID string `json:"commanderFrontierID"`
}
