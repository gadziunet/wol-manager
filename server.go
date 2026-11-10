package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed web
var webFS embed.FS

const (
	sessionTTL  = 30 * 24 * time.Hour
	maxBodySize = 64 << 10
)

type Server struct {
	cfg     Config
	store   *Store
	wakes   *Manager
	tracker *Tracker
	lim     *limiter

	mu         sync.Mutex
	setupToken string
	dummyOnce  sync.Once
	dummyHash  string
}

func newServer(cfg Config, st *Store, wakes *Manager, tracker *Tracker) *Server {
	s := &Server{
		cfg:     cfg,
		store:   st,
		wakes:   wakes,
		tracker: tracker,
		lim:     newLimiter(5, 15*time.Minute, 15*time.Minute),
	}
	if !st.HasUsers() {
		s.setupToken = newToken()[:22]
	}
	return s
}

func (s *Server) SetupToken() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.setupToken
}

func (s *Server) janitor(ctx context.Context) {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			if err := s.store.PurgeSessions(now); err != nil {
				log.Printf("purging sessions: %v", err)
			}
			s.lim.purge(now)
		}
	}
}

func (s *Server) cookieName() string {
	if s.cfg.CookieSecure {
		return "__Host-wol"
	}
	return "wol_session"
}

func (s *Server) clientIP(r *http.Request) string {
	if s.cfg.TrustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if ip := strings.TrimSpace(parts[len(parts)-1]); net.ParseIP(ip) != nil {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, code int, errCode string) {
	writeJSON(w, code, map[string]string{"error": errCode})
}

func failField(w http.ResponseWriter, err error) {
	var fe *fieldError
	if errors.As(err, &fe) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": fe.Code, "field": fe.Field})
		return
	}
	fail(w, http.StatusBadRequest, "bad_request")
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		fail(w, http.StatusBadRequest, "bad_request")
		return false
	}
	return true
}

func (s *Server) sameOrigin(r *http.Request) bool {
	if v := r.Header.Get("Sec-Fetch-Site"); v != "" {
		return v == "same-origin" || v == "none"
	}
	if o := r.Header.Get("Origin"); o != "" {
		u, err := url.Parse(o)
		if err != nil {
			return false
		}
		if strings.EqualFold(u.Host, r.Host) {
			return true
		}
		if s.cfg.TrustProxy {
			return strings.EqualFold(u.Host, r.Header.Get("X-Forwarded-Host"))
		}
		return false
	}
	return true
}

func (s *Server) secure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; form-action 'self'; base-uri 'none'; frame-ancestors 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		if r.TLS != nil || s.cfg.CookieSecure {
			h.Set("Strict-Transport-Security", "max-age=31536000")
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			h.Set("Cache-Control", "no-store")
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
			default:
				if !s.sameOrigin(r) {
					fail(w, http.StatusForbidden, "forbidden_origin")
					return
				}
				mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if mt != "application/json" {
					fail(w, http.StatusUnsupportedMediaType, "bad_content_type")
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

type authInfo struct {
	User      string
	TokenHash string
}

func (s *Server) session(r *http.Request) (authInfo, bool) {
	c, err := r.Cookie(s.cookieName())
	if err != nil || c.Value == "" {
		return authInfo{}, false
	}
	h := tokenHash(c.Value)
	user, ok := s.store.SessionUser(h, time.Now())
	if !ok {
		return authInfo{}, false
	}
	return authInfo{User: user, TokenHash: h}, true
}

func (s *Server) auth(next func(http.ResponseWriter, *http.Request, authInfo)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a, ok := s.session(r)
		if !ok {
			fail(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r, a)
	}
}

func (s *Server) startSession(w http.ResponseWriter, user string) error {
	token := newToken()
	now := time.Now().UTC()
	if err := s.store.AddSession(Session{TokenHash: tokenHash(token), User: user, Created: now, Expires: now.Add(sessionTTL)}); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName(),
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
	return nil
}

func (s *Server) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName(),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("GET /api/session", s.handleSession)
	mux.HandleFunc("POST /api/setup", s.handleSetup)
	mux.HandleFunc("POST /api/login", s.handleLogin)
	mux.HandleFunc("POST /api/logout", s.auth(s.handleLogout))
	mux.HandleFunc("POST /api/password", s.auth(s.handlePassword))
	mux.HandleFunc("GET /api/interfaces", s.auth(s.handleInterfaces))
	mux.HandleFunc("GET /api/devices", s.auth(s.handleDevices))
	mux.HandleFunc("POST /api/devices", s.auth(s.handleDeviceCreate))
	mux.HandleFunc("PUT /api/devices/{id}", s.auth(s.handleDeviceUpdate))
	mux.HandleFunc("DELETE /api/devices/{id}", s.auth(s.handleDeviceDelete))
	mux.HandleFunc("POST /api/devices/{id}/wake", s.auth(s.handleWake))
	mux.HandleFunc("POST /api/devices/{id}/wake/cancel", s.auth(s.handleWakeCancel))

	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		panic(err)
	}
	files := http.FileServerFS(sub)
	noCache := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-cache")
			h.ServeHTTP(w, r)
		})
	}
	mux.Handle("GET /{$}", noCache(files))
	mux.Handle("GET /app.js", noCache(files))
	mux.Handle("GET /app.css", noCache(files))
	return s.secure(mux)
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	a, ok := s.session(r)
	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": ok,
		"user":          a.User,
		"setupRequired": !s.store.HasUsers(),
		"version":       version,
	})
}

