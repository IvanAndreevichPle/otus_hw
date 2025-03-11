package hw02unpackstring

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnpack(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "a4bc2d5e", expected: "aaaabccddddde"},
		{input: "abccd", expected: "abccd"},
		{input: "", expected: ""},
		{input: "aaa0b", expected: "aab"},
		{input: "🙃0", expected: ""},
		{input: "aaф0b", expected: "aab"},
		// uncomment if task with asterisk completed
		{input: `qwe\4\5`, expected: `qwe45`},
		{input: `qwe\45`, expected: `qwe44444`},
		{input: `qwe\\5`, expected: `qwe\\\\\`},
		{input: `qwe\\\3`, expected: `qwe\3`},
		{input: "d\n5abc", expected: "d\n\n\n\n\nabc"},
		{input: "a-4b", expected: "a----b"},
		{input: "😊2🌍3", expected: "😊😊🌍🌍🌍"},
		{input: "a 2b 3", expected: "a  b   "},
		{input: "@3!2", expected: "@@@!!"},
		{input: `\45`, expected: "44444"},
		{input: `a\0b`, expected: "a0b"},
		{input: "a9b", expected: "aaaaaaaaab"},
		{input: `a\2b2c\3d3`, expected: "a2bbc3ddd"},
		{input: "中3国", expected: "中中中国"},
		{input: "日2本", expected: "日日本"},
		{input: "韩1国", expected: "韩国"},
		{input: `中\3国`, expected: `中3国`},
		{input: `本日\\5国`, expected: `本日\\\\\国`},
		{input: `韩\1国`, expected: `韩1国`},
		{input: "中0国", expected: "国"},
		{input: "€2£3", expected: "€€£££"},
		{input: "₹4₺5", expected: "₹₹₹₹₺₺₺₺₺"},
		{input: "α2β3", expected: "ααβββ"},
		{input: "ℝ2ℕ3", expected: "ℝℝℕℕℕ"},
		{input: "п2р3", expected: "ппррр"},
		{input: "м4а5", expected: "ммммааааа"},
		{input: `с\2т3`, expected: "с2ттт"},
		{input: "к0л", expected: "л"},
		{input: "я1й", expected: "яй"},
		{input: "ю2я3", expected: "ююяяя"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			result, err := Unpack(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestUnpackInvalidString(t *testing.T) {
	invalidStrings := []string{"3abc", "45", "aaa10b", "abcdefg10", "1"}
	for _, tc := range invalidStrings {
		tc := tc
		t.Run(tc, func(t *testing.T) {
			_, err := Unpack(tc)
			require.Truef(t, errors.Is(err, ErrInvalidString), "actual error %q", err)
		})
	}
}
