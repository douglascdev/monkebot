package monkebot

type PlatformClient interface {
	Join(channels ...PlatformUser)
	Part(channels ...PlatformUser)

	Say(channel PlatformUser, message string)

	Connect() error
}

type Platform struct {
	ID   int
	Name string
}

type PlatformUser struct {
	Platform
	ID   string
	Name string
}