type credentials struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	ip := s.clientIP(r)
	if ok, wait := s.lim.allowed(ip, time.Now()); !ok {
		w.Header().Set("Retry-After", strconv.Itoa(int(wait.Seconds())+1))
		fail(w, http.StatusTooManyRequests, "rate_limited")
		return
	}
	var in credentials
	if !decodeJSON(w, r, &in) {
		return
	}
	expected := s.SetupToken()
	if expected == "" || s.store.HasUsers() {
		fail(w, http.StatusConflict, "setup_done")
		return
	}
	if !tokensEqual(strings.TrimSpace(in.Token), expected) {
		s.lim.fail(ip, time.Now())
		log.Printf("setup: wrong setup token from %s", ip)
		fail(w, http.StatusForbidden, "bad_setup_token")
		return
	}
	name := strings.ToLower(strings.TrimSpace(in.Username))
	if !userRe.MatchString(name) {
		fail(w, http.StatusBadRequest, "invalid_username")
		return
	}
	if len(in.Password) < minPassword || len(in.Password) > maxPassword {
		fail(w, http.StatusBadRequest, "invalid_password")
		return
	}
	hash, err := hashPassword(in.Password)
	if err != nil {
		fail(w, http.StatusInternalServerError, "internal")
		return
	}
	if err := s.store.CreateFirstUser(name, hash); err != nil {
		if errors.Is(err, errSetupDone) {
			fail(w, http.StatusConflict, "setup_done")
			return
		}
		log.Printf("setup: %v", err)
		fail(w, http.StatusInternalServerError, "internal")
		return
	}
	s.mu.Lock()
	s.setupToken = ""
	s.mu.Unlock()
	s.lim.reset(ip)
	if err := s.startSession(w, name); err != nil {
		log.Printf("setup: session: %v", err)
		fail(w, http.StatusInternalServerError, "internal")
		return
	}
	log.Printf("setup: account %q created", name)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": name})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	ip := s.clientIP(r)
	if ok, wait := s.lim.allowed(ip, time.Now()); !ok {
		w.Header().Set("Retry-After", strconv.Itoa(int(wait.Seconds())+1))
		fail(w, http.StatusTooManyRequests, "rate_limited")
		return
	}
	var in credentials
	if !decodeJSON(w, r, &in) {
		return
	}
	name := strings.ToLower(strings.TrimSpace(in.Username))
	user, found := s.store.User(name)
	stored := user.Hash
	if !found || len(in.Password) > maxPassword {
		s.dummyOnce.Do(func() { s.dummyHash, _ = hashPassword("dummy-password-for-timing") })
		stored = s.dummyHash
		found = false
	}
	valid := verifyPassword(stored, in.Password) && found
	if !valid {
		s.lim.fail(ip, time.Now())
		log.Printf("login: failed attempt from %s", ip)
		fail(w, http.StatusUnauthorized, "bad_credentials")
		return
	}
	s.lim.reset(ip)
	if err := s.startSession(w, user.Name); err != nil {
		log.Printf("login: session: %v", err)
		fail(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": user.Name})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request, a authInfo) {
	_ = s.store.DeleteSession(a.TokenHash)
	s.clearCookie(w)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handlePassword(w http.ResponseWriter, r *http.Request, a authInfo) {
	ip := s.clientIP(r)
	if ok, wait := s.lim.allowed(ip, time.Now()); !ok {
		w.Header().Set("Retry-After", strconv.Itoa(int(wait.Seconds())+1))
		fail(w, http.StatusTooManyRequests, "rate_limited")
		return
	}
	var in struct {
		Current string `json:"current"`
		New     string `json:"new"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	user, ok := s.store.User(a.User)
	if !ok || !verifyPassword(user.Hash, in.Current) {
		s.lim.fail(ip, time.Now())
		fail(w, http.StatusForbidden, "bad_credentials")
		return
	}
	if len(in.New) < minPassword || len(in.New) > maxPassword {
		fail(w, http.StatusBadRequest, "invalid_password")
		return
	}
	hash, err := hashPassword(in.New)
	if err != nil {
		fail(w, http.StatusInternalServerError, "internal")
		return
	}
	if err := s.store.SetPassword(a.User, hash, a.TokenHash); err != nil {
		fail(w, http.StatusInternalServerError, "internal")
		return
	}
	s.lim.reset(ip)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleInterfaces(w http.ResponseWriter, r *http.Request, _ authInfo) {
	type view struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	}
	out := []view{}
	for _, ifc := range localInterfaces("") {
		ones, _ := ifc.Mask.Size()
		out = append(out, view{Name: ifc.Name, Address: ifc.IP.String() + "/" + strconv.Itoa(ones)})
	}
	writeJSON(w, http.StatusOK, out)
}

type deviceView struct {
	Device
	Status StatusView `json:"status"`
	Wake   WakeState  `json:"wake"`
}

func (s *Server) view(d Device, now time.Time) deviceView {
	if d.CheckPorts == nil {
		d.CheckPorts = []int{}
	}
	return deviceView{Device: d, Status: s.tracker.View(d.ID, now), Wake: s.wakes.State(d.ID)}
}

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request, _ authInfo) {
	now := time.Now()
	devs := s.store.Devices()
	out := make([]deviceView, 0, len(devs))
	for _, d := range devs {
		out = append(out, s.view(d, now))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) storeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errDupMAC):
		fail(w, http.StatusConflict, "duplicate_mac")
	case errors.Is(err, errTooMany):
		fail(w, http.StatusConflict, "too_many_devices")
	case errors.Is(err, errNotFound):
		fail(w, http.StatusNotFound, "not_found")
	default:
		log.Printf("store: %v", err)
		fail(w, http.StatusInternalServerError, "internal")
	}
}

func (s *Server) handleDeviceCreate(w http.ResponseWriter, r *http.Request, _ authInfo) {
	var in deviceInput
	if !decodeJSON(w, r, &in) {
		return
	}
	d, err := in.device()
	if err != nil {
		failField(w, err)
		return
	}
	d, err = s.store.AddDevice(d)
	if err != nil {
		s.storeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, s.view(d, time.Now()))
}

func (s *Server) handleDeviceUpdate(w http.ResponseWriter, r *http.Request, _ authInfo) {
	id := r.PathValue("id")
	var in deviceInput
	if !decodeJSON(w, r, &in) {
		return
	}
	d, err := in.device()
	if err != nil {
		failField(w, err)
		return
	}
	d.ID = id
	if err := s.store.UpdateDevice(d); err != nil {
		s.storeError(w, err)
		return
	}
	d, _ = s.store.Device(id)
	writeJSON(w, http.StatusOK, s.view(d, time.Now()))
}

func (s *Server) handleDeviceDelete(w http.ResponseWriter, r *http.Request, _ authInfo) {
	id := r.PathValue("id")
	if err := s.store.DeleteDevice(id); err != nil {
		s.storeError(w, err)
		return
	}
	s.wakes.Forget(id)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleWake(w http.ResponseWriter, r *http.Request, a authInfo) {
	d, ok := s.store.Device(r.PathValue("id"))
	if !ok {
		fail(w, http.StatusNotFound, "not_found")
		return
	}
	log.Printf("wake: %q requested by %s", d.Name, a.User)
	writeJSON(w, http.StatusAccepted, s.wakes.Start(d))
}

func (s *Server) handleWakeCancel(w http.ResponseWriter, r *http.Request, _ authInfo) {
	id := r.PathValue("id")
	if _, ok := s.store.Device(id); !ok {
		fail(w, http.StatusNotFound, "not_found")
		return
	}
	s.wakes.Cancel(id)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
