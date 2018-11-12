// +build windows

/*
 * Copyright (c) 2018, Jeremy Bingham (<jeremy@goiardi.gl>)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// windows specific functions and methods for creating users

// just stubs right now though

package users

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ctdk/spqr/internal/processes"
	"github.com/hectane/go-acl"
	"github.com/tideland/golib/logger"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
	"unsafe"
)

// Obviously these need to change drastically
const sshDirPerm = 0700
const authKeyPerm = 0644

const inWindowsDev = true

const DefaultShell = "C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe"
const DefaultHomeBase = "c:/Users"

var notImpErr = errors.New("This Windows functionality is not implemented yet.")

func (u *User) osCreateUser() error {
	dummyPass := genRandomPassword()

	fullName := u.Name
	if fullName == "" {
		fullName = u.Username
	}
	nfullName, _ := windows.UTF16PtrFromString(fullName)
	nuComment, _ := windows.UTF16PtrFromString("Managed by spqr")
	nuname, _ := windows.UTF16PtrFromString(u.Username)
	npass, _ := windows.UTF16PtrFromString(dummyPass)
	newUserInfo := userInfo2{
		name: nuname,
		password: npass,
		fullName: nfullName,
		comment: nuComment,
		acctExpires: dword(timeqForever),
		priv: USER_PRIV_USER, // TODO: investigate being able to 
				      // set this to _ADMIN
		flags: UF_SCRIPT | UF_NORMAL_ACCOUNT | UF_DONT_EXPIRE_PASSWD,
	}
	ret, _, err := userAdd.Call(uintptr(0), uintptr(userInfoLevel), uintptr(unsafe.Pointer(&newUserInfo)), uintptr(0))
	if ret != NERR_Success {
		logger.Errorf("failed to add user %s, bailing with error '%s'", u.Username, err.Error())
		return err
	}

	// Now the fake login.
	//
	// Don't forget - we'll need to close the handles.
	var lHandle windows.Handle
	logonUserDomain, _ := windows.UTF16PtrFromString(".")

	retl, _, err := logonUser.Call(uintptr(unsafe.Pointer(nuname)), uintptr(unsafe.Pointer(logonUserDomain)), uintptr(unsafe.Pointer(npass)), uintptr(LOGON32_LOGON_BATCH), uintptr(LOGON32_PROVIDER_DEFAULT), uintptr(unsafe.Pointer(&lHandle)))
	// Sometimes 0 is success. Sometimes it is a failure. Thanks, y'all.
	if retl == 0 {
		logonErr := fmt.Errorf("Error doing simulated login with user %s: ret %d, msg: '%d'", u.Name, ret, err.Error())
		logger.Errorf(logonErr.Error())
		return logonErr
	}

	// And load the profile. This is what creates the home directory.

	var pinfo profileInfo
	pinfo.dwSize = dword(unsafe.Sizeof(pinfo))
	pinfo.lpUserName = nuname
	rlp, _, err := loadUserProfile.Call(uintptr(lHandle), uintptr(unsafe.Pointer(&pinfo)))

	// This may not be a failure-failure; don't bail if there's an err for
	// now.
	if rlp == 0 {
		lperr := fmt.Errorf("Error loading profile. Ret is %d, msg is: '%s'", rlp, err.Error())
		logger.Errorf(lperr.Error())
		return lperr
	}

	defer func() {
		r, _, rerr := unloadUserProfile.Call(uintptr(lHandle), uintptr(pinfo.hProfile))
		if r == 0 {
			logger.Errorf("unload user profile failed: %s", rerr.Error())
		}
		c, _, cerr := closeHandle.Call(uintptr(lHandle))
		if c == 0 {
			logger.Errorf("closeHandle failed: %s", cerr.Error())
		}
	}()

	// TODO: groups. This depends on if we *can* do groups, though.

	nu, err := Get(u.Username)
	if err != nil {
		return err
	}
	authKeys := u.AuthorizedKeys
	u = nu

	// save the keys
	err = u.writeOutKeys(authKeys)
	if err != nil {
		return err
	}

	return nil
}

func (u *User) update() error {
	if u.updated.authorizedKeys != nil {
		if err := u.writeOutKeys(u.updated.authorizedKeys); err != nil {
			return err
		}
	}

	if u.updated.name != "" {
		if err := u.updateName(); err != nil {
			return err
		}
	}

	if u.updated.reenable {
		if err := u.reactivate(); err != nil {
			return err
		}
	}

	return nil
}

// Windows specific disabling account bits go here, when they come along.
func (u *User) disable() error {
	nuname, _ := windows.UTF16PtrFromString(u.Username)
	uinfo := userInfo1008{
		flags: UF_ACCOUNTDISABLE,
	}
	ret, _, err := userSetInfo.Call(uintptr(0), uintptr(unsafe.Pointer(nuname)), uintptr(enableDisableLevel), uintptr(unsafe.Pointer(&uinfo)), uintptr(0))
	if ret != NERR_Success {
		logger.Errorf("failed to disable user %s, bailing with error '%s'", u.Username, err.Error())
		return err
	}
	return nil
}

func (u *User) active(enable bool) error {
	var enableArg string
	var errDesc string
	if enable {
		errDesc = "enable"
		enableArg = "yes"
	} else {
		errDesc = "disable"
		enableArg = "no"
	}

	netUserArgs := []string{"USER", u.Username, fmt.Sprintf("/ACTIVE:%s", enableArg)}
	if err := runNetCmd(netUserArgs); err != nil {
		ferr := fmt.Errorf("Error received while trying to %s user %s: %s", errDesc, u.Username, err.Error())
		return ferr
	}
	return nil
}

func (u *User) deactivate() error {
	return u.active(false)
}

func (u *User) reactivate() error {
	return u.active(true)
}

func (u *User) changeShell(shell string) error {
	return notImpErr
}

func (u *User) updateGroupInfo(uEntry *UserInfo, uUp *userUpdated) error {
	return nil
}

func getDefaultShell() string {
	// Try getting the default shell for OpenSSH from the registry. If that
	// somehow fails, return the DefaultShell const.
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\OpenSSH`, registry.QUERY_VALUE); err == nil {
		defer k.Close()
		shl, _, err := k.GetStringValue("DefaultShell")
		if err != nil {
			logger.Warningf("Looking up HKLM:\\SOFTWARE\\OpenSSH worked, but we could not retrieve the DefaultShell registry key and will return the DefaultShell const. Error was: %s", err.Error())
		}
		return shl
	} else {
		logger.Warningf("Could not look up HKLM:\\SOFTWARE\\OpenSSH in registry, will return what we think the DefaultShell would be. Error was: %s", err.Error())
	}
	return DefaultShell
}

func osMakeNewGroup(groupName string) error {
	return notImpErr
}

func getShell(username string) (string, error) {
	// At this time, it's not real obvious how to set individual user shells
	// with Windows OpenSSH. For the time being, return whatever the default
	// shell is.
	return getDefaultShell(), nil
}

func (u *User) clearExtraGroups() error {
	return notImpErr
}

func (u *User) reviewGroups(existingGroups map[string]bool) error {
	// Doesn't work in Windows, apparently, so don't try now.
	return nil
}

func (u *User) fillInGroups() error {
	// Doesn't work in Windows, apparently, so don't try now.
	return nil
}

func (u *User) updateName() error {
	netUserArgs := []string{"USER", u.Username, fmt.Sprintf("/FULLNAME:%s", u.updated.name)}
	if err := runNetCmd(netUserArgs); err != nil {
		ferr := fmt.Errorf("Error received while trying to disable user %s: %s", u.Username, err.Error())
		return ferr
	}

	return nil
}

func (u *User) createTempAuthKeyFile(baseDir string) (*os.File, error) {
	tmpAuthKeyFile, err := ioutil.TempFile(baseDir, "authorized_keys")
	if err != nil {
		return nil, err
	}

	err = acl.Chmod(tmpAuthKeyFile.Name(), authKeyPerm)
	if err != nil {
		return nil, err
	}

	err = acl.Apply(
		tmpAuthKeyFile.Name(),
		false,
		false,
		acl.GrantName(windows.GENERIC_READ|windows.GENERIC_WRITE, u.Username),
	)
	if err != nil {
		return nil, err
	}

	return tmpAuthKeyFile, nil
}

func (u *User) setSshDirOwnership(dir string) error {
	// set ACL on ~/.ssh to the Windows equivalent of 0700.
	if err := acl.Chmod(dir, sshDirPerm); err != nil {
		return err
	}

	// Better give Administrator and, uh, System? access and forbid others.
	if err := acl.Apply(
		dir,
		false,
		false,
		acl.GrantName(windows.GENERIC_READ|windows.GENERIC_WRITE, u.Username),
		// Find what these groups would actually be called, hrm.
		acl.GrantName(windows.GENERIC_READ|windows.GENERIC_WRITE, "Administrators"),
		acl.GrantName(windows.GENERIC_READ|windows.GENERIC_WRITE, "System"),
		acl.DenyName(windows.GENERIC_ALL, "Everyone"),
	); err != nil {
		return err
	}
	return nil
}

func (u *User) killProcesses() error {
	err := processes.KillUserProcesses(u.Uid)

	if err != nil && inWindowsDev {
		logger.Infof("Received an error from killing all %s processes, but since windows dev work is still in progress we'll ignore it. The error was %s.", u.Username, err.Error())
	}

	return err
}

func genRandomPassword() string {
	return randStringBytesMaskImprSrc(randStrLen)
}

func runNetCmd(netUserArgs []string) error {
	netCmdPath, err := exec.LookPath("NET")
	if err != nil {
		return err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	netUser := exec.Command(netCmdPath, netUserArgs...)
	netUser.Stdout = &stdout
	netUser.Stderr = &stderr

	if err := netUser.Run(); err != nil {
		ferr := errors.New(strings.Join([]string{err.Error(), stderr.String()}, " -- "))
		return ferr
	}
	return nil
}

// Stealing this *again*, from the goiardi sandbox test code. Originally:
// borrowing this from Stack Overflow (such as it ever is), located at
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang

const randStrLen = 12

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func randStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
