package main

type handler interface {
	execute(*GenericMessage)
	setNext(handler)
}


func main() {
	runTelegramBot()	

	return 
}
