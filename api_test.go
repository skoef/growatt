package growatt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{"foo", "acbd18db4cc2f85cedef654fccc4a4d8"},
		{"bar", "37b51d194a7513e45b56f6524f2d51f2"},
		{"foobar", "3858f62230ac3c915f30cc664312c63f"},
		{"barfoo", "96948aad3fcae8ccc8a35c9b5958cd89"},
		{"growatt", "6c649d2d285d62d3c3c6182ed6863920"},
		{"solarpower", "8859a64a777a2488b958ab6e2158f4b6"},
		{"greenenergyisthefuture", "36cb2355b63f3513a318ce7d23e726a3"},
	}

	for _, test := range tests {
		assert.Equal(t, test.output, hashPassword(test.input))
	}
}

func TestGetAPIURL(t *testing.T) {
	tests := []struct {
		path   string
		query  string
		output string
	}{
		{"/foo", "foo=bar", "https://server.growatt.com/foo?foo=bar"},
		{"foo", "", "https://server.growatt.com/foo"},
	}

	for _, test := range tests {
		assert.Equal(t, test.output, getAPIURL(test.path, test.query))
	}
}

func TestAPIIsLoggedIn(t *testing.T) {
	a := API{}
	assert.Equal(t, a.isLoggedIn(), false)

	// only serverID set
	a.serverID = "foo"
	assert.Equal(t, a.isLoggedIn(), false)

	// only sessionID set
	a.serverID = ""
	a.sessionID = "bar"
	assert.Equal(t, a.isLoggedIn(), false)

	// both set
	a.serverID = "foo"
	assert.Equal(t, a.isLoggedIn(), true)
}

func TestNewAPI(t *testing.T) {
	a := NewAPI("foo", "bar")

	assert.Equal(t, "foo", a.username)
	assert.Equal(t, "bar", a.password)
}
