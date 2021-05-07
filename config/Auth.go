package config

// Auth holds authentication configuration
type Auth struct {
	Algo     string `json:"algo"`
	Secret   string `json:"secret"`
	SignKey  string `json:"signKey"`
	Timeout  int64  `json:"timeout"`
	HTTPOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
}
