// (c) softland 2020
// softlandia@gmail.com
// constants

package glasio

///format strings for output LAS file
const (
	_LasFirstLine      = "~Version information\n"
	_LasVersion        = "VERS.                          %3.1f : glas (c) softlandia@gmail.com\n"
	_LasWrap           = "WRAP.                          NO  : ONE LINE PER DEPTH STEP\n"
	_LasWellInfoSec    = "~Well information\n"
	_LasMnemonicFormat = "#MNEM.UNIT DATA                                  :DESCRIPTION\n"
	_LasStrt           = " STRT.M %8.3f                                    :START DEPTH\n"
	_LasStop           = " STOP.M %8.3f                                    :STOP  DEPTH\n"
	_LasStep           = " STEP.M %8.3f                                    :STEP\n"
	_LasNull           = " NULL.  %9.3f                                   :NULL VALUE\n"
	_LasRkb            = " RKB.M %8.3f                                     :KB or GL\n"
	_LasXcoord         = " XWELL.M %8.3f                                   :Well head X coordinate\n"
	_LasYcoord         = " YWELL.M %8.3f                                   :Well head Y coordinate\n"
	_LasOilComp        = " COMP.  %-43.43s:OIL COMPANY\n"
	_LasWell           = " WELL.   %-43.43s:WELL\n"
	_LasField          = " FLD .  %-43.43s:FIELD\n"
	_LasLoc            = " LOC .  %-43.43s:LOCATION\n"
	_LasCountry        = " CTRY.  %-43.43s:COUNTRY\n"
	_LasServiceComp    = " SRVC.  %-43.43s:SERVICE COMPANY\n"
	_LasDate           = " DATE.  %-43.43s:DATE\n"
	_LasAPI            = " API .  %-43.43s:API NUMBER\n"
	_LasUwi            = " UWI .  %-43.43s:UNIVERSAL WELL INDEX\n"
	_LasCurvSec        = "~Curve Information Section\n"
	_LasCurvFormat     = "#MNEM.UNIT                 :DESCRIPTION\n"
	_LasCurvDept       = " DEPT.M                    :\n"
	_LasCurvLine       = " %s.%s                     :\n"
	_LasDataSec        = "~ASCII Log Data\n"

	//secName: 0 - empty, 1 - Version, 2 - Well info, 3 - Curve info, 4 - dAta
	lasSecIgnore   = 0
	lasSecVersion  = 1
	lasSecWellInfo = 2
	lasSecCurInfo  = 3
	lasSecData     = 4
)
