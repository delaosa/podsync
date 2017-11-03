package feeds

import (
	"net/url"
	"strings"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/pkg/errors"
)

func parseURL(link string) (*api.Feed, error) {
	if !strings.HasPrefix(link, "http") {
		link = "https://" + link
	}

	parsed, err := url.Parse(link)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse url: %s", link)
		return nil, err
	}

	feed := &api.Feed{}

	if strings.HasSuffix(parsed.Host, "youtube.com") {
		kind, id, err := parseYoutubeURL(parsed)
		if err != nil {
			return nil, err
		}

		feed.Provider = api.ProviderYoutube
		feed.LinkType = kind
		feed.ItemId = id

		return feed, nil
	}

	if strings.HasSuffix(parsed.Host, "vimeo.com") {
		kind, id, err := parseVimeoURL(parsed)
		if err != nil {
			return nil, err
		}

		feed.Provider = api.ProviderVimeo
		feed.LinkType = kind
		feed.ItemId = id

		return feed, nil
	}

	return nil, errors.New("unsupported URL host")
}

func parseYoutubeURL(parsed *url.URL) (kind api.LinkType, id string, err error) {
	path := parsed.EscapedPath()

	// https://www.youtube.com/playlist?list=PLCB9F975ECF01953C
	// https://www.youtube.com/watch?v=rbCbho7aLYw&list=PLMpEfaKcGjpWEgNtdnsvLX6LzQL0UC0EM
	if strings.HasPrefix(path, "/playlist") || strings.HasPrefix(path, "/watch") {
		kind = api.LinkTypePlaylist

		id = parsed.Query().Get("list")
		if id != "" {
			return
		}

		err = errors.New("invalid playlist link")
		return
	}

	// - https://www.youtube.com/channel/UC5XPnUk8Vvv_pWslhwom6Og
	// - https://www.youtube.com/channel/UCrlakW-ewUT8sOod6Wmzyow/videos
	if strings.HasPrefix(path, "/channel") {
		kind = api.LinkTypeChannel
		parts := strings.Split(parsed.EscapedPath(), "/")
		if len(parts) <= 2 {
			err = errors.New("invalid youtube channel link")
			return
		}

		id = parts[2]
		if id == "" {
			err = errors.New("invalid id")
		}

		return
	}

	// - https://www.youtube.com/user/fxigr1
	if strings.HasPrefix(path, "/user") {
		kind = api.LinkTypeUser

		parts := strings.Split(parsed.EscapedPath(), "/")
		if len(parts) <= 2 {
			err = errors.New("invalid user link")
			return
		}

		id = parts[2]
		if id == "" {
			err = errors.New("invalid id")
		}

		return
	}

	err = errors.New("unsupported link format")
	return
}

func parseVimeoURL(parsed *url.URL) (kind api.LinkType, id string, err error) {
	parts := strings.Split(parsed.EscapedPath(), "/")

	if len(parts) <= 1 {
		err = errors.New("invalid vimeo link path")
		return
	}

	if parts[1] == "groups" {
		kind = api.LinkTypeGroup
	} else if parts[1] == "channels" {
		kind = api.LinkTypeChannel
	} else {
		kind = api.LinkTypeUser
	}

	if kind == api.LinkTypeGroup || kind == api.LinkTypeChannel {
		if len(parts) <= 2 {
			err = errors.New("invalid channel link")
			return
		}

		id = parts[2]
		if id == "" {
			err = errors.New("invalid id")
		}

		return
	}

	if kind == api.LinkTypeUser {
		id = parts[1]
		if id == "" {
			err = errors.New("invalid id")
		}

		return
	}

	err = errors.New("unsupported link format")
	return
}
