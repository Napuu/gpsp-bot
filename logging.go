package main

import (
	"log"
	"reflect"
)

type loggingDecorator struct {
	handler ContextHandler
	name    string
}

func HandlerLogger(h ContextHandler) *loggingDecorator {
	return &loggingDecorator{
			handler: h,
			name:    getTypeName(h),
	}
}

func getTypeName(h interface{}) string {
	t := reflect.TypeOf(h)
	if t.Kind() == reflect.Ptr {
			t = t.Elem()
	}
	return t.Name()
}


func (l *loggingDecorator) execute(m *Context) {
	log.Printf("entering %s", l.name)
	
	l.handler.execute(m)
}

func (l *loggingDecorator) setNext(next ContextHandler) {
	l.handler.setNext(next)
}