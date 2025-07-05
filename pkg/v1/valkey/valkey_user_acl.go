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
{{ range $index, $user := .Users }}
user {{ $user.Username }} on >{{ $user.Password }} -@ALL {{ range $cmd := $user.EnabledCommands }}{{ $cmd }} {{ end }}
{{ end }}

# Enable AOF https://valkey.io/docs/topics/persistence.html
appendonly yes
# Disable RDB persistence, AOF persistence already enabled.
save ""
`

// valkeyUser describes a user in the Valkey ACL file and their allowed commands. All users start with -@ALL by default.
type valkeyUser struct {
	Username        string
	Password        string
	EnabledCommands []string
}

// newValkeyUserACL uses text templating to compose the contents of an ACL file as a string.
func newValkeyUserACL(
	users ...*valkeyUser,
) (string, error) {
	tmp, err := template.New("valkey.acl").Parse(configTemplate)
	if err != nil {
		return "", err
	}

	templateData := struct {
		Users []*valkeyUser
	}{
		Users: users,
	}
	templateDestination := bytes.NewBuffer(nil)
	err = tmp.Execute(templateDestination, &templateData)
	if err != nil {
		return "", err
	}

	contents := templateDestination.String()
	if contents == "" || contents == "<nil>" {
		return "", ErrTemplateParsingError
	}

	return contents, nil
}
