package types

type WgPeer struct {
	PublicKey  string `json:"public_key"`
	AllowedIPs string `json:"allowed_ips"`
	EndPoint   string `json:"endpoint"`
}

type WireguardConfig struct {
	PrivateKey string   `json:"private_key"`
	Address    string   `json:"address"`
	Peers      []WgPeer `json:"peers"`
}
