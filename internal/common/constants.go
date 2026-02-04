package common

var (
	TokenKey    = "Authorization"
	UserInfoKey = "claims"
)

// redis key
var (
	LoginAccessPrefix  = "login:access:"
	LoginRefreshPrefix = "login:refresh:"
	BlockedTokenPrefix = "blacklist:token:"
)
