//go:build !bench
// +build !bench

package hw10programoptimization

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDomainStat(t *testing.T) {
	data := `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`

	t.Run("find 'com'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{
			"browsecat.com": 2,
			"linktype.com":  1,
		}, result)
	})

	t.Run("find 'gov'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "gov")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"browsedrive.gov": 1}, result)
	})

	t.Run("find 'unknown'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "unknown")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})
}

// Пустой ввод
func TestGetDomainStat_EmptyInput(t *testing.T) {
	result, err := GetDomainStat(bytes.NewBufferString(""), "com")
	require.NoError(t, err)
	require.Equal(t, DomainStat{}, result)
}

// Строка без поля Email
func TestGetDomainStat_NoEmailField(t *testing.T) {
	data := `{"Id":1,"Name":"Test User"}`
	result, err := GetDomainStat(bytes.NewBufferString(data), "com")
	require.NoError(t, err)
	require.Equal(t, DomainStat{}, result)
}

// Некорректный JSON
func TestGetDomainStat_InvalidJSON(t *testing.T) {
	data := `{"Id":1,"Email":"user@site.com"`
	result, err := GetDomainStat(bytes.NewBufferString(data), "com")
	require.NoError(t, err)
	require.Equal(t, DomainStat{"site.com": 1}, result)
}

// Email без символа @
func TestGetDomainStat_EmailNoAt(t *testing.T) {
	data := `{"Email":"invalidemail.com"}`
	result, err := GetDomainStat(bytes.NewBufferString(data), "com")
	require.NoError(t, err)
	require.Equal(t, DomainStat{}, result)
}

// Email с верхним регистром в домене
func TestGetDomainStat_EmailUpperCaseDomain(t *testing.T) {
	data := `{"Email":"user@Example.Com"}`
	result, err := GetDomainStat(bytes.NewBufferString(data), "com")
	require.NoError(t, err)
	require.Equal(t, DomainStat{"example.com": 1}, result)
}

// Несколько строк, смешанные валидные и невалидные
func TestGetDomainStat_MixedValidInvalid(t *testing.T) {
	data := `
{"Email":"user1@site.com"}
{"Id":2,"Name":"NoEmail"}
{"Email":"user2@site.com"}
invalid_json
{"Email":"user3@site.org"}
`
	result, err := GetDomainStat(bytes.NewBufferString(data), "com")
	require.NoError(t, err)
	require.Equal(t, DomainStat{"site.com": 2}, result)
}
