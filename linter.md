las.go:679:21: Error return value of `writer.WriteString` is not checked (errcheck)
		writer.WriteString(w.String())
		                  ^
las.go:680:21: Error return value of `writer.WriteString` is not checked (errcheck)
		writer.WriteString("\n")
		                  ^
las.go:693:19: Error return value of `oFile.WriteString` is not checked (errcheck)
	oFile.WriteString("**file: " + las.FileName + "**\n")
	                 ^
las.go:695:19: Error return value of `oFile.WriteString` is not checked (errcheck)
	oFile.WriteString("\n")
	                 ^
las_header_test.go:44:11: Error return value of `las.Open` is not checked (errcheck)
		las.Open(tmp.fn)
		        ^
las_logger.go:50:16: Error return value of `f.WriteString` is not checked (errcheck)
		f.WriteString(msg)
		             ^
las_logger.go:82:16: Error return value of `f.WriteString` is not checked (errcheck)
		f.WriteString(s)
		             ^
las_logger.go:84:15: Error return value of `f.WriteString` is not checked (errcheck)
	f.WriteString("\n")
	             ^
las_test.go:50:10: Error return value of `las.Open` is not checked (errcheck)
	las.Open(fp.Join("data/more_20_warnings.las"))
	        ^
las_test.go:57:10: Error return value of `las.Open` is not checked (errcheck)
	las.Open(fp.Join("data/more_20_warnings.las"))
	        ^
las_test.go:62:17: Error return value of `las.SaveWarning` is not checked (errcheck)
	las.SaveWarning(fp.Join("data/more_20_warnings.wrn"))
	               ^
las_test.go:136:10: Error return value of `las.Save` is not checked (errcheck)
	las.Save("-tmp.las")
	        ^
las_test.go:277:11: Error return value of `las.Save` is not checked (errcheck)
		las.Save(tmp.fn)
		        ^
las_util.go:94:16: Error return value of `las.LoadHeader` is not checked (errcheck)
	las.LoadHeader()
	              ^
las_constants.go:13:2: `_LasMnemonicFormat` is unused (deadcode)
	_LasMnemonicFormat = "#MNEM.UNIT DATA                                  :DESCRIPTION\n"
	^
las_constants.go:18:2: `_LasRkb` is unused (deadcode)
	_LasRkb            = " RKB.M %8.3f                                     :KB or GL\n"
	^
las_constants.go:19:2: `_LasXcoord` is unused (deadcode)
	_LasXcoord         = " XWELL.M %8.3f                                   :Well head X coordinate\n"
	^
las_constants.go:20:2: `_LasYcoord` is unused (deadcode)
	_LasYcoord         = " YWELL.M %8.3f                                   :Well head Y coordinate\n"
	^
las_constants.go:21:2: `_LasOilComp` is unused (deadcode)
	_LasOilComp        = " COMP.  %-43.43s:OIL COMPANY\n"
	^
las_constants.go:23:2: `_LasField` is unused (deadcode)
	_LasField          = " FLD .  %-43.43s:FIELD\n"
	^
las_constants.go:24:2: `_LasLoc` is unused (deadcode)
	_LasLoc            = " LOC .  %-43.43s:LOCATION\n"
	^
las_constants.go:25:2: `_LasCountry` is unused (deadcode)
	_LasCountry        = " CTRY.  %-43.43s:COUNTRY\n"
	^
las_constants.go:26:2: `_LasServiceComp` is unused (deadcode)
	_LasServiceComp    = " SRVC.  %-43.43s:SERVICE COMPANY\n"
	^
las_constants.go:27:2: `_LasDate` is unused (deadcode)
	_LasDate           = " DATE.  %-43.43s:DATE\n"
	^
las_constants.go:28:2: `_LasAPI` is unused (deadcode)
	_LasAPI            = " API .  %-43.43s:API NUMBER\n"
	^
las_constants.go:29:2: `_LasUwi` is unused (deadcode)
	_LasUwi            = " UWI .  %-43.43s:UNIVERSAL WELL INDEX\n"
	^
las_constants.go:31:2: `_LasCurvFormat` is unused (deadcode)
	_LasCurvFormat     = "#MNEM.UNIT                 :DESCRIPTION\n"
	^
las_warning.go:13:2: `directOnWrite` is unused (deadcode)
	directOnWrite = 2
	^
las_summary_test.go:25:2: `werr` is unused (structcheck)
	werr bool //не используется
	^
las_test.go:73:5: ineffectual assignment to `err` (ineffassign)
	f, err := os.Create(fp.Join("data/w1_more_20_warnings.txt"))
	   ^
las_checker.go:102:28: S1019: should use make(CheckResults) instead (gosimple)
	res := make(CheckResults, 0)
	                          ^
las_logger.go:102:23: func `(*tMMnemonic).save` is unused (unused)
las_checker.go:74:25: func `CheckResults.isFatal` is unused (unused)
las_logger.go:48:21: func `(*tCheckMsg).save` is unused (unused)
las_logger.go:79:22: func `(*tCurvRprt).save` is unused (unused)
