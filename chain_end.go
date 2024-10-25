package main

type ChainEnd struct {}

func (h *ChainEnd) execute(m *GenericMessage) {
	panic("unimplemented")
}

func (h *ChainEnd) setNext(handler handler) {
	panic("cannot set next handler on ChainEnd")
}
