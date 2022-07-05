package auth

// Config holds authentication configuration
type Config struct {
	Algo               string   `json:"algo"`
	Secret             string   `json:"secret"`
	SignKey            string   `json:"signKey"`
	Timeout            int64    `json:"timeout"`
	RefreshBefore      int64    `json:"refreshBefore"`
	HTTPOnly           bool     `json:"httpOnly"`
	Secure             bool     `json:"secure"`
	OwnershipBypassIPs []string `json:"ownershipBypassIPs,omitempty"`
	DeviceIDCheck      bool     `json:"deviceIDCheck"`
	IPCheck            bool     `json:"ipCheck"`
}
