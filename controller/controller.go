package controller

import "github.com/pa-pe/wedyta/service"

type Controller struct {
	Service *service.Service
}

func NewController(service *service.Service) *Controller {
	return &Controller{
		Service: service,
	}
}
