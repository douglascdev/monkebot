module monkebot

go 1.23.0

require (
	github.com/Potat-Industries/go-potatFilters v0.1.1
	github.com/douglascdev/buttifier v0.1.3
	github.com/gempir/go-twitch-irc/v4 v4.0.0
	github.com/ncruces/go-sqlite3 v0.18.2
	github.com/rs/zerolog v1.33.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/julianday v1.0.0 // indirect
	github.com/speedata/hyphenation v1.0.2 // indirect
	github.com/tetratelabs/wazero v1.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
)

replace github.com/gempir/go-twitch-irc/v4 => github.com/douglascdev/go-twitch-irc/v4 v4.0.0-20240923162405-9bd425cb6891
