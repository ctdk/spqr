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

package users

// Segregating definitions that work directly with Windows DLLs in
// here for organization's sake, because there's a lot of this gunk.

import (
	"golang.org/x/sys/windows"
)

var (
	netApi32 = windows.NewLazyDLL("netapi32.dll")
	advApi32 = windows.NewLazyDLL("advapi32.dll")
	userEnv = windows.NewLazyDLL("userenv.dll")
	userGetInfo = netApi32.NewProc("NetUserGetInfo")
	userAdd = netApi32.NewProc("NetUserAdd")
	logonUser = advApi32.NewProc("LogonUserW")
 	lsaAddAccountRights = advApi32.NewProc("LsaAddAccountRights")
	loadUserProfile = userEnv.NewProc("LoadUserProfileW")
	unloadUserProfile = userEnv.NewProc("UnloadUserProfile")
)

type dword uint32
type lpwstr *uint16

type userInfo2 struct {
	name lpwstr
	password lpwstr
	passwordAge dword
	priv dword
	homeDir lpwstr
	comment lpwstr
	flags dword
	scriptPath lpwstr
	authFlags dword
	fullName lpwstr
	usrComment lpwstr
	parms lpwstr
	workstations lpwstr
	lastLogon dword
	lastLogoff dword
	acctExpires dword
	maxStorage dword
	unitsPerWeek dword
	logonHours *byte
	badPwCount dword
	numLogons dword
	logonServer lpwstr
	countryCode dword
	codePage dword
}

type profileInfo struct {
	dwSize dword
	dwFlags dword
	lpUserName lpwstr
	lpProfilePath lpwstr
	lpDefaultPath lpwstr
	lpServerName lpwstr
	lpPolicyPath lpwstr
	hProfile windows.Handle
}

const (
	USER_PRIV_GUEST = 0
	USER_PRIV_USER  = 1
	USER_PRIV_ADMIN = 2

	UF_SCRIPT                          = 0x0001
	UF_ACCOUNTDISABLE                  = 0x0002
	UF_HOMEDIR_REQUIRED                = 0x0008
	UF_LOCKOUT                         = 0x0010
	UF_PASSWD_NOTREQD                  = 0x0020
	UF_PASSWD_CANT_CHANGE              = 0x0040
	UF_ENCRYPTED_TEXT_PASSWORD_ALLOWED = 0x0080

	UF_TEMP_DUPLICATE_ACCOUNT    = 0x0100
	UF_NORMAL_ACCOUNT            = 0x0200
	UF_INTERDOMAIN_TRUST_ACCOUNT = 0x0800
	UF_WORKSTATION_TRUST_ACCOUNT = 0x1000
	UF_SERVER_TRUST_ACCOUNT      = 0x2000

	UF_DONT_EXPIRE_PASSWD                     = 0x10000
	UF_MNS_LOGON_ACCOUNT                      = 0x20000
	UF_SMARTCARD_REQUIRED                     = 0x40000
	UF_TRUSTED_FOR_DELEGATION                 = 0x80000
	UF_NOT_DELEGATED                          = 0x100000
	UF_USE_DES_KEY_ONLY                       = 0x200000
	UF_DONT_REQUIRE_PREAUTH                   = 0x400000
	UF_PASSWORD_EXPIRED                       = 0x800000
	UF_TRUSTED_TO_AUTHENTICATE_FOR_DELEGATION = 0x1000000
	UF_NO_AUTH_DATA_REQUIRED                  = 0x2000000
	UF_PARTIAL_SECRETS_ACCOUNT                = 0x4000000
	UF_USE_AES_KEYS                           = 0x8000000

	LOGON32_LOGON_INTERACTIVE = 2
	LOGON32_LOGON_NETWORK = 3
	LOGON32_LOGON_BATCH = 4
	LOGON32_LOGON_SERVICE = 5
	LOGON32_LOGON_UNLOCK = 7
	LOGON32_LOGON_NETWORK_CLEARTEXT = 8
	LOGON32_LOGON_NEW_CREDENTIALS = 9

	LOGON32_PROVIDER_DEFAULT = 0
	LOGON32_PROVIDER_WINNT35 = 1
	LOGON32_PROVIDER_WINNT40 = 2
	LOGON32_PROVIDER_WINNT50 = 3

	userInfoLevel = 2

	NERR_Success = 0
)

const SE_BATCH_LOGON_NAME = "SeBatchLogonRight"

const (
	minUint32 uint32 = 0
	timeqForever = ^minUint32
)

func decodeLpwstr(winStr lpwstr) string {
	if winStr == nil {
		return ""
	}

	dref := make([]uint16, 0, 256)
	for p := uintptr(unsafe.Pointer(winStr));; p += 2 {
		u := *(*uint16)(unsafe.Pointer(p))
		if u == 0 {
			break
		}
		dref = append(dref, u)
	}
	return windows.UTF16ToString(dref)
}
