package bottlerocket

import (
	"bytes"
	"text/template"

	etcdbootstrapv1 "github.com/aws/etcdadm-bootstrap-provider/api/v1beta1"
	"github.com/aws/etcdadm-bootstrap-provider/pkg/userdata"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

const (
	cloudConfigHeader = `## template: jinja
#cloud-config
`
	// sentinelFileCommand writes a file to /run/cluster-api to signal successful Kubernetes bootstrapping in a way that
	// works both for Linux and Windows OS.
	sentinelFileCommand = "echo success > /run/cluster-api/bootstrap-success.complete"
)

var defaultTemplateFuncMap = template.FuncMap{
	"Indent": userdata.TemplateYAMLIndent,
}

func generateUserData(kind string, tpl string, data interface{}, input *userdata.BaseUserData, config etcdbootstrapv1.EtcdadmConfigSpec, log logr.Logger) ([]byte, error) {
	bootstrapContainerUserData, err := generateBootstrapContainerUserData(kind, tpl, data)
	if err != nil {
		return nil, err
	}

	return generateBottlerocketNodeUserData(bootstrapContainerUserData, input.Users, input.RegistryMirrorCredentials, input.Hostname, config, log)
}

func generateBootstrapContainerUserData(kind string, tpl string, data interface{}) ([]byte, error) {
	tm := template.New(kind).Funcs(defaultTemplateFuncMap)
	if _, err := tm.Parse(filesTemplate); err != nil {
		return nil, errors.Wrap(err, "failed to parse files template")
	}

	t, err := tm.Parse(tpl)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s template", kind)
	}

	var out bytes.Buffer
	if err := t.Execute(&out, data); err != nil {
		return nil, errors.Wrapf(err, "failed to generate %s template", kind)
	}

	return out.Bytes(), nil
}
