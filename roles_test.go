package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"
)

// second returns a client with its own cookies talking to the same server.
func (a *testApp) second() *testApp {
	jar, _ := cookiejar.New(nil)
	return &testApp{t: a.t, srv: a.srv, ts: a.ts, client: &http.Client{Jar: jar, Timeout: 10 * time.Second}}
}

func (a *testApp) mustDo(method, path string, body any, want int) map[string]any {
	a.t.Helper()
	code, out, _ := a.do(method, path, body, nil)
	if code != want {
		a.t.Fatalf("%s %s: got %d %v, want %d", method, path, code, out, want)
	}
	return out
}

func (a *testApp) addDevice(name string, n int, group string) string {
	a.t.Helper()
	out := a.mustDo("POST", "/api/devices", map[string]any{
		"name": name, "mac": fmt.Sprintf("aa:bb:cc:dd:ee:%02x", n), "host": "127.0.0.1",
		"methods": map[string]bool{"global": false, "directed": false, "unicast": true},
		"icmp":    false, "group": group,
	}, 201)
	return out["id"].(string)
}

func (a *testApp) addGroup(name string) string {
	a.t.Helper()
	return a.mustDo("POST", "/api/groups", map[string]string{"name": name}, 201)["id"].(string)
}

func (a *testApp) login(user, pass string) {
	a.t.Helper()
	a.mustDo("POST", "/api/login", map[string]string{"username": user, "password": pass}, 200)
}

func rolesApp(t *testing.T) *testApp {
	t.Helper()
	pc, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("no UDP socket: %v", err)
	}
	t.Cleanup(func() { pc.Close() })
	a := newTestApp(t, []int{pc.LocalAddr().(*net.UDPAddr).Port})
	a.setup()
	return a
}

func ids(list []map[string]any) map[string]bool {
	m := map[string]bool{}
	for _, x := range list {
		m[x["id"].(string)] = true
	}
	return m
}

func TestRegularUserSeesAndWakesOnlyAssigned(t *testing.T) {
	a := rolesApp(t)
	g1 := a.addGroup("Room")
	d1 := a.addDevice("In group", 1, g1)
	d2 := a.addDevice("Direct", 2, "")
	d3 := a.addDevice("Hidden", 3, "")

	a.mustDo("POST", "/api/users", map[string]any{
		"username": "Bob", "password": "bobs-long-password", "role": "user",
		"devices": []string{d2}, "groups": []string{g1},
	}, 201)

	bob := a.second()
	bob.login("bob", "bobs-long-password")

	code, list := bob.doList("/api/devices")
	if code != 200 {
		t.Fatalf("devices: %d", code)
	}
	got := ids(list)
	if len(got) != 2 || !got[d1] || !got[d2] || got[d3] {
		t.Fatalf("bob sees %v, want %s and %s only", got, d1, d2)
	}
	for _, d := range list {
		if d["mac"] != "" || d["host"] != "" {
			t.Fatalf("addresses leaked to a regular user: %v", d)
		}
	}

	bob.mustDo("POST", "/api/devices", map[string]any{"name": "x", "mac": "aa:bb:cc:dd:ee:99"}, 403)
	bob.mustDo("PUT", "/api/devices/"+d2, map[string]any{"name": "x", "mac": "aa:bb:cc:dd:ee:02"}, 403)
	bob.mustDo("DELETE", "/api/devices/"+d2, nil, 403)
	bob.mustDo("GET", "/api/users", nil, 403)
	bob.mustDo("POST", "/api/users", map[string]any{"username": "eve", "password": "long-enough-pass", "role": "admin"}, 403)
	bob.mustDo("GET", "/api/interfaces", nil, 403)
	bob.mustDo("POST", "/api/groups", map[string]string{"name": "Mine"}, 403)

	bob.mustDo("POST", "/api/devices/"+d3+"/wake", nil, 404)
	bob.mustDo("POST", "/api/devices/"+d3+"/wake/cancel", nil, 404)
	bob.mustDo("POST", "/api/devices/"+d2+"/wake", nil, 202)
	bob.mustDo("POST", "/api/devices/"+d1+"/wake", nil, 202)

	code, groups := bob.doList("/api/groups")
	if code != 200 || len(groups) != 1 || groups[0]["id"] != g1 || groups[0]["devices"].(float64) != 1 {
		t.Fatalf("bob groups: %d %v", code, groups)
	}
	out := bob.mustDo("POST", "/api/groups/"+g1+"/wake", nil, 202)
	if out["started"].(float64) != 1 {
		t.Fatalf("group wake started %v, want 1", out["started"])
	}
}

