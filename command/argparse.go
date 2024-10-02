package command

import "strings"

type argType int

const (
	positional argType = iota
	named
)

type cmdArg struct {
	Name     string
	Value    string
	Type     argType
	ArgIndex int // intended index for this arg in the args list(for positional args)
}

func parseArgs(msg string) *[]cmdArg {
	args := strings.Split(strings.TrimSpace(msg), " ")
	if len(args) <= 1 {
		return nil
	}

	result := make([]cmdArg, len(args)-1)

	for i, arg := range args[1:] {
		before, after, found := strings.Cut(arg, ":")
		if found {
			result[i] = cmdArg{Name: before, Value: after, Type: named}
		} else {
			result[i] = cmdArg{Name: arg, Value: arg, Type: positional, ArgIndex: i}
		}
	}

	return &result
}
