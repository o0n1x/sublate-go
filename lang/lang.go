package lang

import "strings"

type Language string

const (
	AutoDetect         Language = "Nil"
	English            Language = "EN"
	EnglishUS          Language = "EN-US"
	EnglishUK          Language = "EN-GB"
	Arabic             Language = "AR"
	Bulgarian          Language = "BG"
	Czech              Language = "CS"
	Danish             Language = "DA"
	German             Language = "DE"
	Greek              Language = "EL"
	Spanish            Language = "ES"
	Estonian           Language = "ET"
	Finnish            Language = "FI"
	French             Language = "FR"
	Hungarian          Language = "HU"
	Indonesian         Language = "ID"
	Italian            Language = "IT"
	Japanese           Language = "JA"
	Korean             Language = "KO"
	Lithuanian         Language = "LT"
	Latvian            Language = "LV"
	NorwegianBokmal    Language = "NB"
	Dutch              Language = "NL"
	Polish             Language = "PL"
	Portuguese         Language = "PT"
	PortugueseBrazil   Language = "PT-BR"
	PortuguesePortugal Language = "PT-PT"
	Romanian           Language = "RO"
	Russian            Language = "RU"
	Slovak             Language = "SK"
	Slovenian          Language = "SL"
	Swedish            Language = "SV"
	Thai               Language = "TH"
	Turkish            Language = "TR"
	Ukrainian          Language = "UK"
	Vietnamese         Language = "VI"
	Chinese            Language = "ZH"
	ChineseSimplified  Language = "ZH-HANS"
	ChineseTraditional Language = "ZH-HANT"
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
