package main

type ContextHandler interface {
	execute(*Context)
	setNext(ContextHandler)
}


func main() {
	runTelegramBot()	
}
