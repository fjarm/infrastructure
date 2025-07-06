package valkey

import (
	"bytes"
	"fmt"
	"text/template"
)

// ErrTemplateParsingError indicates that there was an issue parsing the acls template that resulted in an empty result
// or a result containing `<nil>`.
var ErrTemplateParsingError = fmt.Errorf("error parsing ACL file template")

const configTemplate = `
# SEE: https://valkey.io/topics/acl/
user default on >{{ .DefaultUserCredentials }} allchannels +multi +slaveof +ping +exec +subscribe +config|rewrite +role +publish +info +client|setname +client|kill +script|kill
user sentinel-user on >{{ .SentinelUserCredentials }} allchannels +multi +slaveof +ping +exec +subscribe +config|rewrite +role +publish +info +client|setname +client|kill +script|kill
user replica-user on >{{ .ReplicaUserCredentials }} +psync +replconf +ping

{{ range $index, $user := .Users }}
user {{ $user.Username }} on >{{ $user.Password }} -@ALL {{ range $cmd := $user.EnabledCommands }}{{ $cmd }} {{ end }}
{{ end }}

# Disable AOF https://valkey.io/docs/topics/persistence.html
appendonly no

# Enable RDB persistence, AOF persistence is disabled.
# Unless specified otherwise, by default the server will save the DB:
#   * After 3600 seconds (an hour) if at least 1 change was performed
#   * After 300 seconds (5 minutes) if at least 100 changes were performed
#   * After 60 seconds if at least 10000 changes were performed
save 3600 1 300 100 60 10000
`

type valkeyConfig struct {
	DefaultUserCredentials  string
	SentinelUserCredentials string
	ReplicaUserCredentials  string
	Users                   []*valkeyUser
}

// valkeyUser describes a user in the Valkey ACL file and their allowed commands. All users start with -@ALL by default.
type valkeyUser struct {
	Username        string
	Password        string
	EnabledCommands []string
}

// newValkeyCommonConfig uses text templating to compose the contents of an ACL file as a string.
func newValkeyCommonConfig(
	cfg *valkeyConfig,
) (string, error) {
	tmp, err := template.New("valkey.conf").Parse(configTemplate)
	if err != nil {
		return "", err
	}

	templateDestination := bytes.NewBuffer(nil)
	err = tmp.Execute(templateDestination, cfg)
	if err != nil {
		return "", err
	}

	contents := templateDestination.String()
	if contents == "" || contents == "<nil>" {
		return "", ErrTemplateParsingError
	}

	return contents, nil
}
