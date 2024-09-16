package client

import "database/sql"

type PlatformClient interface {
	Join(channels ...string)
	Part(channels ...string)

	Say(channel string, message string)

	OnSelfJoin(channel string)
	OnSelfPart(channel string)
	OnMessage(channel string, message string)

	OnConnect()

	BeginTransaction() (*sql.Tx, error)
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
