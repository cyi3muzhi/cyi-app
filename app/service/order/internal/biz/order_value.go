package biz

type Money struct {
	Amount   int64  // 单位：分
	Currency string // CNY, USD
}

func (m Money) Add(other Money) Money {
	return Money{m.Amount + other.Amount, m.Currency}
}

func (m Money) Subtract(other Money) Money {
	return Money{m.Amount - other.Amount, m.Currency}
}

func (m Money) Multiply(n int) Money {
	return Money{m.Amount * int64(n), m.Currency}
}

func (m Money) LessThan(other Money) bool {
	return m.Amount < other.Amount
}

func (m Money) IsZero() bool {
	return m.Amount == 0
}
