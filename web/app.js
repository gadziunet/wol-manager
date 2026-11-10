'use strict';

(() => {
  const I18N = {
    en: {
      add: 'Add device', password: 'Password', logout: 'Log out',
      f_name: 'Name', f_mac: 'MAC address',
      f_host: 'IP address or host name (optional)',
      h_host: 'Used to check whether the computer is up and for unicast packets. Host names are resolved with DNS.',
      advanced: 'Advanced',
      f_netmask: 'Network mask (optional)',
      h_netmask: 'Together with an IP address it adds a directed broadcast for that network, e.g. 255.255.255.0.',
      f_interface: 'Network interface',
      h_interface: 'Auto sends through every interface.',
      f_methods: 'Wake methods',
      m_global: 'Limited broadcast (255.255.255.255) on every interface',
      m_directed: 'Directed broadcast (e.g. 192.168.1.255)',
      m_unicast: 'Unicast to the IP address',
      f_icmp: 'Check with ping (ICMP)',
      f_ports: 'TCP ports to check (optional)',
      h_ports: 'Comma separated, up to 5, e.g. 22, 3389. A closed port that answers still counts as online.',
      f_timeout: 'Give up after (s)', f_retry: 'Retry every (s)',
      delete: 'Delete', cancel: 'Cancel', save: 'Save', confirm_delete: 'Tap again to delete',
      f_current: 'Current password', f_new: 'New password (min. 10 characters)', f_repeat: 'Repeat new password',
      auto: 'Auto (all interfaces)',
      loading: 'Loading…',
      empty: 'No devices yet. Use “Add device”.',
      wake: 'Wake', stop: 'Stop', edit: 'Edit',
      add_device: 'Add device', edit_device: 'Edit device',
      details: 'Details (attempt {n})',
      online: 'Online', offline: 'Offline', unknown: 'Status unknown',
      s_waking: 'Waking… attempt {n}, {s} s of {max} s',
      s_online: 'Online after {s} s',
      s_timeout: 'No response after {s} s',
      s_sent: 'Packets sent (cannot verify without a host and a check)',
      s_cancelled: 'Stopped',
      login: 'Log in', username: 'Username', password_label: 'Password',
      setup_title: 'First start',
      setup_hint: 'Enter the setup token from the application log and choose an account.',
      f_token: 'Setup token', create_account: 'Create account',
      pw_changed: 'Password changed', pw_mismatch: 'The passwords do not match',
      e_error: 'Something went wrong.',
      e_network: 'Cannot reach the server.',
      e_bad_credentials: 'Wrong username or password.',
      e_rate_limited: 'Too many attempts. Try again later.',
      e_bad_setup_token: 'Wrong setup token.',
      e_setup_done: 'The account already exists.',
      e_invalid_username: 'Username: 3 to 32 characters (a–z, 0–9, _ . -).',
      e_invalid_password: 'Password: 10 to 128 characters.',
      e_invalid_name: 'Enter a name (up to 80 characters).',
      e_invalid_mac: 'Invalid MAC address.',
      e_invalid_host: 'Invalid IPv4 address or host name.',
      e_invalid_netmask: 'Invalid network mask.',
      e_invalid_interface: 'Invalid interface name.',
      e_no_method: 'Select at least one wake method.',
      e_invalid_ports: 'Ports: up to 5 numbers from 1 to 65535.',
      e_invalid_timeout: 'Timeout must be 10 to 600 seconds.',
      e_invalid_retry: 'Retry interval must be 2 to 60 seconds.',
      e_duplicate_mac: 'A device with this MAC address already exists.',
      e_too_many_devices: 'Too many devices.',
      e_not_found: 'Not found.',
      e_forbidden_origin: 'Request blocked (origin).',
      tab_devices: 'Devices', tab_logs: 'Logs', tab_settings: 'Settings',
      f_group: 'Group', no_group: 'No group', ungrouped: 'Other devices',
      wake_group: 'Wake group ({n})', group_started: 'Waking {n} devices…',
      empty_user: 'No devices are assigned to your account yet.',
      settings_users: 'Users', settings_groups: 'Groups',
      add_user: 'Add user', add_group: 'Add group',
      new_user: 'New user', edit_user: 'Edit user', new_group: 'New group', edit_group: 'Edit group',
      badge_admin: 'administrator', badge_user: 'user',
      all_devices: 'Everything', u_summary: '{d} devices · {g} groups', n_devices: '{n} devices',
      no_groups: 'No groups yet.',
      f_user_password: 'Password',
      h_pw_new: 'Required, 10 to 128 characters.',
      h_pw_keep: 'Leave empty to keep the current password. A new password signs the user out everywhere.',
      f_role: 'Role', role_user: 'User (wake and view status)', role_admin: 'Administrator',
      h_role: 'A user can only wake and see the devices and groups assigned to them and change their own password. An administrator can do everything.',
      f_user_groups: 'Groups', f_user_devices: 'Single devices', none_yet: 'None yet.',
      f_group_name: 'Group name',
      h_group: 'A device belongs to one group. Choose it when editing the device.',
      logs_empty: 'No events yet.',
      logs_note: 'The log is kept in memory (the last 500 events) and is cleared when the application restarts.',
      targets: 'Packet targets',
      ev_requested: '{user} started waking {device}',
      ev_requested_anon: 'Waking {device} was started',
      ev_requested_group: '{user} woke {device} with its group',
      ev_requested_group_anon: '{device} was woken with its group',
      ev_attempt: '{device}: attempt {n}',
      ev_online: '{device}: online after {n} s',
      ev_timeout: '{device}: no response after {n} s',
      ev_sent: '{device}: packets sent (cannot verify)',
      ev_cancelled: '{device}: stopped',
      ev_login_ok: '{user} logged in{ip}',
      ev_login_failed: 'Failed login for {user}{ip}',
      ev_password_changed: '{user} changed their password',
      ev_device_added: '{user} added device {device}',
      ev_device_changed: '{user} changed device {device}',
      ev_device_deleted: '{user} deleted device {device}',
      ev_group_added: '{user} added group {device}',
      ev_group_changed: '{user} changed group {device}',
      ev_group_deleted: '{user} deleted group {device}',
      ev_user_added: '{user} added user {device}',
      ev_user_changed: '{user} changed user {device}',
      ev_user_deleted: '{user} deleted user {device}',
      e_forbidden: 'Only an administrator can do that.',
      e_invalid_group: 'Invalid group.',
      e_invalid_reference: 'A selected device or group no longer exists.',
      e_invalid_role: 'Invalid role.',
      e_invalid_group_name: 'Group name: 1 to 40 characters.',
      e_duplicate_user: 'A user with this name already exists.',
      e_duplicate_group: 'A group with this name already exists.',
      e_last_admin: 'There must be at least one administrator.',
      e_cannot_delete_self: 'You cannot delete your own account.',
      e_limit_reached: 'Limit reached.',
    },
    pl: {
      add: 'Dodaj urządzenie', password: 'Hasło', logout: 'Wyloguj',
      f_name: 'Nazwa', f_mac: 'Adres MAC',
      f_host: 'Adres IP lub nazwa hosta (opcjonalnie)',
      h_host: 'Służy do sprawdzania, czy komputer działa, i do pakietów unicast. Nazwy hostów są rozwiązywane przez DNS.',
      advanced: 'Zaawansowane',
      f_netmask: 'Maska sieci (opcjonalnie)',
      h_netmask: 'Razem z adresem IP dodaje broadcast kierowany dla tej sieci, np. 255.255.255.0.',
      f_interface: 'Interfejs sieciowy',
      h_interface: 'Auto wysyła przez każdy interfejs.',
      f_methods: 'Metody budzenia',
      m_global: 'Broadcast ogólny (255.255.255.255) przez każdy interfejs',
      m_directed: 'Broadcast kierowany (np. 192.168.1.255)',
      m_unicast: 'Unicast na adres IP',
      f_icmp: 'Sprawdzaj pingiem (ICMP)',
      f_ports: 'Porty TCP do sprawdzania (opcjonalnie)',
      h_ports: 'Oddzielone przecinkami, do 5, np. 22, 3389. Zamknięty port, który odpowiada, też oznacza, że komputer działa.',
      f_timeout: 'Poddaj się po (s)', f_retry: 'Ponawiaj co (s)',
      delete: 'Usuń', cancel: 'Anuluj', save: 'Zapisz', confirm_delete: 'Kliknij ponownie, aby usunąć',
      f_current: 'Obecne hasło', f_new: 'Nowe hasło (min. 10 znaków)', f_repeat: 'Powtórz nowe hasło',
      auto: 'Auto (wszystkie interfejsy)',
      loading: 'Ładowanie…',
      empty: 'Brak urządzeń. Użyj „Dodaj urządzenie”.',
      wake: 'Wybudź', stop: 'Zatrzymaj', edit: 'Edytuj',
      add_device: 'Dodaj urządzenie', edit_device: 'Edytuj urządzenie',
      details: 'Szczegóły (próba {n})',
      online: 'Działa', offline: 'Wyłączony', unknown: 'Stan nieznany',
      s_waking: 'Budzenie… próba {n}, {s} s z {max} s',
      s_online: 'Wstał po {s} s',
      s_timeout: 'Brak odpowiedzi po {s} s',
      s_sent: 'Pakiety wysłane (nie da się sprawdzić bez hosta i sprawdzania)',
      s_cancelled: 'Zatrzymano',
      login: 'Zaloguj', username: 'Użytkownik', password_label: 'Hasło',
      setup_title: 'Pierwsze uruchomienie',
      setup_hint: 'Wpisz token konfiguracji z logu aplikacji i wybierz konto.',
      f_token: 'Token konfiguracji', create_account: 'Utwórz konto',
      pw_changed: 'Hasło zmienione', pw_mismatch: 'Hasła się nie zgadzają',
      e_error: 'Coś poszło nie tak.',
      e_network: 'Brak połączenia z serwerem.',
      e_bad_credentials: 'Błędny użytkownik lub hasło.',
      e_rate_limited: 'Za dużo prób. Spróbuj później.',
      e_bad_setup_token: 'Błędny token konfiguracji.',
      e_setup_done: 'Konto już istnieje.',
      e_invalid_username: 'Użytkownik: 3 do 32 znaków (a–z, 0–9, _ . -).',
      e_invalid_password: 'Hasło: od 10 do 128 znaków.',
      e_invalid_name: 'Podaj nazwę (do 80 znaków).',
      e_invalid_mac: 'Błędny adres MAC.',
      e_invalid_host: 'Błędny adres IPv4 lub nazwa hosta.',
      e_invalid_netmask: 'Błędna maska sieci.',
      e_invalid_interface: 'Błędna nazwa interfejsu.',
      e_no_method: 'Wybierz co najmniej jedną metodę budzenia.',
      e_invalid_ports: 'Porty: do 5 liczb od 1 do 65535.',
      e_invalid_timeout: 'Limit czasu: od 10 do 600 sekund.',
      e_invalid_retry: 'Ponawianie: od 2 do 60 sekund.',
      e_duplicate_mac: 'Urządzenie z tym adresem MAC już istnieje.',
      e_too_many_devices: 'Za dużo urządzeń.',
      e_not_found: 'Nie znaleziono.',
      e_forbidden_origin: 'Żądanie zablokowane (origin).',
      tab_devices: 'Urządzenia', tab_logs: 'Logi', tab_settings: 'Ustawienia',
      f_group: 'Grupa', no_group: 'Bez grupy', ungrouped: 'Pozostałe urządzenia',
      wake_group: 'Wybudź grupę ({n})', group_started: 'Budzenie urządzeń: {n}…',
      empty_user: 'Do Twojego konta nie przypisano jeszcze żadnych urządzeń.',
      settings_users: 'Użytkownicy', settings_groups: 'Grupy',
      add_user: 'Dodaj użytkownika', add_group: 'Dodaj grupę',
      new_user: 'Nowy użytkownik', edit_user: 'Edytuj użytkownika', new_group: 'Nowa grupa', edit_group: 'Edytuj grupę',
      badge_admin: 'administrator', badge_user: 'użytkownik',
      all_devices: 'Wszystko', u_summary: 'Urządzenia: {d} · grupy: {g}', n_devices: 'Urządzenia: {n}',
      no_groups: 'Brak grup.',
      f_user_password: 'Hasło',
      h_pw_new: 'Wymagane, od 10 do 128 znaków.',
      h_pw_keep: 'Zostaw puste, aby nie zmieniać hasła. Nowe hasło wylogowuje użytkownika wszędzie.',
      f_role: 'Rola', role_user: 'Użytkownik (budzi i widzi stan)', role_admin: 'Administrator',
      h_role: 'Użytkownik może tylko budzić i widzieć przypisane mu urządzenia i grupy oraz zmieniać własne hasło. Administrator może wszystko.',
      f_user_groups: 'Grupy', f_user_devices: 'Pojedyncze urządzenia', none_yet: 'Brak.',
      f_group_name: 'Nazwa grupy',
      h_group: 'Urządzenie należy do jednej grupy. Wybierzesz ją przy edycji urządzenia.',
      logs_empty: 'Brak zdarzeń.',
      logs_note: 'Log jest przechowywany w pamięci (ostatnie 500 zdarzeń) i znika po restarcie aplikacji.',
      targets: 'Cele pakietów',
      ev_requested: '{user} rozpoczął budzenie: {device}',
      ev_requested_anon: 'Rozpoczęto budzenie: {device}',
      ev_requested_group: '{user} wybudził {device} razem z grupą',
      ev_requested_group_anon: 'Wybudzono {device} razem z grupą',
      ev_attempt: '{device}: próba {n}',
      ev_online: '{device}: wstał po {n} s',
      ev_timeout: '{device}: brak odpowiedzi po {n} s',
      ev_sent: '{device}: pakiety wysłane (nie da się sprawdzić)',
      ev_cancelled: '{device}: zatrzymano',
      ev_login_ok: '{user} zalogował się{ip}',
      ev_login_failed: 'Nieudane logowanie: {user}{ip}',
      ev_password_changed: '{user} zmienił hasło',
      ev_device_added: '{user} dodał urządzenie {device}',
      ev_device_changed: '{user} zmienił urządzenie {device}',
      ev_device_deleted: '{user} usunął urządzenie {device}',
      ev_group_added: '{user} dodał grupę {device}',
      ev_group_changed: '{user} zmienił grupę {device}',
      ev_group_deleted: '{user} usunął grupę {device}',
      ev_user_added: '{user} dodał użytkownika {device}',
      ev_user_changed: '{user} zmienił użytkownika {device}',
      ev_user_deleted: '{user} usunął użytkownika {device}',
      e_forbidden: 'Tylko administrator może to zrobić.',
      e_invalid_group: 'Błędna grupa.',
      e_invalid_reference: 'Wybrane urządzenie lub grupa już nie istnieje.',
      e_invalid_role: 'Błędna rola.',
      e_invalid_group_name: 'Nazwa grupy: od 1 do 40 znaków.',
      e_duplicate_user: 'Użytkownik o tej nazwie już istnieje.',
      e_duplicate_group: 'Grupa o tej nazwie już istnieje.',
      e_last_admin: 'Musi zostać co najmniej jeden administrator.',
      e_cannot_delete_self: 'Nie możesz usunąć własnego konta.',
      e_limit_reached: 'Osiągnięto limit.',
    },
  };

  const $ = (sel, root = document) => root.querySelector(sel);

  function h(tag, props, ...kids) {
    const el = document.createElement(tag);
    for (const [k, v] of Object.entries(props || {})) {
      if (v === false || v == null) continue;
      if (k === 'class') el.className = v;
      else if (k.startsWith('on')) el.addEventListener(k.slice(2), v);
      else el.setAttribute(k, v === true ? '' : v);
    }
    for (const kid of kids.flat()) {
      if (kid == null || kid === false) continue;
      el.append(kid instanceof Node ? kid : document.createTextNode(String(kid)));
    }
    return el;
  }

  function storedLang() {
    try { return localStorage.getItem('wol_lang'); } catch { return null; }
  }
  function saveLang(l) {
    try { localStorage.setItem('wol_lang', l); } catch { /* storage unavailable */ }
  }

  let lang = storedLang();
  if (!I18N[lang]) lang = (navigator.language || 'en').toLowerCase().startsWith('pl') ? 'pl' : 'en';

  function t(key, vars) {
    let s = (I18N[lang] && I18N[lang][key]) ?? I18N.en[key] ?? key;
    if (vars) for (const [k, v] of Object.entries(vars)) s = s.replaceAll('{' + k + '}', v);
    return s;
  }

  function applyI18n() {
    document.documentElement.lang = lang;
    $('#lang').value = lang;
    document.querySelectorAll('[data-i18n]').forEach((el) => { el.textContent = t(el.dataset.i18n); });
  }

  class ApiError extends Error {
    constructor(status, code, field) {
      super(code);
      this.status = status;
      this.code = code;
      this.field = field;
    }
  }

  async function api(method, path, body) {
    const init = { method, credentials: 'same-origin', headers: {} };
    if (method !== 'GET') {
      init.headers['Content-Type'] = 'application/json';
      init.body = JSON.stringify(body ?? {});
    }
    let res;
    try { res = await fetch('/api' + path, init); } catch { throw new ApiError(0, 'network'); }
    let data = null;
    try { data = await res.json(); } catch { /* empty body */ }
    if (!res.ok) throw new ApiError(res.status, (data && data.error) || 'error', data && data.field);
    return data;
  }

  function errText(e) {
    const key = 'e_' + (e && e.code);
    return (I18N[lang] && I18N[lang][key]) || I18N.en[key] || t('e_error');
  }

  const view = $('#view');
  const setView = (...nodes) => view.replaceChildren(...nodes);

  let mainActive = false;
  let busy = false;
  let timer = 0;
  let hasWaking = false;
  let editing = null;
  const cards = new Map();

  /* ---------- sign in / first start ---------- */

  function showLogin() {
    leaveMain();
    const user = h('input', { name: 'username', autocomplete: 'username', autocapitalize: 'off', spellcheck: 'false', required: true });
    const pass = h('input', { name: 'password', type: 'password', autocomplete: 'current-password', required: true });
    const err = h('p', { class: 'err', role: 'alert' });
    const btn = h('button', { type: 'submit', class: 'btn primary' }, t('login'));
    const form = h('form', { class: 'panel' },
      h('h2', {}, t('login')),
      h('label', {}, t('username'), user),
      h('label', {}, t('password_label'), pass),
      err, btn);
    form.addEventListener('submit', async (ev) => {
      ev.preventDefault();
      err.textContent = '';
      btn.disabled = true;
      try {
        await api('POST', '/login', { username: user.value, password: pass.value });
        await boot();
      } catch (e) {
        err.textContent = errText(e);
        pass.value = '';
      } finally {
        btn.disabled = false;
      }
    });
    setView(form);
    user.focus();
  }

  function showSetup() {
    leaveMain();
    const token = h('input', { name: 'token', autocomplete: 'off', autocapitalize: 'off', spellcheck: 'false', required: true });
    const user = h('input', { name: 'username', autocomplete: 'username', autocapitalize: 'off', spellcheck: 'false', required: true, minlength: '3', maxlength: '32' });
    const pass = h('input', { name: 'password', type: 'password', autocomplete: 'new-password', required: true, minlength: '10', maxlength: '128' });
    const repeat = h('input', { name: 'repeat', type: 'password', autocomplete: 'new-password', required: true, minlength: '10', maxlength: '128' });
    const err = h('p', { class: 'err', role: 'alert' });
    const btn = h('button', { type: 'submit', class: 'btn primary' }, t('create_account'));
    const form = h('form', { class: 'panel' },
      h('h2', {}, t('setup_title')),
      h('p', {}, t('setup_hint')),
      h('label', {}, t('f_token'), token),
      h('label', {}, t('username'), user),
      h('label', {}, t('f_new'), pass),
      h('label', {}, t('f_repeat'), repeat),
      err, btn);
    form.addEventListener('submit', async (ev) => {
      ev.preventDefault();
      err.textContent = '';
      if (pass.value !== repeat.value) { err.textContent = t('pw_mismatch'); return; }
      btn.disabled = true;
      try {
        await api('POST', '/setup', { token: token.value, username: user.value, password: pass.value });
        await boot();
      } catch (e) {
        err.textContent = errText(e);
      } finally {
        btn.disabled = false;
      }
    });
    setView(form);
    token.focus();
  }

  /* ---------- state ---------- */

  let me = { user: '', role: 'user' };
  let tab = 'devices';
  let groupsCache = [];
  let layoutSig = '';
  let logSig = '';
  let gen = 0;
  const isAdmin = () => me.role === 'admin';

  function notify(text, ok) {
    const box = $('#flash');
    if (!box) return;
    box.textContent = text;
    box.className = 'err flash' + (ok ? ' ok' : '');
    setTimeout(() => { if (box.textContent === text) box.textContent = ''; }, 6000);
  }

  function schedule() {
    clearTimeout(timer);
    if (!mainActive || document.hidden || tab === 'settings') return;
    timer = setTimeout(refresh, tab === 'logs' ? 3000 : (hasWaking ? 1000 : 4000));
  }

  function refresh() {
    if (!mainActive) return;
    if (tab === 'devices') refreshDevices();
    else if (tab === 'logs') refreshLogs();
  }

  /* ---------- devices ---------- */

  function wakeText(w) {
    switch (w.state) {
      case 'waking': return t('s_waking', { n: w.attempt, s: w.elapsedSec, max: w.timeoutSec });
      case 'online': return t('s_online', { s: w.elapsedSec });
      case 'timeout': return t('s_timeout', { s: w.elapsedSec });
      case 'sent': return t('s_sent');
      case 'cancelled': return t('s_cancelled');
      default: return '';
    }
  }

  function resultLine(r) {
    const route = (r.from ? r.from + ' → ' : '') + (r.to || '') + (r.port ? ':' + r.port : '');
    return `${r.method} ${route} ${r.ok ? 'OK' : (r.err || 'ERR')}`.trim();
  }

  function buildCard(d) {
    const w = d.wake;
    const up = d.status.known && d.status.online;
    const statusText = d.status.known ? (up ? t('online') : t('offline')) : t('unknown');
    const stateClass = w.state === 'online' ? ' good' : (w.state === 'timeout' ? ' warn' : '');
    return h('article', { class: 'card', 'data-id': d.id },
      h('div', { class: 'card-head' },
        h('span', { class: 'dot' + (up ? ' online' : ''), title: statusText }),
        h('h2', {}, d.name),
        isAdmin() ? h('button', { type: 'button', class: 'btn', onclick: () => openDevice(d) }, t('edit')) : null),
      h('p', { class: 'meta' }, [d.host, d.mac, statusText].filter(Boolean).join(' · ')),
      h('p', { class: 'state' + stateClass }, wakeText(w)),
      h('div', { class: 'actions' },
        w.state === 'waking'
          ? h('button', { type: 'button', class: 'btn', onclick: () => cancelWake(d) }, t('stop'))
          : h('button', { type: 'button', class: 'btn primary wake', onclick: () => wake(d) }, t('wake'))));
  }

  function ensureCard(d) {
    const sig = JSON.stringify(d);
    const c = cards.get(d.id);
    if (!c || c.sig !== sig) {
      const el = buildCard(d);
      if (c && c.el.isConnected) c.el.replaceWith(el);
      cards.set(d.id, { sig, el });
    }
    return cards.get(d.id).el;
  }

  function renderDevices(list, groups) {
    const sections = $('#sections');
    if (!sections) return;
    const empty = $('#empty');
    empty.hidden = list.length > 0;
    empty.textContent = isAdmin() ? t('empty') : t('empty_user');

    const seen = new Set(list.map((d) => d.id));
    for (const id of [...cards.keys()]) if (!seen.has(id)) cards.delete(id);
    const els = new Map(list.map((d) => [d.id, ensureCard(d)]));

    const known = new Set(groups.map((g) => g.id));
    const buckets = groups
      .map((g) => ({ g, devs: list.filter((d) => d.group === g.id) }))
      .filter((b) => b.devs.length);
    const loose = list.filter((d) => !d.group || !known.has(d.group));

    const sig = lang + JSON.stringify([buckets.map((b) => [b.g.id, b.g.name, b.devs.map((d) => d.id)]), loose.map((d) => d.id)]);
    if (sig === layoutSig) return;
    layoutSig = sig;

    const grid = (devs) => h('div', { class: 'grid' }, devs.map((d) => els.get(d.id)));
    const out = buckets.map((b) => h('section', { class: 'group' },
      h('div', { class: 'group-head' },
        h('h2', {}, b.g.name),
        h('button', { type: 'button', class: 'btn', onclick: () => wakeGroup(b.g) }, t('wake_group', { n: b.devs.length }))),
      grid(b.devs)));
    if (loose.length) {
      out.push(h('section', { class: 'group' },
        buckets.length ? h('div', { class: 'group-head' }, h('h2', {}, t('ungrouped'))) : null,
        grid(loose)));
    }
    sections.replaceChildren(...out);
  }

  async function refreshDevices() {
    if (!mainActive || tab !== 'devices' || busy) return;
    busy = true;
    const g = gen;
    try {
      const [list, groups] = await Promise.all([api('GET', '/devices'), api('GET', '/groups')]);
      if (g !== gen) return;
      groupsCache = groups;
      hasWaking = list.some((d) => d.wake.state === 'waking');
      renderDevices(list, groups);
    } catch (e) {
      if (e.status === 401 && g === gen) { busy = false; boot(); return; }
    } finally {
      if (g === gen) busy = false;
    }
    if (g === gen) schedule();
  }

  async function wake(d) {
    try { await api('POST', `/devices/${encodeURIComponent(d.id)}/wake`); } catch (e) { notify(errText(e)); }
    refresh();
  }

  async function cancelWake(d) {
    try { await api('POST', `/devices/${encodeURIComponent(d.id)}/wake/cancel`); } catch (e) { notify(errText(e)); }
    refresh();
  }

  async function wakeGroup(g) {
    try {
      const r = await api('POST', `/groups/${encodeURIComponent(g.id)}/wake`);
      notify(t('group_started', { n: r.started }), true);
    } catch (e) { notify(errText(e)); }
    refresh();
  }

  /* ---------- logs ---------- */

  function eventText(e) {
    const v = { user: e.user || '', device: e.device || '', n: e.n || 0, ip: e.ip ? ' (' + e.ip + ')' : '' };
    switch (e.code) {
      case 'requested': return t(v.user ? 'ev_requested' : 'ev_requested_anon', v);
      case 'requested_group': return t(v.user ? 'ev_requested_group' : 'ev_requested_group_anon', v);
      case 'attempt': return t('ev_attempt', v);
      case 'online': return t('ev_online', v);
      case 'timeout': return t('ev_timeout', v);
      case 'sent': return t('ev_sent', v);
      case 'cancelled': return t('ev_cancelled', v);
      default: {
        const key = 'ev_' + e.code;
        return (I18N[lang] && I18N[lang][key]) || I18N.en[key] ? t(key, v) : e.code;
      }
    }
  }

  function eventRow(e, open) {
    const key = [e.time, e.code, e.deviceId, e.n].join('|');
    const cls = e.code === 'online' ? ' good' : (e.code === 'timeout' || e.code === 'login_failed' ? ' warn' : '');
    const lines = (e.results || []).map((r) => h('li', { class: r.ok ? '' : 'bad' }, resultLine(r)));
    return h('div', { class: 'logrow' },
      h('time', { datetime: e.time }, new Date(e.time).toLocaleString(lang)),
      h('span', { class: 'msg' + cls }, eventText(e)),
      lines.length ? h('details', { 'data-key': key, open: open.has(key) }, h('summary', {}, t('targets')), h('ul', {}, lines)) : null);
  }

  function renderLogs(list) {
    const box = $('#logList');
    if (!box) return;
    const sig = lang + JSON.stringify(list);
    if (sig === logSig) return;
    logSig = sig;
    const open = new Set([...box.querySelectorAll('details[open]')].map((d) => d.dataset.key));
    box.replaceChildren(...(list.length ? list.map((e) => eventRow(e, open)) : [h('p', { class: 'empty' }, t('logs_empty'))]));
  }

  async function refreshLogs() {
    if (!mainActive || tab !== 'logs' || busy) return;
    busy = true;
    const g = gen;
    try {
      const list = await api('GET', '/logs');
      if (g !== gen) return;
      hasWaking = false;
      renderLogs(list);
    } catch (e) {
      if (e.status === 401 && g === gen) { busy = false; boot(); return; }
    } finally {
      if (g === gen) busy = false;
    }
    if (g === gen) schedule();
  }

  /* ---------- settings (administrators) ---------- */

  function armDelete(btn, onConfirm) {
    let armed = 0;
    const reset = () => { clearTimeout(armed); armed = 0; btn.textContent = t('delete'); };
    btn.addEventListener('click', async () => {
      if (!armed) { btn.textContent = t('confirm_delete'); armed = setTimeout(reset, 4000); return; }
      reset();
      await onConfirm();
    });
    return reset;
  }

  async function renderSettings() {
    const g = gen;
    let users; let groups; let devices;
    try {
      [users, groups, devices] = await Promise.all([api('GET', '/users'), api('GET', '/groups'), api('GET', '/devices')]);
    } catch (e) {
      if (e.status === 401) { boot(); return; }
      notify(errText(e));
      return;
    }
    if (g !== gen || tab !== 'settings') return;
    groupsCache = groups;

    const userRows = users.map((u) => h('div', { class: 'item' },
      h('div', { class: 'grow' },
        h('div', { class: 'name' }, u.name,
          h('span', { class: 'badge' + (u.role === 'admin' ? ' admin' : '') }, t(u.role === 'admin' ? 'badge_admin' : 'badge_user'))),
        h('div', { class: 'sub' }, u.role === 'admin' ? t('all_devices') : t('u_summary', { d: u.devices.length, g: u.groups.length }))),
      h('button', { type: 'button', class: 'btn', onclick: () => openUser(u, groups, devices) }, t('edit'))));

    const groupRows = groups.map((gr) => h('div', { class: 'item' },
      h('div', { class: 'grow' },
        h('div', { class: 'name' }, gr.name),
        h('div', { class: 'sub' }, t('n_devices', { n: gr.devices }))),
      h('button', { type: 'button', class: 'btn', onclick: () => openGroup(gr) }, t('edit'))));

    $('#pane').replaceChildren(h('div', { class: 'settings' },
      h('section', {},
        h('div', { class: 'section-head' },
          h('h2', {}, t('settings_users')),
          h('button', { type: 'button', class: 'btn primary', onclick: () => openUser(null, groups, devices) }, t('add_user'))),
        h('div', { class: 'rows' }, userRows)),
      h('section', {},
        h('div', { class: 'section-head' },
          h('h2', {}, t('settings_groups')),
          h('button', { type: 'button', class: 'btn primary', onclick: () => openGroup(null) }, t('add_group'))),
        groupRows.length
          ? h('div', { class: 'rows' }, groupRows)
          : h('p', { class: 'empty' }, t('no_groups')))));
  }

  const userDialog = $('#userDialog');
  const userForm = $('#userForm');
  let userEditing = null;

  function checkList(box, items, selected, name) {
    box.replaceChildren(...(items.length
      ? items.map((i) => h('label', { class: 'check' },
        h('input', { type: 'checkbox', name, value: i.id, checked: selected.includes(i.id) }),
        h('span', {}, i.name)))
      : [h('p', { class: 'none' }, t('none_yet'))]));
  }

  function syncRole() {
    $('#userAccess').hidden = userForm.elements.role.value === 'admin';
  }

  function openUser(u, groups, devices) {
    userEditing = u;
    const f = userForm.elements;
    $('#userTitle').textContent = u ? t('edit_user') : t('new_user');
    f.username.value = u ? u.name : '';
    f.username.readOnly = !!u;
    f.password.value = '';
    f.password.required = !u;
    $('#userPasswordHint').textContent = t(u ? 'h_pw_keep' : 'h_pw_new');
    f.role.value = u ? u.role : 'user';
    checkList($('#userGroups'), groups, u ? u.groups : [], 'group');
    checkList($('#userDevices'), devices, u ? u.devices : [], 'device');
    syncRole();
    $('#btnUserDelete').hidden = !u || u.name === me.user;
    resetUserDelete();
    $('#userError').textContent = '';
    userDialog.showModal();
  }

  userForm.elements.role.addEventListener('change', syncRole);
  $('#btnUserCancel').addEventListener('click', () => userDialog.close());
  const resetUserDelete = armDelete($('#btnUserDelete'), async () => {
    try {
      await api('DELETE', '/users/' + encodeURIComponent(userEditing.name));
      userDialog.close();
      renderSettings();
    } catch (e) {
      $('#userError').textContent = errText(e);
    }
  });

  userForm.addEventListener('submit', async (ev) => {
    ev.preventDefault();
    const f = userForm.elements;
    const err = $('#userError');
    err.textContent = '';
    const pick = (sel) => [...$(sel).querySelectorAll('input:checked')].map((i) => i.value);
    const admin = f.role.value === 'admin';
    const body = { role: f.role.value, devices: admin ? [] : pick('#userDevices'), groups: admin ? [] : pick('#userGroups') };
    try {
      if (userEditing) {
        if (f.password.value) body.password = f.password.value;
        await api('PUT', '/users/' + encodeURIComponent(userEditing.name), body);
      } else {
        body.username = f.username.value;
        body.password = f.password.value;
        await api('POST', '/users', body);
      }
      const self = userEditing && userEditing.name === me.user;
      userDialog.close();
      if (self) boot(); else renderSettings();
    } catch (e) {
      err.textContent = errText(e);
    }
  });

  const groupDialog = $('#groupDialog');
  const groupForm = $('#groupForm');
  let groupEditing = null;

  function openGroup(g) {
    groupEditing = g;
    $('#groupTitle').textContent = g ? t('edit_group') : t('new_group');
    groupForm.elements.name.value = g ? g.name : '';
    $('#btnGroupDelete').hidden = !g;
    resetGroupDelete();
    $('#groupError').textContent = '';
    groupDialog.showModal();
  }

  $('#btnGroupCancel').addEventListener('click', () => groupDialog.close());
  const resetGroupDelete = armDelete($('#btnGroupDelete'), async () => {
    try {
      await api('DELETE', '/groups/' + encodeURIComponent(groupEditing.id));
      groupDialog.close();
      renderSettings();
    } catch (e) {
      $('#groupError').textContent = errText(e);
    }
  });

  groupForm.addEventListener('submit', async (ev) => {
    ev.preventDefault();
    const name = groupForm.elements.name.value;
    try {
      if (groupEditing) await api('PUT', '/groups/' + encodeURIComponent(groupEditing.id), { name });
      else await api('POST', '/groups', { name });
      groupDialog.close();
      renderSettings();
    } catch (e) {
      $('#groupError').textContent = errText(e);
    }
  });

  /* ---------- main view and tabs ---------- */

  function showTab(name) {
    tab = name;
    gen++;
    busy = false;
    clearTimeout(timer);
    layoutSig = '';
    logSig = '';
    cards.clear();
    const flash = $('#flash');
    if (flash) flash.textContent = '';
    document.querySelectorAll('#tabs .tab').forEach((b) => b.setAttribute('aria-current', b.dataset.tab === tab ? 'page' : 'false'));
    $('#btnAdd').hidden = !(isAdmin() && tab === 'devices');
    const pane = $('#pane');
    if (tab === 'devices') {
      pane.replaceChildren(h('div', { id: 'sections' }), h('p', { class: 'empty', id: 'empty', hidden: true }));
      refreshDevices();
    } else if (tab === 'logs') {
      pane.replaceChildren(isAdmin() ? h('p', { class: 'logs-note' }, t('logs_note')) : null, h('div', { id: 'logList' }));
      refreshLogs();
    } else {
      pane.replaceChildren(h('p', { class: 'center' }, t('loading')));
      renderSettings();
    }
  }

  function enterMain(s) {
    me = { user: s.user, role: s.role === 'admin' ? 'admin' : 'user' };
    mainActive = true;
    $('#btnPassword').hidden = false;
    $('#btnLogout').hidden = false;
    $('#tabs').hidden = false;
    $('#tabSettings').hidden = !isAdmin();
    if (tab === 'settings' && !isAdmin()) tab = 'devices';
    setView(h('p', { class: 'err flash', id: 'flash', role: 'alert' }), h('div', { id: 'pane' }));
    showTab(tab);
  }

  function leaveMain() {
    mainActive = false;
    gen++;
    clearTimeout(timer);
    cards.clear();
    tab = 'devices';
    for (const id of ['#btnAdd', '#btnPassword', '#btnLogout', '#tabs']) $(id).hidden = true;
  }

  /* ---------- device dialog ---------- */

  const deviceDialog = $('#deviceDialog');
  const deviceForm = $('#deviceForm');
  const deleteBtn = $('#btnDelete');
  let deleteArmed = 0;

  async function fillInterfaces(selected) {
    let list = [];
    try { list = await api('GET', '/interfaces'); } catch { /* optional */ }
    const sel = deviceForm.elements.interface;
    sel.replaceChildren(
      h('option', { value: '' }, t('auto')),
      ...list.map((i) => h('option', { value: i.name }, `${i.name} (${i.address})`)));
    if (selected && ![...sel.options].some((o) => o.value === selected)) {
      sel.append(h('option', { value: selected }, selected));
    }
    sel.value = selected || '';
  }

  async function fillGroups(selected) {
    let list = groupsCache;
    try { list = await api('GET', '/groups'); groupsCache = list; } catch { /* keep the cached list */ }
    const sel = deviceForm.elements.group;
    sel.replaceChildren(
      h('option', { value: '' }, t('no_group')),
      ...list.map((g) => h('option', { value: g.id }, g.name)));
    sel.value = list.some((g) => g.id === selected) ? selected : '';
  }

  async function openDevice(d) {
    editing = d || null;
    const f = deviceForm.elements;
    await Promise.all([fillInterfaces(d ? d.interface : ''), fillGroups(d ? d.group : '')]);
    $('#deviceTitle').textContent = d ? t('edit_device') : t('add_device');
    f.name.value = d ? d.name : '';
    f.mac.value = d ? d.mac : '';
    f.host.value = d ? d.host : '';
    f.netmask.value = d ? d.netmask : '';
    f.m_global.checked = d ? d.methods.global : true;
    f.m_directed.checked = d ? d.methods.directed : true;
    f.m_unicast.checked = d ? d.methods.unicast : true;
    f.icmp.checked = d ? d.icmp : true;
    f.checkPorts.value = d ? d.checkPorts.join(', ') : '';
    f.timeoutSec.value = d ? d.timeoutSec : 120;
    f.retrySec.value = d ? d.retrySec : 10;
    deleteBtn.hidden = !d;
    resetDelete();
    $('#deviceError').textContent = '';
    deviceDialog.showModal();
  }

  function resetDelete() {
    clearTimeout(deleteArmed);
    deleteArmed = 0;
    deleteBtn.textContent = t('delete');
  }

  deviceForm.addEventListener('submit', async (ev) => {
    ev.preventDefault();
    const f = deviceForm.elements;
    const body = {
      name: f.name.value,
      mac: f.mac.value,
      host: f.host.value,
      netmask: f.netmask.value,
      interface: f.interface.value,
      group: f.group.value,
      methods: { global: f.m_global.checked, directed: f.m_directed.checked, unicast: f.m_unicast.checked },
      checkPorts: f.checkPorts.value.split(/[\s,;]+/).filter(Boolean).map(Number),
      icmp: f.icmp.checked,
      timeoutSec: Number(f.timeoutSec.value),
      retrySec: Number(f.retrySec.value),
    };
    try {
      if (editing) await api('PUT', '/devices/' + encodeURIComponent(editing.id), body);
      else await api('POST', '/devices', body);
      deviceDialog.close();
      refresh();
    } catch (e) {
      $('#deviceError').textContent = errText(e);
    }
  });

  $('#btnDeviceCancel').addEventListener('click', () => deviceDialog.close());

  deleteBtn.addEventListener('click', async () => {
    if (!editing) return;
    if (!deleteArmed) {
      deleteBtn.textContent = t('confirm_delete');
      deleteArmed = setTimeout(resetDelete, 4000);
      return;
    }
    try {
      await api('DELETE', '/devices/' + encodeURIComponent(editing.id));
      deviceDialog.close();
      refresh();
    } catch (e) {
      $('#deviceError').textContent = errText(e);
    }
  });

  /* ---------- password dialog ---------- */

  const passwordDialog = $('#passwordDialog');
  const passwordForm = $('#passwordForm');

  $('#btnPassword').addEventListener('click', () => {
    passwordForm.reset();
    $('#passwordError').textContent = '';
    passwordDialog.showModal();
  });
  $('#btnPasswordCancel').addEventListener('click', () => passwordDialog.close());

  passwordForm.addEventListener('submit', async (ev) => {
    ev.preventDefault();
    const f = passwordForm.elements;
    const err = $('#passwordError');
    err.textContent = '';
    if (f.new.value !== f.repeat.value) { err.textContent = t('pw_mismatch'); return; }
    try {
      await api('POST', '/password', { current: f.current.value, new: f.new.value });
      passwordDialog.close();
      notify(t('pw_changed'));
    } catch (e) {
      err.textContent = errText(e);
    }
  });

  /* ---------- shell ---------- */

  $('#btnAdd').addEventListener('click', () => openDevice(null));
  document.querySelectorAll('#tabs .tab').forEach((b) => b.addEventListener('click', () => showTab(b.dataset.tab)));
  $('#btnLogout').addEventListener('click', async () => {
    try { await api('POST', '/logout'); } catch { /* ignore */ }
    boot();
  });
  $('#lang').addEventListener('change', (ev) => {
    lang = ev.target.value;
    saveLang(lang);
    boot();
  });
  document.addEventListener('visibilitychange', () => {
    if (!document.hidden && mainActive) refresh();
  });

  async function boot() {
    applyI18n();
    resetDelete();
    let s;
    try {
      s = await api('GET', '/session');
    } catch (e) {
      leaveMain();
      setView(h('p', { class: 'center' }, errText(e)));
      return;
    }
    if (s.authenticated) enterMain(s);
    else if (s.setupRequired) showSetup();
    else showLogin();
  }

  setView(h('p', { class: 'center' }, t('loading')));
  boot();
})();
