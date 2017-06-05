package client

import (
	"github.com/azrle/go-nss-cache/account"
)

// Client will fetch data from data source
type Client interface {
	// Fetch user accounts
	GetUsers() (account.Users, error)
	// Fetch group accounts
	GetGroups() (account.Groups, error)
	// Fetch SSH keys
	GetSSHKeys() (account.SSHKeys, error)
}
