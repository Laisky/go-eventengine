package eventengine

import (
	"fmt"
	"reflect"

	"github.com/Laisky/go-eventengine/types"
)

// GetHandlerID calculate handler func's address as id
func GetHandlerID(handler Handler) types.HandlerID {
	return types.HandlerID(GetFuncAddress(handler))
}

// GetFuncAddress get address of func
func GetFuncAddress(v interface{}) string {
	ele := reflect.ValueOf(v)
	if ele.Kind() != reflect.Func {
		panic("only accept func")
	}

	return fmt.Sprintf("%x", ele.Pointer())
}
