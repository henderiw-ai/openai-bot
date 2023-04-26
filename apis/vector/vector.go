package vector

type Vector struct {
	Id     string    `json:"id,omitempty"`
	Tokens int       `json:"tokens,omitempty"`
	Data   string    `json:"data,omitempty"`
	Values []float32 `json:"values,omitempty"`
}
