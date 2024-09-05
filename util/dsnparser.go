package util

import (
	"net/url"
	"strings"
)

func DSNParse(dsn string) (host, port, dbname, user, password string) {
    u, _ := url.Parse(dsn)

    if u.User != nil {
        user = u.User.Username()
        password, _ = u.User.Password()
    }

    host = u.Hostname()
    port = u.Port()

    dbname = strings.TrimPrefix(u.Path, "/")

    return host, port, dbname, user, password
}