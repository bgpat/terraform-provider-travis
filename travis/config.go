package travis

import (
	"github.com/shuheiktgw/go-travis"
)

type Config struct {
	URL   string
	Org   bool
	Com   bool
	Token string
}

func (c *Config) Client() *Client {
	if c.Org {
		c.URL = travis.ApiOrgUrl
	}
	if c.Com {
		c.URL = travis.ApiComUrl
	}
	return NewClient(c.URL, c.Token)
}
