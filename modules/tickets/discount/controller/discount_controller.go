package discount_controller

import (
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/tickets/discount/services"
)

type DiscountController struct {
	service    discount_service.DiscountService
	logHandler *handler.LogHandler
}

func NewDiscountController(service discount_service.DiscountService, logHandler *handler.LogHandler) *DiscountController {
	return &DiscountController{
		service:    service,
		logHandler: logHandler,
	}
}
