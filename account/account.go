package account

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	PASSWD_CACHE     = "/etc/passwd.cache"
	GROUP_CACHE      = "/etc/group.cache"
	SSHKEY_CACHE     = "/etc/sshkey.cache"
	CACHE_PERMISSION = 0644
)

// NSS passwd db
type User struct {
	Name     string
	Password string
	Uid      string
	Gid      string
	Gecos    string
	HomeDir  string
	Shell    string
}

// NSS group db
type Group struct {
	Name     string
	Password string
	Gid      string
	Members  []string
}

// SSH Keys
type SSHKey struct {
	UserName string
	Pubkey   string
}

type Users []*User
type Groups []*Group
type SSHKeys []*SSHKey

// Linux account interface
type Account interface {
	// Update NSS db cache file
	UpdateNSSCache() error
}

// Update passwd cache file
func (u Users) UpdateNSSCache() error {
	var content string
	for _, user := range u {
		line := fmt.Sprintf(
			"%s:%s:%s:%s:%s:%s:%s\n",
			user.Name, user.Password, user.Uid, user.Gid,
			user.Gecos, user.HomeDir, user.Shell,
		)
		content += line
	}
	return writeFile(PASSWD_CACHE, []byte(content), CACHE_PERMISSION)
}

// Update group cache file
func (g Groups) UpdateNSSCache() error {
	var content string
	for _, group := range g {
		line := fmt.Sprintf(
			"%s:%s:%s:%s\n",
			group.Name, group.Password,
			group.Gid, strings.Join(group.Members, ","),
		)
		content += line
	}
	return writeFile(GROUP_CACHE, []byte(content), CACHE_PERMISSION)
}

// Update SSHKey cache file
func (s SSHKeys) UpdateNSSCache() error {
	var content string
	for _, sshkey := range s {
		line := fmt.Sprintf("%s:%s\n", sshkey.UserName, sshkey.Pubkey)
		content += line
	}
	return writeFile(SSHKEY_CACHE, []byte(content), CACHE_PERMISSION)
}

// Write content to filepath atomically with permission perm
func writeFile(filepath string, content []byte, perm os.FileMode) error {
	directory, filename := path.Split(filepath)
	tmpfile, err := ioutil.TempFile(directory, filename+"-")
	if err != nil {
		return err
	}

	for {
		err = os.Chmod(tmpfile.Name(), perm)
		if err != nil {
			break
		}
		_, err = tmpfile.Write(content)
		if err != nil {
			break
		}
		err = tmpfile.Sync()
		if err != nil {
			break
		}
		err = tmpfile.Close()
		if err != nil {
			break
		}
		err = os.Rename(tmpfile.Name(), filepath)
		break
	}

	if err != nil {
		os.Remove(tmpfile.Name())
	}
	return err
}
