package auth

// Config holds authentication configuration
type Config struct {
	Algo               string   `json:"algo"`
	Secret             string   `json:"secret"`
	SignKey            string   `json:"signKey"`
	Timeout            int64    `json:"timeout"`
	HTTPOnly           bool     `json:"httpOnly"`
	Secure             bool     `json:"secure"`
	OwnershipBypassIPs []string `json:"ownershipBypassIPs,omitempty"`
	DeviceIDCheck      bool     `json:"deviceIDCheck"`
}
