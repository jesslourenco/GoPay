package internal

import "github.com/julienschmidt/httprouter"

func Router(registers ...HandlerRegister) *httprouter.Router {
	router := httprouter.New()

	for _, r := range registers {
		r.Register(router)
	}
	return router
}
