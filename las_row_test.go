//(c) softland 2020
//softlandia@gmail.com
package glasio

import (
	"testing"
)

func TestOpen2(t *testing.T) {
	/*
		las := NewLas()
		m, _ := las.Open2(fp.Join("data/logging_levels.las"))
		r := las.GetRows()
		assert.Equal(t, 29139, len(r))
		assert.Equal(t, 44, m)
		assert.Equal(t, "DATE.          05-Nov-08                          : Log Date                                                                 ", r[18])
		assert.Equal(t, "7273.50  -999.2500  -999.2500  -999.2500  -999.2500  -999.2500  -999.2500  -999.2500  -999.2500  -999.2500  -999.2500  -999.2500  -999.2500  -999.2500  -999.2500  -999.2500  -999.2500  -999.2500 ", r[29138])

		las = NewLas()
		m, _ = las.Open2(fp.Join("data/encodings_utf16be.las"))
		r = las.GetRows()
		assert.Equal(t, 46, len(r))
		assert.Equal(t, 43, m)
		assert.Equal(t, "~WELL ºᶟᵌᴬń BLOCK", r[3])
		assert.Equal(t, " WELL.                WELL:   Скв #12Ω", r[11])
	*/
}

func TestOpenData(t *testing.T) {
	/*
		las := NewLas()
		las.Open2(fp.Join("data/6038187_v1.2_short.las"))
		//assert.Equal(t, 12.0, las.Logs["SP"].D[0])
	*/
}
