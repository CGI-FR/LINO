package localstorage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

const (
	// Version of the YAML strcuture.
	Version string = "v1"
	// Path of the credentials storage file relative to the current user's home directory.
	FilePath string = ".lino"
	// Name of the credentials storage file.
	FileName string = "credentials.yaml"
)

type YAMLCredentialsStore struct {
	Version         string            `yaml:"version"`
	CredentialsList []YAMLCredentials `yaml:"credentials,omitempty"`
}

type YAMLCredentials struct {
	ServerURL string `yaml:"serverURL"`
	Username  string `yaml:"username"`
	Secret    string `yaml:"secret"`
}

// Store credentials in a local file.
type YAMLStorage struct{}

func Store(creds *credentials.Credentials) error {
	h := YAMLStorage{}
	return h.Add(creds)
}

func Read(serverURL string) (*credentials.Credentials, error) {
	h := YAMLStorage{}
	username, secret, err := h.Get(serverURL)
	if err != nil {
		return nil, err
	}
	return &credentials.Credentials{serverURL, username, secret}, nil
}

// Add adds new credentials to the storage.
func (h YAMLStorage) Add(creds *credentials.Credentials) error {
	store, err := readFile()
	if err != nil {
		return err
	}

	yml := YAMLCredentials{
		ServerURL: creds.ServerURL,
		Username:  creds.Username,
		Secret:    creds.Secret,
	}

	added := false
	newList := []YAMLCredentials{}
	for _, credential := range store.CredentialsList {
		if credential.ServerURL != creds.ServerURL {
			newList = append(newList, credential)
		} else {
			newList = append(newList, yml)
			added = true
		}
	}
	if !added {
		newList = append(newList, yml)
	}
	store.CredentialsList = newList

	err = writeFile(store)
	if err != nil {
		return err
	}

	return nil
}

// Delete removes credentials from storage.
func (h YAMLStorage) Delete(serverURL string) error {
	store, err := readFile()
	if err != nil {
		return err
	}

	newList := []YAMLCredentials{}
	for _, credential := range store.CredentialsList {
		if credential.ServerURL != serverURL {
			newList = append(newList, credential)
		}
	}

	store.CredentialsList = newList

	err = writeFile(store)
	if err != nil {
		return err
	}

	return nil
}

// Get returns the username and secret to use for a given registry server URL.
func (h YAMLStorage) Get(serverURL string) (string, string, error) {
	store, err := readFile()
	if err != nil {
		return "", "", err
	}

	for _, credential := range store.CredentialsList {
		if credential.ServerURL == serverURL {
			return credential.Username, credential.Secret, nil
		}
	}
	return "", "", nil
}

// List returns the stored URLs and corresponding usernames.
func (h YAMLStorage) List() (map[string]string, error) {
	store, err := readFile()
	if err != nil {
		return nil, err
	}

	result := map[string]string{}
	for _, credential := range store.CredentialsList {
		result[credential.ServerURL] = credential.Username
	}

	return result, nil
}

func readFile() (*YAMLCredentialsStore, error) {
	store := &YAMLCredentialsStore{
		Version: Version,
	}

	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	storeFile := path.Join(home, FilePath, FileName)

	if _, err := os.Stat(storeFile); os.IsNotExist(err) {
		return store, nil
	}

	dat, err := ioutil.ReadFile(storeFile)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(dat, store)
	if err != nil {
		return nil, err
	}

	if store.Version != Version {
		return nil, fmt.Errorf("%w : invalid storage version (%s)", ErrInvalidStorage, store.Version)
	}

	return store, nil
}

func writeFile(list *YAMLCredentialsStore) error {
	out, err := yaml.Marshal(list)
	if err != nil {
		return err
	}

	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	storeDir := path.Join(home, FilePath)
	if _, err2 := os.Stat(storeDir); os.IsNotExist(err2) {
		err = os.MkdirAll(storeDir, 0700)
		if err != nil {
			return err
		}
	} else if err2 != nil {
		return err2
	}

	storeFile := path.Join(storeDir, FileName)

	err = ioutil.WriteFile(storeFile, out, 0600)
	if err != nil {
		return err
	}

	return nil
}
