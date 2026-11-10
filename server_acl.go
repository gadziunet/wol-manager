package main

import (
	"log"
	"net/http"
	"strings"
	"time"
)

func roleOf(u User) string {
	if u.IsAdmin() {
		return roleAdmin
	}
	return roleUser
}

/* ---------- groups ---------- */

type groupView struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Devices int    `json:"devices"`
}

// handleGroups lists the groups. A regular user only gets the groups that
// contain at least one device they may use, counted over those devices.
func (s *Server) handleGroups(w http.ResponseWriter, r *http.Request, a authInfo) {
	counts := map[string]int{}
	for _, d := range s.store.Devices() {
		if d.Group != "" && a.Account.canUse(d) {
			counts[d.Group]++
		}
	}
	out := []groupView{}
	for _, g := range s.store.Groups() {
		if a.Account.IsAdmin() || counts[g.ID] > 0 {
			out = append(out, groupView{ID: g.ID, Name: g.Name, Devices: counts[g.ID]})
		}
	}
	writeJSON(w, http.StatusOK, out)
}

type groupInput struct {
	Name string `json:"name"`
}

func (s *Server) handleGroupCreate(w http.ResponseWriter, r *http.Request, a authInfo) {
	var in groupInput
	if !decodeJSON(w, r, &in) {
		return
	}
	name := strings.TrimSpace(in.Name)
	if !validGroupName(name) {
		fail(w, http.StatusBadRequest, "invalid_group_name")
		return
	}
	g, err := s.store.AddGroup(name)
	if err != nil {
		s.storeError(w, err)
		return
	}
	s.events.Add(Event{Kind: "admin", Code: "group_added", User: a.User, Device: g.Name})
	writeJSON(w, http.StatusCreated, groupView{ID: g.ID, Name: g.Name})
}

