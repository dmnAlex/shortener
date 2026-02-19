package pool

import "testing"

type MockItem struct {
	Value int
}

func (m *MockItem) Reset() {
	m.Value = 0
}

func TestPool(t *testing.T) {
	factory := func() *MockItem { return &MockItem{Value: 42} }
	p := New(factory)

	// Тест Get из пустого пула (должна отработать фабрика)
	item := p.Get()
	if item.Value != 42 {
		t.Errorf("expected 42, got %d", item.Value)
	}

	// Тест Put и Reset
	item.Value = 100
	p.Put(item) // Здесь вызывается Reset(), Value должно стать 0

	// Тест Get после Put (должен вернуться тот же объект, но сброшенный)
	reusedItem := p.Get()
	if reusedItem.Value != 0 {
		t.Errorf("expected 0 after reset, got %d", reusedItem.Value)
	}

	// Проверка, что это тот же указатель
	if reusedItem != item {
		t.Error("expected the same pointer to be reused")
	}
}
