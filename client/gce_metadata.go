package client

import (
	"cloud.google.com/go/compute/metadata"
	"github.com/azrle/go-nss-cache/account"
	"strings"
)

type GCEMetadataConf struct {
	KeyPrefix   string
	UserKey     string
	GroupKey    string
	PrimaryGid  string
	HomeDirBase string
}

type gceMetadataClient struct {
	config GCEMetadataConf
}

// Create a new gceMetadataClient
// The client will make request to the metadata server
func NewGCEMetadatClient(conf GCEMetadataConf) Client {
	return &gceMetadataClient{
		config: conf,
	}
}

// Fetch users from metadata server.
//
// The format should be like this:
//   [USERNAME]:[UID]:[PUBKEYS]
func (c *gceMetadataClient) GetUsers() (account.Users, error) {
	val, err := metadata.Get(c.config.KeyPrefix + c.config.UserKey)
	if err != nil {
		return nil, err
	}

	var users account.Users
	lines := strings.Split(val, "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		user := &account.User{
			Name:     fields[0],
			Password: "*",
			Uid:      fields[1],
			Gid:      c.config.PrimaryGid,
			Gecos:    fields[0],
			HomeDir:  c.config.HomeDirBase + "/" + fields[0],
			Shell:    "/bin/bash",
		}
		users = append(users, user)
	}
	return users, nil
}

// Fetch groups from metadata server.
//
// The format should be like this:
//   [GROUPNAME]:[GID]:[MEMBERS]
func (c *gceMetadataClient) GetGroups() (account.Groups, error) {
	val, err := metadata.Get(c.config.KeyPrefix + c.config.GroupKey)
	if err != nil {
		return nil, err
	}

	var groups account.Groups
	lines := strings.Split(val, "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 3 {
			continue
		}
		group := &account.Group{
			Name:     fields[0],
			Password: "*",
			Gid:      fields[1],
			Members:  strings.Split(fields[2], ","),
		}
		groups = append(groups, group)
	}
	return groups, nil
}

// Fetch public keys from metadata server which
// are stored along with users.
//
// The format should be like this:
//   [USERNAME]:[UID]:[PUBKEYS]
func (c *gceMetadataClient) GetSSHKeys() (account.SSHKeys, error) {
	val, err := metadata.Get(c.config.KeyPrefix + c.config.UserKey)
	if err != nil {
		return nil, err
	}

	var pubkeys account.SSHKeys
	lines := strings.Split(val, "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		// TODO: ACL by username
		if len(fields) < 3 {
			continue
		}
		pubkey := &account.SSHKey{
			UserName: fields[0],
			Pubkey:   fields[2],
		}
		pubkeys = append(pubkeys, pubkey)
	}
	return pubkeys, nil
}
