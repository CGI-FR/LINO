package urlbuilder

import "github.com/docker/docker-credential-helpers/client"

func defaultCredentialsStore() client.ProgramFunc {
	return client.NewShellProgramFunc("docker-credential-wincred")
}
