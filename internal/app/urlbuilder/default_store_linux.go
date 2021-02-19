package urlbuilder

import (
	"github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker-credential-helpers/pass"
)

func defaultCredentialsStore() client.ProgramFunc {
	p := pass.Pass{}
	if p.CheckInitialized() {
		return client.NewShellProgramFunc("docker-credential-pass")
	}

	return client.NewShellProgramFunc("docker-credential-secretservice")
}
