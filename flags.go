package fangs

import (
	"fmt"
	"reflect"

	"github.com/spf13/pflag"

	logger "github.com/nextlinux/gologger"
)

// FlagAdder interface can be implemented by structs in order to add flags when AddFlags is called
type FlagAdder interface {
	AddFlags(flags FlagSet)
}

// AddFlags traverses the object graphs from the structs provided and calls all AddFlags methods implemented on them
func AddFlags(log logger.Logger, flags *pflag.FlagSet, structs ...any) {
	flagSet := NewPFlagSet(log, flags)
	for _, o := range structs {
		addFlags(log, flagSet, o)
	}
}

func addFlags(log logger.Logger, flags FlagSet, o any) {
	v := reflect.ValueOf(o)
	if !isPtr(v.Type()) {
		panic(fmt.Sprintf("AddFlags must be called with pointers, got: %#v", o))
	}

	invokeAddFlags(log, flags, o)

	v, t := base(v)

	if isStruct(t) {
		for i := 0; i < t.NumField(); i++ {
			v := v.Field(i)
			v = v.Addr()
			if !v.CanInterface() {
				continue
			}

			addFlags(log, flags, v.Interface())
		}
	}
}

func invokeAddFlags(log logger.Logger, flags FlagSet, o any) {
	defer func() {
		// we need to handle embedded structs having AddFlags methods called,
		// potentially adding flags with existing names
		if err := recover(); err != nil {
			log.Debugf("got error while invoking AddFlags: %#v", err)
		}
	}()

	if o, ok := o.(FlagAdder); ok && !isPromotedMethod(o, "AddFlags") {
		o.AddFlags(flags)
	}
}
