package enums

type Length int64

const (
	Millimeter Length = 1
	Centimeter        = 10 * Millimeter
	Meter             = 100 * Centimeter
	Kilometer         = 1000 * Meter
)
