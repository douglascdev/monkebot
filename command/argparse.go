package command

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type argType int

const (
	positional argType = iota
	named
)

var (
	ErrInvalidArgLength = errors.New("invalid arg length")
	ErrInvalidArgType   = errors.New("invalid arg type")
	ErrNoMatchingSpec   = errors.New("no spec matches args")
)

type cmdArg struct {
	Name     string
	Value    string
	Type     argType
	ArgIndex int // intended index for this arg in the args list(for positional args)
}

func parseArgs(msg string) *map[string]cmdArg {
	args := strings.Split(strings.TrimSpace(msg), " ")
	if len(args) <= 1 {
		return nil
	}

	cmdArgs := make([]cmdArg, len(args)-1)

	for i, arg := range args[1:] {
		before, after, found := strings.Cut(arg, ":")
		if found {
			cmdArgs[i] = cmdArg{Name: before, Value: after, Type: named}
		} else {
			cmdArgs[i] = cmdArg{Name: arg, Value: arg, Type: positional, ArgIndex: i}
		}
	}

	nameToCmdArg := make(map[string]cmdArg, len(cmdArgs))
	for _, arg := range cmdArgs {
		nameToCmdArg[arg.Name] = arg
	}
	return &nameToCmdArg
}

func getFirstMatchingArgSpec(nameToCmdArg *map[string]cmdArg, argSpecs ...interface{}) (result interface{}, err error) {
	var hasValidFieldCount bool
	for _, argSpec := range argSpecs {
		t := reflect.TypeOf(argSpec)
		nFields := 0
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.IsExported() {
				nFields++
			}
		}
		if t.NumField() == len(*nameToCmdArg) {
			hasValidFieldCount = true
			break
		}
	}

	if !hasValidFieldCount {
		err = ErrInvalidArgLength
		return
	}

	for _, argSpec := range argSpecs {
		t := reflect.TypeOf(argSpec)
		if t.Kind() != reflect.Ptr {
			err = fmt.Errorf("%s is not a pointer", t)
			return
		}

		t = t.Elem()
		if t.Kind() != reflect.Struct {
			err = fmt.Errorf("%s is not a struct", t)
			return
		}

		validFields := 0
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}

			var (
				required   bool
				argType    argType      = positional
				argVarType reflect.Type = field.Type
			)

			switch field.Tag.Get("argtype") {
			case "named":
				argType = named
			case "positional":
				argType = positional
			default:
				err = fmt.Errorf("invalid argtype, must be 'named' or 'positional': %s", field.Tag.Get("argtype"))
				return
			}

			switch field.Tag.Get("required") {
			case "true":
				required = true
			case "false":
				required = false
			default:
				err = fmt.Errorf("invalid required value, must be 'true' or 'false': %s", field.Tag.Get("required"))
				return
			}

		}

	}

	err = ErrNoMatchingSpec
	return
}
