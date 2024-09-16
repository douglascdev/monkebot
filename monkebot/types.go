package monkebot

type User struct {
	ID           int
	PermissionID int64
}

type Platform struct {
	ID   int
	Name string
}

type PlatformUser struct {
	Platform
	User
	ID   string
	Name string
}
