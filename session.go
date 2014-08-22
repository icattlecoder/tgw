package tgw

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"net/http"
	"sync"
	"time"
)

type SessionStoreInterface interface {
	Get(sid string, key string, val interface{}) error
	Set(sid string, key string, val interface{}) error
	Clear(sid string, key string)
	SetString(sid string, key string, val string) error
	GetString(sid string, key string) (string, error)
}

type SessionInterface interface {
	Get(key string, val interface{}) error
	Set(key string, val interface{}) error
	SetString(key string, val string) error
	GetString(key string) (string, error)
	Clear(key string)
	Id() string
}

var (
	SESSION_NAME = "TGW_SESSION_ID"
	sidMux       = sync.RWMutex{}
)

//==================================================

type SimpleSession struct {
	id    string
	rw    http.ResponseWriter
	req   *http.Request
	store SessionStoreInterface
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

func NewSimpleSession(rw http.ResponseWriter, req *http.Request, store SessionStoreInterface) (session *SimpleSession) {

	session = &SimpleSession{
		rw:    rw,
		req:   req,
		store: store,
	}

	coki, err := req.Cookie(SESSION_NAME)
	if err == nil && coki != nil {
		session.id = coki.Value
	} else {
		session.id = getSid()
		session.Flush()
		// (*session.value)[session.id] = make(d)
	}
	return
}

func getSid() string {

	sidMux.Lock()
	defer sidMux.Unlock()
	now := time.Now().UnixNano()
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(now))
	hash := sha1.New()
	hash.Write(b)
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
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

func (s *SimpleSession) Id() string {
	return s.id
}

func (s *SimpleSession) Clear(key string) {
	s.store.Clear(s.id, key)
}

func (s *SimpleSession) Get(key string, val interface{}) (err error) {
	return s.store.Get(s.id, key, val)
}

func (s *SimpleSession) Set(key string, val interface{}) (err error) {
	return s.store.Set(s.id, key, val)
}

func (s *SimpleSession) SetString(key string, val string) (err error) {
	return s.store.SetString(s.id, key, val)
}

func (s *SimpleSession) GetString(key string) (string, error) {
	return s.store.GetString(s.id, key)
}

func (s *SimpleSession) Flush() {
	coki := NewCookie(SESSION_NAME, s.id)
	http.SetCookie(s.rw, coki)
}
