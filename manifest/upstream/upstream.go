package upstream

import "net/url"

var AvailableUpstreams map[string]Upstream

type Upstream interface {
	Auth(path string) (result Guard, err error)
	HaveWriteAccess(token string, url url.URL) (hasAccess bool, err error)
	Fork(token string, url url.URL) (result string, err error)
	OpenPR(opts PrOpts) (err error)
	SearchPrByBranch(url url.URL, token, branchName string) (err error)
}

type Guard interface {
	GetAccessToken() (token string)
	GetCode() (code string)
	GetInterval() (interval int)
	GetExpireInterval() (interval int)
	CheckStatus() (status string, err error)
	GetUser() (username string, err error)
}

type PrOpts struct {
	Origin     url.URL
	Fork       url.URL
	Token      string
	MainBranch string
	PrBranch   string
	PrTitle    string
	PrBody     string
}

func registerUpstream(upstream Upstream, host string) {
	if AvailableUpstreams == nil {
		AvailableUpstreams = make(map[string]Upstream)
	}
	AvailableUpstreams[host] = upstream
}
