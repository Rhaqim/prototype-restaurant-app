package controllers

import (
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
)

// Signup ...
func AccountCreat(c *gin.Context) {
	var request hp.BankAccount
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var init = ut.InitBankApi(ut.POST, "customers", ut.StructToJSON(request), "askmdjongjors")

	var response ut.BankAPI = init

	status, body, err := response.Call()

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, gin.H{"body": body})

}