func (s *Server) handleGroupRename(w http.ResponseWriter, r *http.Request, a authInfo) {
	var in groupInput
	if !decodeJSON(w, r, &in) {
		return
	}
	name := strings.TrimSpace(in.Name)
	if !validGroupName(name) {
		fail(w, http.StatusBadRequest, "invalid_group_name")
		return
	}
	if err := s.store.RenameGroup(r.PathValue("id"), name); err != nil {
		s.storeError(w, err)
		return
	}
	s.events.Add(Event{Kind: "admin", Code: "group_changed", User: a.User, Device: name})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleGroupDelete(w http.ResponseWriter, r *http.Request, a authInfo) {
	id := r.PathValue("id")
	g, _ := s.store.Group(id)
	if err := s.store.DeleteGroup(id); err != nil {
		s.storeError(w, err)
		return
	}
	s.events.Add(Event{Kind: "admin", Code: "group_deleted", User: a.User, Device: g.Name})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// handleGroupWake wakes every device of the group that the caller may use,
// each with its own settings.
func (s *Server) handleGroupWake(w http.ResponseWriter, r *http.Request, a authInfo) {
	id := r.PathValue("id")
	g, ok := s.store.Group(id)
	if !ok {
		fail(w, http.StatusNotFound, "not_found")
		return
	}
	started := 0
	for _, d := range s.store.Devices() {
		if d.Group != id || !a.Account.canUse(d) {
			continue
		}
		s.events.Add(Event{Kind: "wake", Code: "requested_group", User: a.User, DeviceID: d.ID, Device: d.Name})
		s.wakes.Start(d)
		started++
	}
	if started == 0 {
		fail(w, http.StatusNotFound, "not_found")
		return
	}
	log.Printf("wake: group %q (%d devices) requested by %s", g.Name, started, a.User)
	writeJSON(w, http.StatusAccepted, map[string]any{"started": started})
}

/* ---------- users ---------- */

type userView struct {
	Name    string    `json:"name"`
	Role    string    `json:"role"`
	Devices []string  `json:"devices"`
	Groups  []string  `json:"groups"`
	Created time.Time `json:"created"`
}

func viewUser(u User) userView {
	return userView{
		Name:    u.Name,
		Role:    roleOf(u),
		Devices: append([]string{}, u.Devices...),
		Groups:  append([]string{}, u.Groups...),
		Created: u.Created,
	}
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request, _ authInfo) {
	users := s.store.Users()
	out := make([]userView, 0, len(users))
	for _, u := range users {
		out = append(out, viewUser(u))
	}
	writeJSON(w, http.StatusOK, out)
}

func validRole(r string) bool { return r == roleAdmin || r == roleUser }

type userCreateInput struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Role     string   `json:"role"`
	Devices  []string `json:"devices"`
	Groups   []string `json:"groups"`
}

func (s *Server) handleUserCreate(w http.ResponseWriter, r *http.Request, a authInfo) {
	var in userCreateInput
	if !decodeJSON(w, r, &in) {
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
	if !validRole(in.Role) {
		fail(w, http.StatusBadRequest, "invalid_role")
		return
	}
	hash, err := hashPassword(in.Password)
	if err != nil {
		fail(w, http.StatusInternalServerError, "internal")
		return
	}
	u := User{Name: name, Hash: hash, Created: time.Now().UTC(), Role: in.Role, Devices: in.Devices, Groups: in.Groups}
	if err := s.store.AddUser(u); err != nil {
		s.storeError(w, err)
		return
	}
	saved, _ := s.store.User(name)
	s.events.Add(Event{Kind: "admin", Code: "user_added", User: a.User, Device: name})
	writeJSON(w, http.StatusCreated, viewUser(saved))
}

type userUpdateInput struct {
	Role     string   `json:"role"`
	Devices  []string `json:"devices"`
	Groups   []string `json:"groups"`
	Password string   `json:"password"`
}

func (s *Server) handleUserUpdate(w http.ResponseWriter, r *http.Request, a authInfo) {
	name := r.PathValue("name")
	var in userUpdateInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if !validRole(in.Role) {
		fail(w, http.StatusBadRequest, "invalid_role")
		return
	}
	p := UserPatch{Role: in.Role, Devices: in.Devices, Groups: in.Groups}
	if in.Password != "" {
		if len(in.Password) < minPassword || len(in.Password) > maxPassword {
			fail(w, http.StatusBadRequest, "invalid_password")
			return
		}
		hash, err := hashPassword(in.Password)
		if err != nil {
			fail(w, http.StatusInternalServerError, "internal")
			return
		}
		p.Hash = hash
	}
	if err := s.store.UpdateUser(name, p); err != nil {
		s.storeError(w, err)
		return
	}
	saved, _ := s.store.User(name)
	s.events.Add(Event{Kind: "admin", Code: "user_changed", User: a.User, Device: name})
	writeJSON(w, http.StatusOK, viewUser(saved))
}

func (s *Server) handleUserDelete(w http.ResponseWriter, r *http.Request, a authInfo) {
	name := r.PathValue("name")
	if name == a.User {
		fail(w, http.StatusConflict, "cannot_delete_self")
		return
	}
	if err := s.store.DeleteUser(name); err != nil {
		s.storeError(w, err)
		return
	}
	s.events.Add(Event{Kind: "admin", Code: "user_deleted", User: a.User, Device: name})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

/* ---------- logs ---------- */

// handleLogs returns the newest events first. Administrators see everything,
// a regular user only sees wake events of devices they may use, without the
// packet targets.
func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request, a authInfo) {
	var out []Event
	if a.Account.IsAdmin() {
		out = s.events.Snapshot(func(Event) bool { return true })
	} else {
		usable := map[string]bool{}
		for _, d := range s.store.Devices() {
			if a.Account.canUse(d) {
				usable[d.ID] = true
			}
		}
		out = s.events.Snapshot(func(e Event) bool { return e.Kind == "wake" && usable[e.DeviceID] })
		for i := range out {
			out[i].Results = nil
			out[i].User = ""
			out[i].IP = ""
		}
	}
	writeJSON(w, http.StatusOK, out)
}
