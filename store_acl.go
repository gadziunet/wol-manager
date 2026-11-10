package main

import (
	"slices"
	"strings"
	"unicode/utf8"
)

func hasGroup(st *state, id string) bool {
	return slices.ContainsFunc(st.Groups, func(g Group) bool { return g.ID == id })
}

func hasDevice(st *state, id string) bool {
	return slices.ContainsFunc(st.Devices, func(d Device) bool { return d.ID == id })
}

func countAdmins(users []User) int {
	n := 0
	for _, u := range users {
		if u.IsAdmin() {
			n++
		}
	}
	return n
}

// cleanRefs removes duplicates and refuses ids that do not exist.
func cleanRefs(ids []string, ok func(string) bool) ([]string, error) {
	var out []string
	for _, id := range ids {
		if !ok(id) {
			return nil, errBadRef
		}
		if !slices.Contains(out, id) {
			out = append(out, id)
		}
	}
	return out, nil
}

func (s *Store) Users() []User {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.st.Users)
}

func (s *Store) AddUser(u User) error {
	return s.update(func(st *state) error {
		if len(st.Users) >= maxUsers {
			return errLimit
		}
		if slices.ContainsFunc(st.Users, func(x User) bool { return x.Name == u.Name }) {
			return errDupUser
		}
		var err error
		if u.Devices, err = cleanRefs(u.Devices, func(id string) bool { return hasDevice(st, id) }); err != nil {
			return err
		}
		if u.Groups, err = cleanRefs(u.Groups, func(id string) bool { return hasGroup(st, id) }); err != nil {
			return err
		}
		st.Users = append(st.Users, u)
		return nil
	})
}

// UserPatch changes an account. Hash, when not empty, replaces the password
// and signs the user out everywhere.
type UserPatch struct {
	Role    string
	Devices []string
	Groups  []string
	Hash    string
}

func (s *Store) UpdateUser(name string, p UserPatch) error {
	return s.update(func(st *state) error {
		idx := slices.IndexFunc(st.Users, func(x User) bool { return x.Name == name })
		if idx < 0 {
			return errNotFound
		}
		u := st.Users[idx]
		var err error
		if u.Devices, err = cleanRefs(p.Devices, func(id string) bool { return hasDevice(st, id) }); err != nil {
			return err
		}
		if u.Groups, err = cleanRefs(p.Groups, func(id string) bool { return hasGroup(st, id) }); err != nil {
			return err
		}
		u.Role = p.Role
		if p.Hash != "" {
			u.Hash = p.Hash
		}
		st.Users = slices.Clone(st.Users)
		st.Users[idx] = u
		if countAdmins(st.Users) == 0 {
			return errLastAdmin
		}
		if p.Hash != "" {
			st.Sessions = slices.DeleteFunc(slices.Clone(st.Sessions), func(x Session) bool { return x.User == name })
		}
		return nil
	})
}

func (s *Store) DeleteUser(name string) error {
	return s.update(func(st *state) error {
		n := len(st.Users)
		st.Users = slices.DeleteFunc(slices.Clone(st.Users), func(x User) bool { return x.Name == name })
		if len(st.Users) == n {
			return errNotFound
		}
		if countAdmins(st.Users) == 0 {
			return errLastAdmin
		}
		st.Sessions = slices.DeleteFunc(slices.Clone(st.Sessions), func(x Session) bool { return x.User == name })
		return nil
	})
}

func (s *Store) Groups() []Group {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.st.Groups)
}

func (s *Store) Group(id string) (Group, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, g := range s.st.Groups {
		if g.ID == id {
			return g, true
		}
	}
	return Group{}, false
}

func groupNameTaken(st *state, name, exceptID string) bool {
	return slices.ContainsFunc(st.Groups, func(g Group) bool {
		return g.ID != exceptID && strings.EqualFold(g.Name, name)
	})
}

func validGroupName(name string) bool {
	n := utf8.RuneCountInString(name)
	return n >= 1 && n <= 40
}

func (s *Store) AddGroup(name string) (Group, error) {
	g := Group{ID: newID(), Name: name}
	err := s.update(func(st *state) error {
		if len(st.Groups) >= maxGroups {
			return errLimit
		}
		if groupNameTaken(st, name, "") {
			return errDupGroup
		}
		st.Groups = append(st.Groups, g)
		return nil
	})
	return g, err
}

func (s *Store) RenameGroup(id, name string) error {
	return s.update(func(st *state) error {
		idx := slices.IndexFunc(st.Groups, func(g Group) bool { return g.ID == id })
		if idx < 0 {
			return errNotFound
		}
		if groupNameTaken(st, name, id) {
			return errDupGroup
		}
		st.Groups = slices.Clone(st.Groups)
		st.Groups[idx].Name = name
		return nil
	})
}

// DeleteGroup removes the group. Its devices stay, without a group, and the
// group disappears from every account.
func (s *Store) DeleteGroup(id string) error {
	return s.update(func(st *state) error {
		n := len(st.Groups)
		st.Groups = slices.DeleteFunc(slices.Clone(st.Groups), func(g Group) bool { return g.ID == id })
		if len(st.Groups) == n {
			return errNotFound
		}
		for i := range st.Devices {
			if st.Devices[i].Group == id {
				st.Devices[i].Group = ""
			}
		}
		for i := range st.Users {
			if slices.Contains(st.Users[i].Groups, id) {
				st.Users[i].Groups = slices.DeleteFunc(slices.Clone(st.Users[i].Groups), func(x string) bool { return x == id })
			}
		}
		return nil
	})
}
