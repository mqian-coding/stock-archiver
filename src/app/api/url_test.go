package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitDefaultRequest(t *testing.T) {
	t.Run("build request url for ticker 'AMC'", func(t *testing.T) {
		req := InitDefaultRequest("AMC")
		assert.Equal(t,
			"https://query1.finance.yahoo.com/v8/finance/chart/AMC?region=US&lang=en-US&includePrePost=false&interval=1m&useYfid=true&range=1d&corsDomain=finance.yahoo.com&.tsrc=finance",
			req.String())
	})
}
