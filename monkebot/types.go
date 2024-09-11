package monkebot

type PlatformClient interface {
	Join(channels ...PlatformUser)
	Part(channels ...PlatformUser)

	Say(channel PlatformUser, message string)

	Connect() error
}

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
