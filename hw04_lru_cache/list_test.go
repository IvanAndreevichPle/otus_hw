package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})

	t.Run("edge cases", func(t *testing.T) {
		l := NewList()

		// Тест добавления и удаления одного элемента
		l.PushFront(1)
		require.Equal(t, 1, l.Len())
		require.Equal(t, 1, l.Front().Value)
		require.Equal(t, 1, l.Back().Value)

		l.Remove(l.Front())
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())

		// Тест добавления после удаления всех элементов
		l.PushBack(2)
		require.Equal(t, 1, l.Len())
		require.Equal(t, 2, l.Front().Value)
		require.Equal(t, 2, l.Back().Value)
	})

	t.Run("different types", func(t *testing.T) {
		l := NewList()

		// Тест с разными типами данных
		l.PushFront("string")
		l.PushBack(42)
		l.PushBack(true)
		l.PushBack(3.14)

		require.Equal(t, 4, l.Len())
		require.Equal(t, "string", l.Front().Value)
		require.Equal(t, 3.14, l.Back().Value)
	})

	t.Run("move operations", func(t *testing.T) {
		l := NewList()

		// Тест операций перемещения
		l.PushBack(1)
		l.PushBack(2)
		l.PushBack(3)
		l.PushBack(4)

		// Перемещение среднего элемента в начало
		middle := l.Front().Next
		l.MoveToFront(middle)
		require.Equal(t, 2, l.Front().Value)

		// Перемещение последнего элемента в начало
		l.MoveToFront(l.Back())
		require.Equal(t, 4, l.Front().Value)

		// Проверка порядка элементов
		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{4, 2, 1, 3}, elems)
	})

	t.Run("remove operations", func(t *testing.T) {
		l := NewList()

		// Тест операций удаления
		l.PushBack(1)
		l.PushBack(2)
		l.PushBack(3)
		l.PushBack(4)

		// Удаление первого элемента
		l.Remove(l.Front())
		require.Equal(t, 2, l.Front().Value)

		// Удаление последнего элемента
		l.Remove(l.Back())
		require.Equal(t, 3, l.Back().Value)

		// Удаление всех элементов
		l.Remove(l.Front())
		l.Remove(l.Front())
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})
}
