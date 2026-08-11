package tailscale

import "os/exec"

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Installed() bool     { _, err := exec.LookPath("tailscale"); return err == nil }
func (a *Adapter) Install() error      { return InstallTailscale() }
func (a *Adapter) Login() error        { return LoginTailscale() }
func (a *Adapter) LoginHeadscale(loginServer, authKey string) error {
	return LoginHeadscale(loginServer, authKey)
}
func (a *Adapter) Logout() error       { return LogoutTailscale() }
func (a *Adapter) Down() error         { return DownTailscale() }
func (a *Adapter) IP() (string, error) { return TailscaleIP() }
func (a *Adapter) GetMachineName() (string, error) {
	return GetMachineName()
}
