package entity

type Model struct {
	name      string
	maxTokens int
}

func NewModel(name string, maxTokens int) *Model {
	return &Model{
		name:      name,
		maxTokens: maxTokens,
	}
}

func (this *Model) GetName() string {
	return this.name
}

func (this *Model) GetMaxTokens() int {
	return this.maxTokens
}
