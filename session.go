package tgw

import (
	"errors"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

type SessionInterface interface {
	Get(key string) (interface{}, error)
	Set(key string, val interface{}) error
}

//==================================================
type d map[string]interface{}
type D map[string]d

var (
	SESSION_NAME = "TGW_SESSION_ID"
	SESSION_ID   = int32(1)
)

//==================================================

type SimpleSession struct {
	id    string
	rw    http.ResponseWriter
	req   *http.Request
	value *D
}

// Options --------------------------------------------------------------------

// Options stores configuration for a session or session store.
//
// Fields are a subset of http.Cookie fields.
type Options struct {
	Path   string
	Domain string
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	MaxAge   int
	Secure   bool
	HttpOnly bool
}

var DefaultSessionOptions = &Options{
	Path:     "/",
	MaxAge:   86400 * 7,
	HttpOnly: true,
}

func NewSimpleSession(rw http.ResponseWriter, req *http.Request, data *D) (session *SimpleSession) {

	session = &SimpleSession{
		rw:    rw,
		req:   req,
		value: data,
	}

	coki, err := req.Cookie(SESSION_NAME)
	if err == nil && coki != nil {
		return &SimpleSession{
			rw:    rw,
			req:   req,
			id:    coki.Value,
			value: data,
		}
	} else {
		session.id = getSid()
		session.Flush()
	}

	if (*session.value)[coki.Value] == nil {
		(*session.value)[coki.Value] = make(d)
	}
	return
}

func getSid() string {
	atomic.AddInt32(&SESSION_ID, 1)
	return strconv.Itoa(int(SESSION_ID))
}

// NewCookie returns an http.Cookie with the options set. It also sets
// the Expires field calculated based on the MaxAge value, for Internet
// Explorer compatibility.
func NewCookie(name, value string) *http.Cookie {
	options := DefaultSessionOptions
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
	}
	if options.MaxAge > 0 {
		d := time.Duration(options.MaxAge) * time.Second
		cookie.Expires = time.Now().Add(d)
	} else if options.MaxAge < 0 {
		// Set it to the past to expire now.
		cookie.Expires = time.Unix(1, 0)
	}
	return cookie
}

func (s *SimpleSession) Get(key string) (val interface{}, err error) {

	ma, ok := (*s.value)[s.id]
	if !ok {
		(*s.value)[s.id] = make(d)
		err = errors.New("SimpleSession.Get error : No such SESSION_ID " + s.id)
		return
	}

	if _, ok := ma[key]; ok {
		val = ma[key]
	} else {
		err = errors.New("SimpleSession.Get error : No such Key " + key)
	}
	return
}

func (s *SimpleSession) Set(key string, val interface{}) (err error) {
	ma, ok := (*s.value)[s.id]
	if !ok {
		ma = make(d)
	}
	ma[key] = val
	return
}

func (s *SimpleSession) Flush() {
	coki := NewCookie(SESSION_NAME, s.id)
	http.SetCookie(s.rw, coki)
}
