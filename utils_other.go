package wedyta

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func (c *Impl) somethingWentWrong(context *gin.Context, logString string) {
	log.Println("Wedyta: " + logString)
	context.String(http.StatusInternalServerError, "Something went wrong, see log for details.")
}
