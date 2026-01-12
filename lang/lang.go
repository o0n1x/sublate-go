package lang

import "strings"

type Language string

const (
	English         Language = "EN"
	EnglishUS       Language = "EN-US"
	EnglishUK       Language = "EN-UK"
	Arabic          Language = "AR"
	Bulgarian       Language = "BG"
	Czech           Language = "CS"
	Danish          Language = "DA"
	German          Language = "DE"
	Greek           Language = "EL"
	Spanish         Language = "ES"
	Estonian        Language = "ET"
	Finnish         Language = "FI"
	French          Language = "FR"
	Hungarian       Language = "HU"
	Indonesian      Language = "ID"
	Italian         Language = "IT"
	Japanese        Language = "JA"
	Korean          Language = "KO"
	Lithuanian      Language = "LT"
	Latvian         Language = "LV"
	NorwegianBokmal Language = "NB"
	Dutch           Language = "NL"
	Polish          Language = "PL"
	Romanian        Language = "RO"
	Russian         Language = "RU"
	Slovak          Language = "SK"
	Slovenian       Language = "SL"
	Swedish         Language = "SV"
	Thai            Language = "TH"
	Turkish         Language = "TR"
	Ukrainian       Language = "UK"
	Vietnamese      Language = "VI"
	Chinese         Language = "ZH"
)

func (l Language) Upper() string {
	return strings.ToUpper(string(l))
}

func (l Language) Lower() string {
	return strings.ToLower(string(l))
}

func (l Language) String() string {
	return string(l)
}
