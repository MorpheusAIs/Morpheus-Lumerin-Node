package lib

import (
	"net/url"
	"strings"
)

func SetWorkerName(u *url.URL, workerName string) {
	accountName, _, _ := SplitUsername(u.User.Username())
	SetUserName(u, JoinUsername(accountName, workerName))
}

func SetUserName(u *url.URL, userName string) {
	pwd, _ := u.User.Password()
	u.User = url.UserPassword(userName, pwd)
}

func SplitUsername(username string) (accountName string, workerName string, ok bool) {
	return strings.Cut(username, ".")
}

func JoinUsername(accountName, userName string) string {
	return accountName + "." + userName
}

func CopyURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	var userInfo *url.Userinfo
	user := u.User.Username()
	pwd, hasPwd := u.User.Password()
	if hasPwd {
		userInfo = url.UserPassword(user, pwd)
	} else {
		userInfo = url.User(user)
	}
	return &url.URL{
		Scheme:      u.Scheme,
		User:        userInfo,
		Host:        u.Host,
		Path:        u.Path,
		RawQuery:    u.RawQuery,
		OmitHost:    u.OmitHost,
		Opaque:      u.Opaque,
		RawPath:     u.RawPath,
		ForceQuery:  u.ForceQuery,
		RawFragment: u.RawFragment,
		Fragment:    u.Fragment,
	}
}
