package converter

func Convert(currency Currency) (money float64) {
	return currency.IncreaseDenomination()
}

type Currency int

func ReduceDenomination(f float64) uint64 {
	return uint64((f * 100) + 0.5)
}

func (m Currency) IncreaseDenomination() float64 {
	x := float64(m)
	x = x / 100
	return x
}
