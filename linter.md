las.go:42:2: `_LasDate` is unused (varcheck)
	_LasDate           = " DATE.  %-43.43s:DATE\n"
	^
las.go:34:2: `_LasXcoord` is unused (varcheck)
	_LasXcoord         = " XWELL.M %8.3f                                   :Well head X coordinate\n"
	^
las.go:44:2: `_LasUwi` is unused (varcheck)
	_LasUwi            = " UWI .  %-43.43s:UNIVERSAL WELL INDEX\n"
	^
las.go:46:2: `_LasCurvFormat` is unused (varcheck)
	_LasCurvFormat     = "#MNEM.UNIT                 :DESCRIPTION\n"
	^
las.go:39:2: `_LasLoc` is unused (varcheck)
	_LasLoc            = " LOC .  %-43.43s:LOCATION\n"
	^
las.go:43:2: `_LasAPI` is unused (varcheck)
	_LasAPI            = " API .  %-43.43s:API NUMBER\n"
	^
las.go:35:2: `_LasYcoord` is unused (varcheck)
	_LasYcoord         = " YWELL.M %8.3f                                   :Well head Y coordinate\n"
	^
las.go:40:2: `_LasCountry` is unused (varcheck)
	_LasCountry        = " CTRY.  %-43.43s:COUNTRY\n"
	^
las.go:41:2: `_LasServiceComp` is unused (varcheck)
	_LasServiceComp    = " SRVC.  %-43.43s:SERVICE COMPANY\n"
	^
las.go:28:2: `_LasMnemonicFormat` is unused (varcheck)
	_LasMnemonicFormat = "#MNEM.UNIT DATA                                  :DESCRIPTION\n"
	^
las.go:49:2: `_LasCurvLine2` is unused (varcheck)
	_LasCurvLine2      = " %s                        :\n"
	^
las.go:51:2: `_LasDataLine` is unused (varcheck)
	_LasDataLine       = ""
	^
las.go:38:2: `_LasField` is unused (varcheck)
	_LasField          = " FLD .  %-43.43s:FIELD\n"
	^
las.go:36:2: `_LasOilComp` is unused (varcheck)
	_LasOilComp        = " COMP.  %-43.43s:OIL COMPANY\n"
	^
las.go:33:2: `_LasRkb` is unused (varcheck)
	_LasRkb            = " RKB.M %8.3f                                     :KB or GL\n"
	^
example2\main.go:44:2: ineffectual assignment to `err` (ineffassign)
	err = las.Save(las.FileName + "-") //сохраняем с символом минус в расширении
	^
glasio_test.go:108:5: ineffectual assignment to `err` (ineffassign)
	n, err := las.Open("empty.las")
	   ^
glasio_test.go:131:5: ineffectual assignment to `err` (ineffassign)
	n, err = las.Open("empty.las")
	   ^
example\main.go:48:17: Error return value of `las.SaveWarning` is not checked (errcheck)
	las.SaveWarning("1.warning.md")
	               ^
example2\main.go:50:15: Error return value of `filepath.Walk` is not checked (errcheck)
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
	             ^
glasio_test.go:20:10: Error return value of `las.Open` is not checked (errcheck)
	las.Open(fp.Join("data/more_20_warnings.las"))
	        ^
glasio_test.go:45:11: Error return value of `las.Open` is not checked (errcheck)
		las.Open(tmp.fn)
		        ^
glasio_test.go:75:10: Error return value of `las.Open` is not checked (errcheck)
	las.Open(fp.Join("data/expand_points_01.las"))
	        ^
glasio_test.go:77:13: Error return value of `las.SetNull` is not checked (errcheck)
	las.SetNull(-999.25)
	           ^
glasio_test.go:79:10: Error return value of `las.Save` is not checked (errcheck)
	las.Save("-tmp.las")
	        ^
glasio_test.go:126:13: Error return value of `las.SetNull` is not checked (errcheck)
	las.SetNull(100.001)
	           ^
las.go:285:21: Error return value of `writer.WriteString` is not checked (errcheck)
		writer.WriteString(w.String())
		                  ^
las.go:286:21: Error return value of `writer.WriteString` is not checked (errcheck)
		writer.WriteString("\n")
		                  ^
las.go:299:19: Error return value of `oFile.WriteString` is not checked (errcheck)
	oFile.WriteString("**file: " + o.FileName + "**\n")
	                 ^
las.go:301:19: Error return value of `oFile.WriteString` is not checked (errcheck)
	oFile.WriteString("\n")
	                 ^
las.go:649:27: Error return value of `o.setActuallyNumberPoints` is not checked (errcheck)
	o.setActuallyNumberPoints(i)
	                         ^
las_test.go:43:20: Error return value of `las.ReadWellParam` is not checked (errcheck)
		las.ReadWellParam(tmp.s)
		                 ^
las_test.go:81:18: Error return value of `las.SaveWarning` is not checked (errcheck)
		las.SaveWarning(fp.Join("data/more_20_warnings.wrn"))
		               ^
las_test.go:144:17: Error return value of `las.LoadHeader` is not checked (errcheck)
		las.LoadHeader()
		              ^
las_util.go:33:16: Error return value of `las.LoadHeader` is not checked (errcheck)
	las.LoadHeader()
	              ^
las.go:733:17: SA1006: printf-style function with dynamic format string and no further arguments should use print-style function instead (staticcheck)
	fmt.Fprintf(f, _LasDataSec)
	               ^
las.go:736:17: SA1006: printf-style function with dynamic format string and no further arguments should use print-style function instead (staticcheck)
	fmt.Fprintf(f, o.convertStrToOut(s))
	               ^
las.go:509:29: S1019: should use make([]float64, o.expPoints) instead (gosimple)
	newDept := make([]float64, o.expPoints, o.expPoints)
	                           ^
las.go:513:28: S1019: should use make([]float64, o.expPoints) instead (gosimple)
	newLog := make([]float64, o.expPoints, o.expPoints)
	                          ^
las.go:524:30: S1019: should use make([]float64, o.expPoints) instead (gosimple)
		newDept := make([]float64, o.expPoints, o.expPoints)
		                           ^
las_param.go:123:23: S1019: should use make([]float64, n) instead (gosimple)
	t := make([]float64, n, n)
	                     ^
las_param.go:126:22: S1019: should use make([]float64, n) instead (gosimple)
	t = make([]float64, n, n)
	                    ^
las.go:261:10: S1003: should use strings.Contains(strings.ToUpper(o.Wrap), "Y") instead (gosimple)
	return (strings.Index(strings.ToUpper(o.Wrap), "Y") >= 0)
	        ^
