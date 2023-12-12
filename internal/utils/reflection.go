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

	// check if the arguments are of the correct type
	for k, arg := range args {
		// if is greater than or equal last argument and is variadic
		if k+1 >= method.Type.NumIn()-1 && method.Type.IsVariadic() {
			if arg.Type() != method.Type.In(method.Type.NumIn()-1).Elem() {
				return nil, roxy.Wrapf(
					ErrInvalidArgumentType,
					"argument %d is of type %s, expected %s",
					k+1, arg.Type(), method.Type.In(method.Type.NumIn()-1).Elem(),
				)
			}
		} else {
			if arg.Type() != method.Type.In(k+1) {
				return nil, roxy.Wrapf(
					ErrInvalidArgumentType,
					"argument %d is of type %s, expected %s",
					k+1, arg.Type(), method.Type.In(k+1),
				)
			}
		}

		in[k+1] = arg
	}

	return method.Func.Call(in), nil
}
