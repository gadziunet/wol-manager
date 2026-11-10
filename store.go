package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

var (
	errSetupDone = errors.New("setup already done")
	errNotFound  = errors.New("not found")
	errDupMAC    = errors.New("duplicate mac")
	errTooMany   = errors.New("too many devices")
)

const maxSessionsPerUser = 20

type User struct {
	Name    string    `json:"name"`
	Hash    string    `json:"hash"`
	Created time.Time `json:"created"`
}

type Session struct {
	TokenHash string    `json:"tokenHash"`
	User      string    `json:"user"`
	Created   time.Time `json:"created"`
	Expires   time.Time `json:"expires"`
}

type state struct {
	Version  int       `json:"version"`
	Users    []User    `json:"users"`
	Sessions []Session `json:"sessions"`
	Devices  []Device  `json:"devices"`
}

type Store struct {
	mu   sync.Mutex
	path string
	st   state
}

func openStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	s := &Store{path: filepath.Join(dir, "state.json")}
	b, err := os.ReadFile(s.path)
	switch {
	case err == nil:
		if err := json.Unmarshal(b, &s.st); err != nil {
			return nil, fmt.Errorf("%s is not valid JSON (fix or move it away, it will not be overwritten): %w", s.path, err)
		}
	case errors.Is(err, fs.ErrNotExist):
		s.st.Version = 1
	default:
		return nil, err
	}
	return s, nil
}

func (s *Store) write(st state) error {
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".state-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	ok := false
	defer func() {
		if !ok {
			tmp.Close()
			os.Remove(name)
		}
	}()
	if err := tmp.Chmod(0o600); err != nil {
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(name, s.path); err != nil {
		return err
	}
	ok = true
	return nil
}

func (s *Store) update(fn func(st *state) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := state{
		Version:  s.st.Version,
		Users:    slices.Clone(s.st.Users),
		Sessions: slices.Clone(s.st.Sessions),
		Devices:  slices.Clone(s.st.Devices),
	}
	if err := fn(&next); err != nil {
		return err
	}
	if err := s.write(next); err != nil {
		return err
	}
	s.st = next
	return nil
}

func newID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func (s *Store) HasUsers() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.st.Users) > 0
}

func (s *Store) CreateFirstUser(name, hash string) error {
	return s.update(func(st *state) error {
		if len(st.Users) > 0 {
			return errSetupDone
		}
		st.Users = append(st.Users, User{Name: name, Hash: hash, Created: time.Now().UTC()})
		return nil
	})
}

func (s *Store) User(name string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.st.Users {
		if u.Name == name {
			return u, true
		}
	}
	return User{}, false
}

func (s *Store) SetPassword(name, hash, keepToken string) error {
	return s.update(func(st *state) error {
		found := false
		for i := range st.Users {
			if st.Users[i].Name == name {
				st.Users[i].Hash = hash
				found = true
			}
		}
		if !found {
			return errNotFound
		}
		st.Sessions = slices.DeleteFunc(st.Sessions, func(x Session) bool {
			return x.User == name && x.TokenHash != keepToken
		})
		return nil
	})
}

func (s *Store) Devices() []Device {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.st.Devices)
}

func (s *Store) Device(id string) (Device, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, d := range s.st.Devices {
		if d.ID == id {
			return d, true
		}
	}
	return Device{}, false
}

func (s *Store) AddDevice(d Device) (Device, error) {
	d.ID = newID()
	err := s.update(func(st *state) error {
		if len(st.Devices) >= maxDevices {
			return errTooMany
		}
		for _, x := range st.Devices {
			if x.MAC == d.MAC {
				return errDupMAC
			}
		}
		st.Devices = append(st.Devices, d)
		return nil
	})
	return d, err
}

func (s *Store) UpdateDevice(d Device) error {
	return s.update(func(st *state) error {
		idx := -1
		for i, x := range st.Devices {
			if x.ID == d.ID {
				idx = i
			} else if x.MAC == d.MAC {
				return errDupMAC
			}
		}
		if idx < 0 {
			return errNotFound
		}
		if st.Devices[idx].Host == d.Host {
			d.LastIP = st.Devices[idx].LastIP
		}
		st.Devices[idx] = d
		return nil
	})
}

func (s *Store) DeleteDevice(id string) error {
	return s.update(func(st *state) error {
		n := len(st.Devices)
		st.Devices = slices.DeleteFunc(st.Devices, func(d Device) bool { return d.ID == id })
		if len(st.Devices) == n {
			return errNotFound
		}
		return nil
	})
}

func (s *Store) SetLastIP(id, ip string) error {
	return s.update(func(st *state) error {
		for i := range st.Devices {
			if st.Devices[i].ID == id {
				if st.Devices[i].LastIP == ip {
					return nil
				}
				st.Devices[i].LastIP = ip
				return nil
			}
		}
		return errNotFound
	})
}

func (s *Store) AddSession(sess Session) error {
	return s.update(func(st *state) error {
		st.Sessions = append(st.Sessions, sess)
		var mine []int
		for i, x := range st.Sessions {
			if x.User == sess.User {
				mine = append(mine, i)
			}
		}
		if extra := len(mine) - maxSessionsPerUser; extra > 0 {
			drop := map[int]bool{}
			for _, i := range mine[:extra] {
				drop[i] = true
			}
			var kept []Session
			for i, x := range st.Sessions {
				if !drop[i] {
					kept = append(kept, x)
				}
			}
			st.Sessions = kept
		}
		return nil
	})
}

func (s *Store) SessionUser(tokenHash string, now time.Time) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, x := range s.st.Sessions {
		if x.TokenHash == tokenHash && now.Before(x.Expires) {
			return x.User, true
		}
	}
	return "", false
}

func (s *Store) DeleteSession(tokenHash string) error {
	return s.update(func(st *state) error {
		st.Sessions = slices.DeleteFunc(st.Sessions, func(x Session) bool { return x.TokenHash == tokenHash })
		return nil
	})
}

func (s *Store) PurgeSessions(now time.Time) error {
	s.mu.Lock()
	expired := false
	for _, x := range s.st.Sessions {
		if !now.Before(x.Expires) {
			expired = true
			break
		}
	}
	s.mu.Unlock()
	if !expired {
		return nil
	}
	return s.update(func(st *state) error {
		st.Sessions = slices.DeleteFunc(st.Sessions, func(x Session) bool { return !now.Before(x.Expires) })
		return nil
	})
}
