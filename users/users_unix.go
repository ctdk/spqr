// +build linux darwin

// common functions shared across whatever unixes this might someday support

package users

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"log"
)

const sshDirPerm = 0700
const authKeyPerm = 0644
const maxTmpDirNumBase int64 = 0xFFFFFFFF
var maxTmpDirNum *big.Int

func init() {
	maxTmpDirNum = big.NewInt(maxTmpDirNumBase)
}

// get the user's shell and ssh keys
func (u *User) fillInUser() error {
	passwd, err := os.Open("/etc/passwd")
	if err != nil {
		return err
	}
	defer passwd.Close()
	pl := bufio.NewScanner(passwd)
	for pl.Scan() {
		line := pl.Text()
		if strings.HasPrefix(line, u.Username) {
			fields := strings.Split(line, ":")
			u.Shell = fields[len(fields)-1]
			break
		}
	}
	if err = pl.Err(); err != nil {
		return err
	}
	
	authorizedKeyFile := u.authorizedKeyPath()

	if aKeys, err := os.Open(authorizedKeyFile); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		defer aKeys.Close()
		authKeys := bufio.NewScanner(aKeys)
		for authKeys.Scan() {
			u.SSHKeys = append(u.SSHKeys, authKeys.Text())
		}
		if err = authKeys.Err(); err != nil {
			return err
		}
	}
	
	return nil
}

func (u *User) update() error {
	// TODO: update other fields besides just SSH keys, like shell, various
	// /etc/passwd entries, etc.
	err := u.writeOutKeys()
	if err != nil {
		return err
	}

	return nil
}

func (u *User) writeOutKeys() error {
	if len(u.SSHKeys) == 0 {
		err := fmt.Errorf("no SSH keys given for %s", u.Username)
		return err
	}

	authorizedKeyFile := u.authorizedKeyPath()
	authorizedKeyDir := path.Dir(authorizedKeyFile)

	log.Printf("auth key dir: %s", authorizedKeyDir)
	if _, err := os.Stat(authorizedKeyDir); err != nil {
		log.Println("stat failed, as expected")
		if !os.IsNotExist(err) {
			return err
		}
		err = os.Mkdir(authorizedKeyDir, sshDirPerm)
		if err != nil {
			return err
		}
		sshDir, err := os.Open(authorizedKeyDir)
		if err != nil {
			return err
		}
		uid, gid, err := u.getUidGid()
		if err != nil {
			return err
		}

		err = sshDir.Chown(uid, gid)
		if err != nil {
			return err
		}
	} 

	tmpAuthKeys, err := u.createTempAuthKeyFile(authorizedKeyDir)
	if err != nil {
		return err
	}

	for _, l := range u.SSHKeys {
		_, err = tmpAuthKeys.WriteString(l)
		if err != nil {
			tmpAuthKeys.Close()
			return err
		}
		_, err = tmpAuthKeys.WriteString("\n")
		if err != nil {
			tmpAuthKeys.Close()
			return err
		}
	}
	
	tmpAuthKeyPath := tmpAuthKeys.Name()
	tmpAuthKeys.Close()
	err = os.Rename(tmpAuthKeyPath, authorizedKeyFile)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) deleteAuthKeys() error {
	authKeyPath := u.authorizedKeyPath()
	err := os.Remove(authKeyPath)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) authorizedKeyPath() string {
	return path.Join(u.HomeDir, ".ssh", "authorized_keys")
}

func New(userName string, fullName string, homeDir string, shell string, groups []string, keys []string) (*User, error) {
	if homeDir == "" {
		homeDir = path.Join(DefaultHomeBase, userName)
	}
	if shell == "" {
		shell = DefaultShell
	}

	// check for an existing user
	xu, _ := user.Lookup(userName)
	if xu != nil {
		err := fmt.Errorf("user %s already exists", userName)
		return nil, err
	}

	u, err := osCreateUser(userName, fullName, homeDir, shell, groups)
	log.Printf("user '%s' home: '%s' is? %+v", u.Username, u.HomeDir, u)
	if err != nil {
		return nil, err
	}
	u.SSHKeys = keys

	err = u.writeOutKeys()
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (u *User) createTempAuthKeyFile(baseDir string) (*os.File, error) {
	uid, gid, err := u.getUidGid()
	if err != nil {
		return nil, err
	}

	n, err := rand.Int(rand.Reader, maxTmpDirNum)
	if err != nil {
		return nil, err
	}

	tmpAuthKeyPath := path.Join(baseDir, strings.Join([]string{"authorized_keys", n.String()}, "-"))
	tmpAuthKeyFile, err := os.Create(tmpAuthKeyPath)
	if err != nil {
		return nil, err
	}
	err = tmpAuthKeyFile.Chmod(authKeyPerm)
	if err != nil {
		return nil, err
	}
	err = tmpAuthKeyFile.Chown(uid, gid)
	if err != nil {
		return nil, err
	}

	return tmpAuthKeyFile, nil
}

func (u *User) getUidGid() (int, int, error) {
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return 0, 0, err
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return 0, 0, err
	}
	return uid, gid, nil
}