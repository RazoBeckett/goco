package cli

import (
	"image/color"

	"charm.land/fang/v2"
	lipglossv2 "charm.land/lipgloss/v2"
)

const (
	electricOrange = "#FF6A00"
	tangerineShock = "#FF8C1A"
	sunburstSurge  = "#FF9F40"
	mangoVolt      = "#FFC266"
	creamGleam     = "#FFF1E6"
	deepMocha      = "#392E29"
	lipstickRed    = "#FD0040"
)

func FangColorScheme(ld lipglossv2.LightDarkFunc) fang.ColorScheme {
	return fang.ColorScheme{
		Base:           ld(lipglossv2.Color(electricOrange), lipglossv2.Color(creamGleam)),
		Title:          lipglossv2.Color(electricOrange),
		Description:    ld(lipglossv2.Color(tangerineShock), lipglossv2.Color(creamGleam)),
		Codeblock:      nil,
		Program:        ld(lipglossv2.Color(electricOrange), lipglossv2.Color(creamGleam)),
		DimmedArgument: ld(lipglossv2.Color(sunburstSurge), lipglossv2.Color(mangoVolt)),
		Comment:        ld(lipglossv2.Color(tangerineShock), lipglossv2.Color(mangoVolt)),
		Flag:           ld(lipglossv2.Color(tangerineShock), lipglossv2.Color(mangoVolt)),
		FlagDefault:    ld(lipglossv2.Color(sunburstSurge), lipglossv2.Color(creamGleam)),
		Command:        ld(lipglossv2.Color(electricOrange), lipglossv2.Color(tangerineShock)),
		QuotedString:   ld(lipglossv2.Color(tangerineShock), lipglossv2.Color(creamGleam)),
		Argument:       ld(lipglossv2.Color(electricOrange), lipglossv2.Color(creamGleam)),
		Help:           lipglossv2.Color(sunburstSurge),
		Dash:           lipglossv2.Color(tangerineShock),
		ErrorHeader: [2]color.Color{
			lipglossv2.Color(creamGleam),
			lipglossv2.Color(lipstickRed),
		},
		ErrorDetails: lipglossv2.Color(lipstickRed),
	}
}
