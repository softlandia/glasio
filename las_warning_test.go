package glasio

import (
	"testing"
)

func TestToCsvString(t *testing.T) {
	w := TWarning{1, 1, 1, "first"}
	if w.ToCsvString() != "1;1;\"first\"" {
		t.Errorf("<TWarning.ToString()> return not correct string: %s\n", w.ToCsvString())
	}
	w = TWarning{1, 1, 1, "first"}
	if w.ToCsvString(",\t") != "1,\t1,\t\"first\"" {
		t.Errorf("<TWarning.ToString()> return not correct string: %s\n", w.ToCsvString())
	}
}

func TestLasWarningsToString(t *testing.T) {
	warnings := TLasWarnings{}
	if warnings.ToString("") != "" {
		t.Errorf("<TLasWarnings.ToString> on empty input return not empty string\n")
	}
	warnings = TLasWarnings{
		TWarning{1, 1, 1, "first"},
		TWarning{2, 2, 2, "second"},
	}
	s := warnings.ToString("#")
	if s != "1;1;\"first\"#2;2;\"second\"#" {
		t.Errorf("<TLasWarnings.ToString> not correct return: %s\n", s)
	}
	s = warnings.ToString("\n")
	if s != "1;1;\"first\"\n2;2;\"second\"\n" {
		t.Errorf("<TLasWarnings.ToString> not correct return: %s\n", s)
	}
}
