package main

import "fmt"

type TelegramResponder struct {
	next handler
}

func (r *TelegramResponder) execute(m *GenericMessage) {
	if (m.service != Telegram) {
		r.next.execute(m)
	}
	fmt.Println(m.service)
	fmt.Println("pingpong", m)
	switch m.service {
	case Telegram:
		fmt.Println("Responding to telegram")
	}
}

func (u *TelegramResponder) setNext(next handler) {
	u.next = next
}
