package tailscale

import (
	"time"
)

type Status struct {
	Version        string                `json:"Version"`
	Tun            bool                  `json:"TUN"`
	BackendState   string                `json:"BackendState"`
	AuthURL        string                `json:"AuthURL"`
	TailscaleIPs   []string              `json:"TailscaleIPs"`
	Self           PeerStatus            `json:"Self"`
	Health         []string              `json:"Health"`
	MagicDNSSuffix string                `json:"MagicDNSSuffix"`
	CurrentTailnet TailnetStatus         `json:"CurrentTailnet"`
	Peer           map[string]PeerStatus `json:"Peer"`
	User           map[string]User       `json:"User"`
	ClientVersion  ClientVersion         `json:"ClientVersion"`
}

type ClientVersion struct {
	RunningLatest        bool   `json:"RunningLatest,omitempty"`
	LatestVersion        string `json:"LatestVersion,omitempty"`
	UrgentSecurityUpdate bool   `json:"UrgentSecurityUpdate,omitempty"`
	Notify               bool   `json:"Notify,omitempty"`
	NotifyURL            string `json:"NotifyURL,omitempty"`
	NotifyText           string `json:"NotifyText,omitempty"`
}

type TailnetStatus struct {
	Name            string `json:"Name"`
	MagicDNSSuffix  string `json:"MagicDNSSuffix"`
	MagicDNSEnabled bool   `json:"MagicDNSEnabled"`
}

type PeerStatus struct {
	ID             string     `json:"ID"`
	PublicKey      string     `json:"PublicKey"`
	HostName       string     `json:"HostName"`
	DNSName        string     `json:"DNSName"`
	OS             string     `json:"OS"`
	UserID         int64      `json:"UserID"`
	TailscaleIPs   []string   `json:"TailscaleIPs"`
	AllowedIPs     []string   `json:"AllowedIPs"`
	Addrs          []string   `json:"Addrs"`
	CurAddr        string     `json:"CurAddr"`
	Relay          string     `json:"Relay"`
	RxBytes        int        `json:"RxBytes"`
	TxBytes        int        `json:"TxBytes"`
	Created        time.Time  `json:"Created"`
	LastWrite      time.Time  `json:"LastWrite"`
	LastSeen       time.Time  `json:"LastSeen"`
	LastHandshake  time.Time  `json:"LastHandshake"`
	Online         bool       `json:"Online"`
	ExitNode       bool       `json:"ExitNode"`
	ExitNodeOption bool       `json:"ExitNodeOption"`
	Active         bool       `json:"Active"`
	PeerAPIURL     []string   `json:"PeerAPIURL"`
	InNetworkMap   bool       `json:"InNetworkMap"`
	InMagicSock    bool       `json:"InMagicSock"`
	InEngine       bool       `json:"InEngine"`
	Expired        bool       `json:"Expired,omitempty"`
	KeyExpiry      *time.Time `json:"KeyExpiry,omitempty"`
}

type User struct {
	ID            int64  `json:"ID"`
	LoginName     string `json:"LoginName"`
	DisplayName   string `json:"DisplayName"`
	ProfilePicURL string `json:"ProfilePicURL"`
}
