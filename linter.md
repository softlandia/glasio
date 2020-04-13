las_logger.go:49:21: func `(*tCheckMsg).msgFileOpenWarning` is unused (unused)
las_logger.go:65:23: func `(*tMMnemonic).save` is unused (unused)
las_logger.go:45:21: func `(*tCheckMsg).msgFileNoData` is unused (unused)
las_logger.go:41:21: func `(*tCheckMsg).msgFileIsWraped` is unused (unused)
las_logger.go:55:22: func `(*tCurvRprt).save` is unused (unused)
las_logger.go:35:21: func `(*tCheckMsg).save` is unused (unused)
las_logger.go:13:2: field `readedNumPoints` is unused (unused)
las_logger.go:14:2: field `errorOnOpen` is unused (unused)
las_util.go:86:6: func `lasOpenCheck` is unused (unused)
las.go:430:2: ineffectual assignment to `err` (ineffassign)
	err = nil
	^
las.go:25:2: `_LasCodePage` is unused (deadcode)
	_LasCodePage       = "CPAGE.                         1251: code page \n"
	^
las.go:28:2: `_LasMnemonicFormat` is unused (deadcode)
	_LasMnemonicFormat = "#MNEM.UNIT DATA                                  :DESCRIPTION\n"
	^
las.go:33:2: `_LasRkb` is unused (deadcode)
	_LasRkb            = " RKB.M %8.3f                                     :KB or GL\n"
	^
las.go:34:2: `_LasXcoord` is unused (deadcode)
	_LasXcoord         = " XWELL.M %8.3f                                   :Well head X coordinate\n"
	^
las.go:35:2: `_LasYcoord` is unused (deadcode)
	_LasYcoord         = " YWELL.M %8.3f                                   :Well head Y coordinate\n"
	^
las.go:36:2: `_LasOilComp` is unused (deadcode)
	_LasOilComp        = " COMP.  %-43.43s:OIL COMPANY\n"
	^
las.go:38:2: `_LasField` is unused (deadcode)
	_LasField          = " FLD .  %-43.43s:FIELD\n"
	^
las.go:39:2: `_LasLoc` is unused (deadcode)
	_LasLoc            = " LOC .  %-43.43s:LOCATION\n"
	^
las.go:40:2: `_LasCountry` is unused (deadcode)
	_LasCountry        = " CTRY.  %-43.43s:COUNTRY\n"
	^
las.go:41:2: `_LasServiceComp` is unused (deadcode)
	_LasServiceComp    = " SRVC.  %-43.43s:SERVICE COMPANY\n"
	^
las.go:42:2: `_LasDate` is unused (deadcode)
	_LasDate           = " DATE.  %-43.43s:DATE\n"
	^
las.go:43:2: `_LasAPI` is unused (deadcode)
	_LasAPI            = " API .  %-43.43s:API NUMBER\n"
	^
las.go:44:2: `_LasUwi` is unused (deadcode)
	_LasUwi            = " UWI .  %-43.43s:UNIVERSAL WELL INDEX\n"
	^
las.go:46:2: `_LasCurvFormat` is unused (deadcode)
	_LasCurvFormat     = "#MNEM.UNIT                 :DESCRIPTION\n"
	^
las.go:49:2: `_LasCurvLine2` is unused (deadcode)
	_LasCurvLine2      = " %s                        :\n"
	^
las.go:51:2: `_LasDataLine` is unused (deadcode)
	_LasDataLine       = ""
	^
las.go:260:2: `checkHeaderWrap` is unused (deadcode)
	checkHeaderWrap     = iota
	^
las.go:261:2: `checkHeaderCurve` is unused (deadcode)
	checkHeaderCurve    = iota
	^
las.go:262:2: `checkHeaderStrtStop` is unused (deadcode)
	checkHeaderStrtStop = iota
	^
example\main.go:48:17: Error return value of `las.SaveWarning` is not checked (errcheck)
	las.SaveWarning("1.warning.md")
	               ^
example2\main.go:44:10: Error return value of `las.Save` is not checked (errcheck)
	las.Save(las.FileName + "-") //сохраняем с символом минус в расширении
	        ^
example2\main.go:50:15: Error return value of `filepath.Walk` is not checked (errcheck)
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
	             ^
las.go:355:21: Error return value of `writer.WriteString` is not checked (errcheck)
		writer.WriteString(w.String())
		                  ^
las.go:356:21: Error return value of `writer.WriteString` is not checked (errcheck)
		writer.WriteString("\n")
		                  ^
las.go:369:19: Error return value of `oFile.WriteString` is not checked (errcheck)
	oFile.WriteString("**file: " + o.FileName + "**\n")
	                 ^
las.go:371:19: Error return value of `oFile.WriteString` is not checked (errcheck)
	oFile.WriteString("\n")
	                 ^
las.go:424:14: Error return value of `o.LoadHeader` is not checked (errcheck)
	o.LoadHeader()
	            ^
las.go:433:12: Error return value of `o.SetNull` is not checked (errcheck)
		o.SetNull(o.stdNull)
		         ^
las_logger.go:37:16: Error return value of `f.WriteString` is not checked (errcheck)
		f.WriteString(msg)
		             ^
las_logger.go:58:16: Error return value of `f.WriteString` is not checked (errcheck)
		f.WriteString(s)
		             ^
las_logger.go:60:15: Error return value of `f.WriteString` is not checked (errcheck)
	f.WriteString("\n")
	             ^
las_param_test.go:40:20: Error return value of `las.ReadWellParam` is not checked (errcheck)
		las.ReadWellParam(tmp.s)
		                 ^
las_test.go:42:10: Error return value of `las.Open` is not checked (errcheck)
	las.Open(fp.Join("data/more_20_warnings.las"))
	        ^
las_test.go:48:10: Error return value of `las.Open` is not checked (errcheck)
	las.Open(fp.Join("data/more_20_warnings.las"))
	        ^
las_test.go:50:18: Error return value of `las.SaveWarning` is not checked (errcheck)
		las.SaveWarning(fp.Join("data/more_20_warnings.wrn"))
		               ^
las_test.go:102:17: Error return value of `las.LoadHeader` is not checked (errcheck)
		las.LoadHeader()
		              ^
las_test.go:163:10: Error return value of `las.Open` is not checked (errcheck)
	las.Open(fp.Join("data/more_20_warnings.las"))
	        ^
las_test.go:221:13: Error return value of `las.SetNull` is not checked (errcheck)
	las.SetNull(-999.25)
	           ^
las_test.go:223:10: Error return value of `las.Save` is not checked (errcheck)
	las.Save("-tmp.las")
	        ^
las_test.go:258:14: Error return value of `las.SetNull` is not checked (errcheck)
		las.SetNull(tmp.newNull)
		           ^
las_test.go:276:13: Error return value of `las.SetNull` is not checked (errcheck)
	las.SetNull(-1000)
	           ^
las_test.go:284:11: Error return value of `las.Save` is not checked (errcheck)
		las.Save(tmp.fn)
		        ^
las_util.go:62:29: Error return value of `las.setActuallyNumberPoints` is not checked (errcheck)
	las.setActuallyNumberPoints(5)
	                           ^
las_util.go:81:16: Error return value of `las.LoadHeader` is not checked (errcheck)
	las.LoadHeader()
	              ^
las_summary_test.go:25:2: `werr` is unused (structcheck)
	werr bool
	^
