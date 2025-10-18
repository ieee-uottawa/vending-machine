package model

type Slot int

const (
	A1 Slot = iota
	A2
	A3
	A4
	B1
	B2
	B3
	B4
	C1
	C2
	C3
	C4
	D1
	D2
	D3
	D4
	D5
	D6
	D7
	D8
	E1
	E2
	E3
	E4
	E5
	E6
	E7
	E8
	F1
	F2
	F3
	F4
)

type Item struct {
	Id    int     `json:"id"`
	Name  string  `json:"name"`
	Price float32 `json:"price"`
	Stock int     `json:"stock"`
	Slot  Slot    `json:"slot"`
}
