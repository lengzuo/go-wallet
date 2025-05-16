package utils

import "github.com/rs/xid"

func UUID() string {
	guid := xid.New()
	return guid.String()
}