func TestGroupWakeStartsEveryDeviceWithItsOwnSettings(t *testing.T) {
	a := rolesApp(t)
	g := a.addGroup("All")
	a.addDevice("One", 1, g)
	a.addDevice("Two", 2, g)
	a.addDevice("Outside", 3, "")

	out := a.mustDo("POST", "/api/groups/"+g+"/wake", nil, 202)
	if out["started"].(float64) != 2 {
		t.Fatalf("started %v, want 2", out["started"])
	}
	empty := a.addGroup("Empty")
	a.mustDo("POST", "/api/groups/"+empty+"/wake", nil, 404)
	a.mustDo("POST", "/api/groups/0123456789abcdef/wake", nil, 404)
}

func TestGroupRules(t *testing.T) {
	a := rolesApp(t)
	g := a.addGroup("Room")
	a.mustDo("POST", "/api/groups", map[string]string{"name": "room"}, 409)
	a.mustDo("POST", "/api/groups", map[string]string{"name": "  "}, 400)

	a.mustDo("POST", "/api/devices", map[string]any{"name": "x", "mac": "aa:bb:cc:dd:ee:01", "group": "0123456789abcdef"}, 400)
	a.mustDo("POST", "/api/devices", map[string]any{"name": "x", "mac": "aa:bb:cc:dd:ee:01", "group": "not-an-id"}, 400)

	d := a.addDevice("PC", 1, g)
	a.mustDo("POST", "/api/users", map[string]any{
		"username": "bob", "password": "bobs-long-password", "role": "user", "groups": []string{g},
	}, 201)
	a.mustDo("POST", "/api/users", map[string]any{
		"username": "carl", "password": "carls-long-password", "role": "user", "devices": []string{"0123456789abcdef"},
	}, 400)

	a.mustDo("DELETE", "/api/groups/"+g, nil, 200)
	_, list := a.doList("/api/devices")
	if len(list) != 1 || list[0]["id"] != d || list[0]["group"] != "" {
		t.Fatalf("device after group delete: %v", list)
	}
	_, users := a.doList("/api/users")
	for _, u := range users {
		if gs := u["groups"].([]any); len(gs) != 0 {
			t.Fatalf("group left on user %v", u)
		}
	}

	a.mustDo("DELETE", "/api/devices/"+d, nil, 200)
}

func TestDeletingDeviceRemovesItFromUsers(t *testing.T) {
	a := rolesApp(t)
	d := a.addDevice("PC", 1, "")
	a.mustDo("POST", "/api/users", map[string]any{
		"username": "bob", "password": "bobs-long-password", "role": "user", "devices": []string{d},
	}, 201)
	a.mustDo("DELETE", "/api/devices/"+d, nil, 200)
	u, _ := a.srv.store.User("bob")
	if len(u.Devices) != 0 {
		t.Fatalf("deleted device still assigned: %v", u.Devices)
	}
}

func TestAdministratorSafeguards(t *testing.T) {
	a := rolesApp(t)
	a.mustDo("DELETE", "/api/users/admin", nil, 409)
	a.mustDo("PUT", "/api/users/admin", map[string]any{"role": "user"}, 409)
	a.mustDo("POST", "/api/users", map[string]any{"username": "ab", "password": "long-enough-pass", "role": "user"}, 400)
	a.mustDo("POST", "/api/users", map[string]any{"username": "second", "password": "short", "role": "user"}, 400)
	a.mustDo("POST", "/api/users", map[string]any{"username": "second", "password": "long-enough-pass", "role": "root"}, 400)
	a.mustDo("POST", "/api/users", map[string]any{"username": "admin", "password": "long-enough-pass", "role": "user"}, 409)

	a.mustDo("POST", "/api/users", map[string]any{"username": "second", "password": "long-enough-pass", "role": "admin"}, 201)
	a.mustDo("PUT", "/api/users/admin", map[string]any{"role": "user"}, 200)
	// The first account is now a regular user, so the second one is the only administrator.
	second := a.second()
	second.login("second", "long-enough-pass")
	second.mustDo("PUT", "/api/users/second", map[string]any{"role": "user"}, 409)
	second.mustDo("GET", "/api/users", nil, 200)
	a.mustDo("GET", "/api/users", nil, 403)
}

func TestAccountsWithoutRoleAreAdministrators(t *testing.T) {
	if !(User{Name: "old"}).IsAdmin() {
		t.Fatal("an account from before roles existed must stay an administrator")
	}
	if (User{Name: "x", Role: roleUser}).IsAdmin() {
		t.Fatal("role user is not an administrator")
	}
}

