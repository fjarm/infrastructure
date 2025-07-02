package valkey

import (
	"bytes"
	"fmt"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"text/template"
)

// ErrTemplateParsingError indicates that there was an issue parsing the acls template that resulted in an empty result
// or a result containing `<nil>`.
var ErrTemplateParsingError = fmt.Errorf("error parsing ACL file template")

const (
	aclSecretName = "valkey-acl"
)

const aclTemplate = `
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

// deployValkeyUserACLSecret composes a string containing the contents of an ACL file. That string is then deployed as
// a secret for Valkey to use.
func deployValkeyUserACLSecret(
	ctx *pulumi.Context,
	namespace *corev1.Namespace,
	provider *kubernetes.Provider,
	deps []pulumi.Resource,
) (*corev1.Secret, string, error) {
	aclContent, err := newValkeyUserACL(
		// TODO(2025-06-30): Add support to the template for creating users with no password
		&valkeyUser{
			Username:        "test",
			Password:        "test",
			EnabledCommands: []string{"+AUTH", "+PING", "+GET", "+SET", "~*"},
		},
		&valkeyUser{
			Username:        "default",
			Password:        "password",
			EnabledCommands: []string{"+@ALL", "~*"},
		},
	)
	if err != nil {
		return nil, "", err
	}

	aclSecretArgs := newValkeyUserACLSecretArgs(
		namespace,
		aclContent,
	)
	aclSecret, err := corev1.NewSecret(
		ctx,
		aclSecretName,
		aclSecretArgs,
		pulumi.Provider(provider),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return nil, "", err
	}
	return aclSecret, aclContent, nil
}

// newValkeyUserACL uses text templating to compose the contents of an ACL file as a string.
func newValkeyUserACL(
	users ...*valkeyUser,
) (string, error) {
	tmp, err := template.New("valkey.acl").Parse(aclTemplate)
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

// newValkeyUserACLSecretArgs creates the Secret specs that contain the Valkey ACL file contents.
func newValkeyUserACLSecretArgs(
	namespace *corev1.Namespace,
	aclContent string,
) *corev1.SecretArgs {
	return &corev1.SecretArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Secret"),
		Type:       pulumi.String("Opaque"),
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(aclSecretName),
			Namespace: namespace.Metadata.Name(),
			Labels: pulumi.StringMap{
				"app": pulumi.String(clusterAppLabel),
			},
		},
		StringData: pulumi.StringMap{
			aclSecretName: pulumi.String(aclContent),
		},
	}
}
