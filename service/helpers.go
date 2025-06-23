package service

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func (c *Service) SomethingWentWrong(ctx *gin.Context, logString string) {
	log.Println("Wedyta: " + logString + " url=" + ctx.Request.URL.String())
	ctx.String(http.StatusInternalServerError, "Something went wrong, see log for details.")
}
