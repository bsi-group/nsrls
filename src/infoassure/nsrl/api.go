package main

// Holds the functionality for the HTTP API such as looking up values from the BTree

import (
	"github.com/gin-gonic/gin"
	"encoding/json"
	"net/http"
	"strings"
)

// ##### API Methods #########################################################

// Returns JSON data when looking up "hashes" via the URL value
func lookupSingleHash (c *gin.Context) {

	_, ret := bTree.Get(strings.ToUpper(c.Param("hash")))

	var jr JsonResult
	jr.Hash = c.Param("hash")
	jr.Exists = ret

	b, err := json.Marshal(jr)
	if err != nil {
		logger.Error("Error marshalling JSON for API: %v", err)
		c.String(http.StatusInternalServerError, "")
		return
	}

	c.String(http.StatusOK, "%s", string(b))
}

// Returns JSON data when looking up "hashes" via the "hashes" POST data parameter
func lookupMultipleHashes (c *gin.Context) {

	lookupValues := strings.Split(c.PostForm("hashes"), "#")

	returnValues := make([]JsonResult, 0)
	var ret bool

	for _, v := range lookupValues {
		_, ret = bTree.Get(strings.ToUpper(v))

		returnValues = append(returnValues, JsonResult{Hash: v, Exists: ret})
	}

	b, err := json.Marshal(returnValues)
	if err != nil {
		logger.Error("Error marshalling JSON for API: %v", err)
		c.String(http.StatusInternalServerError, "")
		return
	}

	c.String(http.StatusOK, "%s", string(b))
}