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

  /* ---------- device list ---------- */

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
    const lines = (w.last || []).map((r) => h('li', { class: r.ok ? '' : 'bad' }, resultLine(r)));
    return h('article', { class: 'card', 'data-id': d.id },
      h('div', { class: 'card-head' },
        h('span', { class: 'dot' + (up ? ' online' : ''), title: statusText }),
        h('h2', {}, d.name),
        h('button', { type: 'button', class: 'btn', onclick: () => openDevice(d) }, t('edit'))),
      h('p', { class: 'meta' }, [d.host || '—', d.mac, statusText].join(' · ')),
      h('p', { class: 'state' + stateClass }, wakeText(w)),
      h('div', { class: 'actions' },
        w.state === 'waking'
          ? h('button', { type: 'button', class: 'btn', onclick: () => cancelWake(d) }, t('stop'))
          : h('button', { type: 'button', class: 'btn primary wake', onclick: () => wake(d) }, t('wake'))),
      lines.length ? h('details', { class: 'log' }, h('summary', {}, t('details', { n: w.attempt })), h('ul', {}, lines)) : null);
  }

  function renderDevices(list) {
    const grid = $('#grid');
    if (!grid) return;
    $('#empty').hidden = list.length > 0;
    const els = [];
    const seen = new Set();
    for (const d of list) {
      seen.add(d.id);
      const sig = JSON.stringify(d);
      let c = cards.get(d.id);
      if (!c || c.sig !== sig) {
        const el = buildCard(d);
        if (c) {
          const wasOpen = c.el.querySelector('details');
          const nowEl = el.querySelector('details');
          if (wasOpen && nowEl) nowEl.open = wasOpen.open;
        }
        c = { sig, el };
        cards.set(d.id, c);
      }
      els.push(c.el);
    }
    for (const id of [...cards.keys()]) if (!seen.has(id)) cards.delete(id);
    const same = grid.children.length === els.length && els.every((el, i) => grid.children[i] === el);
    if (!same) grid.replaceChildren(...els);
  }

  function notify(text) {
    const box = $('#flash');
    if (!box) return;
    box.textContent = text;
    setTimeout(() => { if (box.textContent === text) box.textContent = ''; }, 6000);
  }

  async function refresh() {
    if (!mainActive || busy) return;
    busy = true;
    try {
      const list = await api('GET', '/devices');
      hasWaking = list.some((d) => d.wake.state === 'waking');
      renderDevices(list);
    } catch (e) {
      if (e.status === 401) { busy = false; boot(); return; }
    } finally {
      busy = false;
    }
    clearTimeout(timer);
    if (mainActive && !document.hidden) timer = setTimeout(refresh, hasWaking ? 1000 : 4000);
  }

  async function wake(d) {
    try { await api('POST', `/devices/${encodeURIComponent(d.id)}/wake`); } catch (e) { notify(errText(e)); }
    refresh();
  }

  async function cancelWake(d) {
    try { await api('POST', `/devices/${encodeURIComponent(d.id)}/wake/cancel`); } catch (e) { notify(errText(e)); }
    refresh();
  }

  function enterMain() {
    mainActive = true;
    for (const id of ['#btnAdd', '#btnPassword', '#btnLogout']) $(id).hidden = false;
    setView(
      h('p', { class: 'err', id: 'flash', role: 'alert' }),
      h('section', { class: 'grid', id: 'grid' }),
      h('p', { class: 'empty', id: 'empty', hidden: true }, t('empty')));
    cards.clear();
    refresh();
  }

  function leaveMain() {
    mainActive = false;
    clearTimeout(timer);
    cards.clear();
    for (const id of ['#btnAdd', '#btnPassword', '#btnLogout']) $(id).hidden = true;
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

  async function openDevice(d) {
    editing = d || null;
    const f = deviceForm.elements;
    await fillInterfaces(d ? d.interface : '');
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
    if (s.authenticated) enterMain();
    else if (s.setupRequired) showSetup();
    else showLogin();
  }

  setView(h('p', { class: 'center' }, t('loading')));
  boot();
})();
