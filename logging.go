package main

import (
	"log"
	"reflect"
)

type loggingDecorator struct {
	handler handler
	name    string
}

func HandlerLogger(h handler) *loggingDecorator {
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


func (l *loggingDecorator) execute(m *GenericMessage) {
	log.Printf("entering %s", l.name)
	
	l.handler.execute(m)
}

func (l *loggingDecorator) setNext(next handler) {
	l.handler.setNext(next)
}