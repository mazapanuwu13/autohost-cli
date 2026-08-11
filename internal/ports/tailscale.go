package ports

type Tailscale interface {
	Installed() bool
	Install() error
	Login() error
	LoginHeadscale(loginServer, authKey string) error
	Logout() error
	Down() error
	IP() (string, error)
	GetMachineName() (string, error)
}

type SplitDNS interface {
	Ensure(domain string, nameservers []string) error
}
