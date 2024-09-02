package main

import "monkebot/monkebot"

func main() {
	mb := monkebot.NewMonkebot([]string{"hash_table"})
	print(mb.InitialChannels)
}
