package v1

type IPTable struct {
	IP            string `json:"ip"`
	SSHPort       string `json:"sshPort"`
	Localhost     bool   `json:"localhost"`
	SSHReachable  bool   `json:"sshReachable"`
	SSHAuthorized bool   `json:"sshAuthorized"`
	Added         bool   `json:"added"`
}
