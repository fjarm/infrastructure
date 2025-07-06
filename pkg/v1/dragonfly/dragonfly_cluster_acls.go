package dragonfly

import (
	"bytes"
	"fmt"
	"text/template"
)

// ErrTemplateParsingError indicates that there was an issue parsing the acls template that resulted in an empty result
// or a result containing `<nil>`.
var ErrTemplateParsingError = fmt.Errorf("error parsing ACL file template")

const acls = `
user default on nopass -@ALL +PING +AUTH

{{ range $index, $user := .Users }}
user {{ $user.Username }} on >{{ $user.Password }} {{ range $cmd := $user.EnabledCommands }}{{ $cmd }} {{ end }}
{{ end }}
`

// dragonflyClusterUser defines a user in the DragonflyDB ACL file.
type dragonflyClusterUser struct {
	Username        string
	Password        string
	EnabledCommands []string
}

func parseDragonflyClusterUserToTemplate(
	dragonflyClusterUsers ...dragonflyClusterUser,
) (string, error) {
	aclTemplate, err := template.New("dragonfly.acl").Parse(acls)
	if err != nil {
		return "", err
	}

	templateData := struct {
		Users []dragonflyClusterUser
	}{
		Users: dragonflyClusterUsers,
	}
	templateDestination := bytes.NewBuffer(nil)
	err = aclTemplate.Execute(templateDestination, &templateData)
	if err != nil {
		return "", err
	}

	contents := templateDestination.String()
	if contents == "" || contents == "<nil>" {
		return "", ErrTemplateParsingError
	}

	return contents, nil
}
