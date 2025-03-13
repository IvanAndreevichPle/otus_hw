package hw03frequencyanalysis

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var taskWithAsteriskIsCompleted = true

var text = `Как видите, он  спускается  по  лестнице  вслед  за  своим
	другом   Кристофером   Робином,   головой   вниз,  пересчитывая
	ступеньки собственным затылком:  бум-бум-бум.  Другого  способа
	сходить  с  лестницы  он  пока  не  знает.  Иногда ему, правда,
		кажется, что можно бы найти какой-то другой способ, если бы  он
	только   мог   на  минутку  перестать  бумкать  и  как  следует
	сосредоточиться. Но увы - сосредоточиться-то ему и некогда.
		Как бы то ни было, вот он уже спустился  и  готов  с  вами
	познакомиться.
	- Винни-Пух. Очень приятно!
		Вас,  вероятно,  удивляет, почему его так странно зовут, а
	если вы знаете английский, то вы удивитесь еще больше.
		Это необыкновенное имя подарил ему Кристофер  Робин.  Надо
	вам  сказать,  что  когда-то Кристофер Робин был знаком с одним
	лебедем на пруду, которого он звал Пухом. Для лебедя  это  было
	очень   подходящее  имя,  потому  что  если  ты  зовешь  лебедя
	громко: "Пу-ух! Пу-ух!"- а он  не  откликается,  то  ты  всегда
	можешь  сделать вид, что ты просто понарошку стрелял; а если ты
	звал его тихо, то все подумают, что ты  просто  подул  себе  на
	нос.  Лебедь  потом  куда-то делся, а имя осталось, и Кристофер
	Робин решил отдать его своему медвежонку, чтобы оно не  пропало
	зря.
		А  Винни - так звали самую лучшую, самую добрую медведицу
	в  зоологическом  саду,  которую  очень-очень  любил  Кристофер
	Робин.  А  она  очень-очень  любила  его. Ее ли назвали Винни в
	честь Пуха, или Пуха назвали в ее честь - теперь уже никто  не
	знает,  даже папа Кристофера Робина. Когда-то он знал, а теперь
	забыл.
		Словом, теперь мишку зовут Винни-Пух, и вы знаете почему.
		Иногда Винни-Пух любит вечерком во что-нибудь поиграть,  а
	иногда,  особенно  когда  папа  дома,  он больше любит тихонько
	посидеть у огня и послушать какую-нибудь интересную сказку.
		В этот вечер...`

func TestTop10(t *testing.T) {
	t.Run("no words in empty string", func(t *testing.T) {
		require.Len(t, Top10(""), 0)
	})

	t.Run("positive test", func(t *testing.T) {
		if taskWithAsteriskIsCompleted {
			expected := []string{
				"а",         // 8
				"он",        // 8
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"в",         // 4
				"его",       // 4
				"если",      // 4
				"кристофер", // 4
				"не",        // 4
			}
			require.Equal(t, expected, Top10(text))
		} else {
			expected := []string{
				"он",        // 8
				"а",         // 6
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"-",         // 4
				"Кристофер", // 4
				"если",      // 4
				"не",        // 4
				"то",        // 4
			}
			require.Equal(t, expected, Top10(text))
		}
	})

	t.Run("words with hyphens and punctuation", func(t *testing.T) {
		input := "какой-то какой-то! какойто, 'какой-то' ----"
		expected := []string{"какой-то", "----", "какойто"}
		require.Equal(t, expected, Top10(input))
	})

	t.Run("ignore single hyphen", func(t *testing.T) {
		input := "- ---"
		expected := []string{"---"}
		require.Equal(t, expected, Top10(input))
	})

	t.Run("case insensitivity", func(t *testing.T) {
		input := "Нога нога НОГА"
		expected := []string{"нога"}
		require.Equal(t, expected, Top10(input))
	})

	t.Run("internal punctuation", func(t *testing.T) {
		input := "dog,cat dog...cat dogcat"
		expected := []string{"dog,cat", "dog...cat", "dogcat"}
		require.Equal(t, expected, Top10(input))
	})

	t.Run("sorting order", func(t *testing.T) {
		input := "b b a a a c"
		expected := []string{"a", "b", "c"}
		require.Equal(t, expected, Top10(input))
	})
}

func BenchmarkTop10(b *testing.B) {
	genText := func(wordCount int, punctuationProb float32) string {
		r := rand.New(rand.NewSource(42)) // Фиксированный генератор
		var buf strings.Builder

		words := []string{"apple", "banana", "cherry", "date", "fig", "grape", "kiwi"}
		punctuations := []rune{'!', ',', '.', ';', ':', '-', '\'', '"'}

		for i := 0; i < wordCount; i++ {
			if r.Float32() < punctuationProb {
				buf.WriteRune(punctuations[r.Intn(len(punctuations))])
			}

			buf.WriteString(words[r.Intn(len(words))])

			if r.Float32() < punctuationProb {
				buf.WriteRune(punctuations[r.Intn(len(punctuations))])
			}

			buf.WriteByte(' ')
		}
		return buf.String()
	}

	benchmarks := []struct {
		name          string
		textGenerator func() string
	}{
		{
			name: "SmallText",
			textGenerator: func() string {
				return genText(100, 0.2)
			},
		},
		{
			name: "MediumText",
			textGenerator: func() string {
				return genText(10_000, 0.3)
			},
		},
		{
			name: "LargeText",
			textGenerator: func() string {
				return genText(1_000_000, 0.1)
			},
		},
		{
			name: "HyphenHeavy",
			textGenerator: func() string {
				return "foo-bar-baz " + strings.Repeat("test-test-test ", 1000)
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			testText := bm.textGenerator()
			b.ReportAllocs()
			b.ResetTimer() // Таймер после генерации данных

			for i := 0; i < b.N; i++ {
				Top10(testText)
			}

			b.ReportMetric(float64(len(testText))/1e6, "MB")
		})
	}
}

//go test -bench=. -benchmem