func TestPasswordResetByAdminSignsUserOut(t *testing.T) {
	a := rolesApp(t)
	a.mustDo("POST", "/api/users", map[string]any{"username": "bob", "password": "bobs-long-password", "role": "user"}, 201)
	bob := a.second()
	bob.login("bob", "bobs-long-password")
	if _, out, _ := bob.do("GET", "/api/session", nil, nil); out["authenticated"] != true || out["role"] != "user" {
		t.Fatalf("session: %v", out)
	}

	a.mustDo("PUT", "/api/users/bob", map[string]any{"role": "user", "password": "a-brand-new-password"}, 200)
	if _, out, _ := bob.do("GET", "/api/session", nil, nil); out["authenticated"] != false {
		t.Fatalf("old session survived a password reset: %v", out)
	}
	bob.mustDo("POST", "/api/login", map[string]string{"username": "bob", "password": "bobs-long-password"}, 401)
	bob.login("bob", "a-brand-new-password")

	a.mustDo("DELETE", "/api/users/bob", nil, 200)
	if _, out, _ := bob.do("GET", "/api/session", nil, nil); out["authenticated"] != false {
		t.Fatalf("session of a deleted user still works: %v", out)
	}
}

func TestLogsFollowPermissions(t *testing.T) {
	a := rolesApp(t)
	d1 := a.addDevice("Mine", 1, "")
	d2 := a.addDevice("Not mine", 2, "")
	a.mustDo("POST", "/api/users", map[string]any{
		"username": "bob", "password": "bobs-long-password", "role": "user", "devices": []string{d1},
	}, 201)
	bob := a.second()
	bob.login("bob", "bobs-long-password")
	bob.mustDo("POST", "/api/devices/"+d1+"/wake", nil, 202)
	a.mustDo("POST", "/api/devices/"+d2+"/wake", nil, 202)

	deadline := time.Now().Add(5 * time.Second)
	var adminLog []map[string]any
	for time.Now().Before(deadline) {
		_, adminLog = a.doList("/api/logs")
		attempts := 0
		for _, e := range adminLog {
			if e["code"] == "attempt" {
				attempts++
			}
		}
		if attempts >= 2 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	kinds := map[string]bool{}
	for _, e := range adminLog {
		kinds[e["kind"].(string)+"/"+e["code"].(string)] = true
	}
	for _, want := range []string{"auth/login_ok", "admin/device_added", "admin/user_added", "wake/requested", "wake/attempt"} {
		if !kinds[want] {
			t.Fatalf("admin log lacks %s: %v", want, kinds)
		}
	}

	code, bobLog := bob.doList("/api/logs")
	if code != 200 || len(bobLog) == 0 {
		t.Fatalf("bob log: %d %v", code, bobLog)
	}
	for _, e := range bobLog {
		if e["kind"] != "wake" || e["deviceId"] != d1 {
			t.Fatalf("regular user sees a foreign or account event: %v", e)
		}
		if e["results"] != nil || e["user"] != nil || e["ip"] != nil {
			t.Fatalf("regular user sees internals: %v", e)
		}
	}
	if code, _ := (&testApp{t: t, srv: a.srv, ts: a.ts, client: &http.Client{}}).doList("/api/logs"); code != 401 {
		t.Fatalf("logs without a session: %d", code)
	}
}

func TestEventLogKeepsNewestOnly(t *testing.T) {
	l := newEventLog()
	for i := 0; i < maxEvents+50; i++ {
		l.Add(Event{Kind: "wake", Code: "attempt", N: i})
	}
	got := l.Snapshot(func(Event) bool { return true })
	if len(got) != maxEvents || got[0].N != maxEvents+49 {
		t.Fatalf("len %d first %d", len(got), got[0].N)
	}
}

func TestGroupsAndRolesSurviveRestart(t *testing.T) {
	dir := t.TempDir()
	st, err := openStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.CreateFirstUser("admin", "hash"); err != nil {
		t.Fatal(err)
	}
	g, err := st.AddGroup("Room")
	if err != nil {
		t.Fatal(err)
	}
	d, err := st.AddDevice(Device{Name: "PC", MAC: "aa:bb:cc:dd:ee:01", Group: g.ID, Methods: Methods{Global: true}})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.AddUser(User{Name: "bob", Hash: "h", Role: roleUser, Groups: []string{g.ID}, Devices: []string{d.ID}}); err != nil {
		t.Fatal(err)
	}
	again, err := openStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	u, _ := again.User("bob")
	dev, _ := again.Device(d.ID)
	if u.IsAdmin() || len(u.Groups) != 1 || dev.Group != g.ID || len(again.Groups()) != 1 {
		t.Fatalf("lost after restart: %+v %+v", u, dev)
	}
}
