package media

import (
	"bytes"
	"errors"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	str "github.com/ilgooz/strings"
)

type Links []*Link

type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type Options struct {
	Page, Limit int
}

type Media struct {
	ls *Links
	r  *http.Request
	o  *Options
}

func New(r *http.Request, ls *Links, o *Options) *Media {
	me := &Media{
		r:  r,
		o:  o,
		ls: ls,
	}
	me.setOffset()
	me.setLimit()
	me.setSelfLink()
	return me
}

func (m *Media) setOffset() {
	page := m.r.FormValue("page")
	if n, err := strconv.Atoi(page); err == nil && n >= 1 {
		m.o.Page = n
	}
}

func (m *Media) setLimit() {
	limit := m.r.FormValue("limit")
	if n, err := strconv.Atoi(limit); err == nil && n >= 1 {
		m.o.Limit = n
	}
}

func (m *Media) SetPageLinks(max int) {
	if m.o.Page >= 2 {
		m.setPageLink("prev", *m.r.URL, m.o.Page-1, m.o.Limit)
	}
	if m.o.Page*m.o.Limit < max {
		m.setPageLink("next", *m.r.URL, m.o.Page+1, m.o.Limit)
	}
}

func (m *Media) setSelfLink() {
	m.SetLink("self", (*m.r.URL).String())
}

func (m *Media) setPageLink(rel string, u url.URL, page, limit int) {
	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("limit", strconv.Itoa(limit))
	u.RawQuery = encodeWithoutEscape(q)
	m.SetLink(rel, u.String())
}

func (m *Media) SetLink(rel string, u string) {
	l := &Link{
		Rel:  rel,
		Href: u,
	}
	*m.ls = append(*m.ls, l)
}

func (m *Media) GetOffset() int {
	return (m.o.Page - 1) * m.o.Limit
}

func (m *Media) GetLimit() int {
	return m.o.Limit
}

func (m *Media) GetPage() int {
	return m.o.Page
}

func (m *Media) GetTotalPageCount(max int) int {
	if max < m.GetLimit() {
		return 1
	} else {
		return int(math.Ceil(float64(max) / float64(m.GetLimit())))
	}
}

func encodeWithoutEscape(v url.Values) string {
	if v == nil {
		return ""
	}
	var buf bytes.Buffer
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		prefix := k + "="
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(prefix)
			buf.WriteString(v)
		}
	}
	return buf.String()
}

func ParseOrder(val string, allowed []string) (string, string, error) {
	v := strings.Split(val, ",")

	if len(v) != 2 {
		return "", "", errors.New("wrong ordering format")
	}

	by := v[0]
	ty := v[1]

	if !str.InSlice(by, allowed) {
		return "", "", errors.New("Unvalid ordering field")
	}

	if !str.InSlice(ty, []string{"desc", "asc"}) {
		return "", "", errors.New("Unvalid ordering type")
	}

	return by, ty, nil
}
