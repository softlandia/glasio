package glasio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToCsvString(t *testing.T) {
	w := TWarning{1, 1, 1, "first"}
	assert.Equal(t, "1; 1; \"first\"", w.ToCsvString())

	w = TWarning{2, 2, 2, " второе сообщение "}
	assert.Equal(t, "2,\t 2,\t \" второе сообщение \"", w.ToCsvString(",\t"))

	w = TWarning{1, 1, 1, "first"}
	assert.Equal(t, "1, 1, \"first\"", w.ToCsvString(","))
}

func TestLasWarningsToString(t *testing.T) {
	warnings := TLasWarnings{}
	assert.Equal(t, "", warnings.ToString(""))

	warnings = TLasWarnings{
		TWarning{1, 1, 1, "first"},
		TWarning{2, 2, 2, "second"},
	}
	assert.Equal(t, "1, 1, \"first\"#2, 2, \"second\"#", warnings.ToString("#"))
	assert.Equal(t, "1, 1, \"first\"\n2, 2, \"second\"\n", warnings.ToString("\n"))
}
