// +build !windows,!darwin,!linux

package urlbuilder

import "github.com/docker/docker-credential-helpers/client"

func defaultCredentialsStore() client.ProgramFunc {
	return nil
}
