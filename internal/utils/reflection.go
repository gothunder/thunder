package utils

import (
	"errors"
	"reflect"

	"github.com/TheRafaBonin/roxy"
)

var (
	ErrMethodNotFound          = errors.New("method not found")
	ErrInvaidNumberOfArguments = errors.New("invalid number of arguments")
	ErrInvalidArgumentType     = errors.New("invalid argument type")
)

func HasMethod(i interface{}, methodName string) bool {
	_, ok := reflect.TypeOf(i).MethodByName(methodName)
	return ok
}

func SafeCallMethod(i interface{}, methodName string, args []reflect.Value) ([]reflect.Value, error) {
	method, ok := reflect.TypeOf(i).MethodByName(methodName)
	if !ok {
		return nil, ErrMethodNotFound
	}

	if !method.Type.IsVariadic() && len(args)+1 != method.Type.NumIn() {
		return nil, roxy.Wrapf(ErrInvaidNumberOfArguments, "expected %d arguments, got %d", method.Type.NumIn(), len(args))
	}

	in := make([]reflect.Value, len(args)+1)
	in[0] = reflect.ValueOf(i)

	for k, arg := range args {
		in[k+1] = arg
	}

	return method.Func.Call(in), nil
}
