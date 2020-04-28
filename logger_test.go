//(c) softland 2020
//softlandia@gmail.com
package glasio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckMeaasges(t *testing.T) {
	chMsg := make(tCheckMsg, 2)
	chMsg[0] = "message 1"
	chMsg[1] = "message 2"
	assert.Contains(t, chMsg.String(), "1")
	assert.NotContains(t, chMsg.String(), "0")
}
