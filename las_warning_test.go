package glasio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToCsvString(t *testing.T) {
	w := TWarning{1, 1, 1, "first"}
	assert.Equal(t, "  2; \"first\"", w.ToCsvString())

	w = TWarning{2, 2, 2, " второе сообщение "}
	assert.Equal(t, "  3,\t \" второе сообщение \"", w.ToCsvString(",\t"))

	w = TWarning{1, 1, 1, "first"}
	assert.Equal(t, "  2, \"first\"", w.ToCsvString(","))
}

func TestLasWarningsToString(t *testing.T) {
	warnings := TLasWarnings{}
	assert.Equal(t, "", warnings.ToString(""))

	warnings = TLasWarnings{
		TWarning{1, 1, 1, "first"},
		TWarning{2, 2, 2, "second"},
	}
	assert.Equal(t, " 0,   2, \"first\"# 1,   3, \"second\"#", warnings.ToString("#"))
	assert.Equal(t, " 0,   2, \"first\"\n 1,   3, \"second\"\n", warnings.ToString("\n"))
}
