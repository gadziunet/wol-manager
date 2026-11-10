# WoL

Mała aplikacja webowa Wake-on-LAN do samodzielnego hostowania. Jeden statyczny plik wykonywalny, zero zewnętrznych zależności, dane w jednym pliku JSON. Wygodna na telefonie.

*English: [README.md](README.md)*

> Status: wczesna wersja (0.x). Nazwa jest robocza.

## Po co ta aplikacja

Budzenie komputera, który był wyłączony przez dłuższy czas, jest mniej niezawodne, niż się wydaje, a trzy zwykłe sposoby wysłania magic packet zachowują się w różnych sieciach różnie:

- **Unicast na IP komputera** dociera tylko wtedy, gdy router ma jeszcze wpis ARP uśpionego komputera. Po długiej przerwie wpisu już nie ma i pakiet nie ma dokąd pójść.
- **Broadcast kierowany** (np. `192.168.1.255`) zależy od poprawnej maski sieci i od tego, czy router w ogóle go przekaże.
- **Broadcast ogólny** (`255.255.255.255`) wysyła zwykłe polecenie `wakeonlan`. W wielu sieciach działa tam, gdzie pozostałe dwa zawodzą.

Aplikacja, która wysyła tylko jeden z tych sposobów i tylko raz, w niektórych sieciach zawodzi po cichu, a przyczynę trudno znaleźć. Ta aplikacja wysyła **wszystkie metody w każdej próbie**, **ponawia** próby, aż komputer odpowie albo skończy się czas, i **pokazuje dokładnie, co wysłano**, więc błąd można zdiagnozować zamiast zgadywać. Każde urządzenie ma własny wybór metod.

## Funkcje

- Budzenie po adresie MAC maksymalnie trzema metodami naraz: broadcast ogólny przez każdy interfejs sieciowy, broadcast kierowany, unicast.
- Ponawianie co kilka sekund, aż komputer wstanie (ping i/lub porty TCP) albo minie limit czasu.
- Postęp na żywo oraz zakładka **Logi** z każdą próbą, każdym celem pakietu i jego wynikiem.
- **Grupy** urządzeń z jednym przyciskiem, który budzi całą grupę, każde urządzenie według własnych ustawień.
- Wiele kont: administratorzy i zwykli użytkownicy, którzy mogą tylko budzić i widzieć przypisane im urządzenia i grupy.
- Stan online każdego urządzenia, odświeżany w tle.
- Ustawienia per urządzenie: metody, interfejs sieciowy, maska, porty do sprawdzania, limit czasu, odstęp ponawiania.
- Strona dopasowana do telefonu i komputera, jasny i ciemny motyw, polski i angielski.
- Zahashowane hasła, ograniczanie prób logowania, ścisłe nagłówki bezpieczeństwa.
- Jeden statyczny plik ok. 7 MB. Tylko biblioteka standardowa, więc nie ma nic do aktualizowania poza samym Go.

## Szybki start (Docker Compose)

```yaml
services:
  wol:
    build: .
    container_name: wol
    restart: unless-stopped
    network_mode: host
    read_only: true
    cap_drop: [ALL]
    cap_add: [CHOWN, SETUID, SETGID, NET_RAW]
    security_opt: ["no-new-privileges:true"]
    environment:
      WOL_ADDR: ":8080"
    volumes:
      - wol-data:/data

volumes:
  wol-data:
```

```sh
docker compose up -d --build
docker compose logs wol
```

`network_mode: host` jest wymagane. W domyślnej sieci mostkowej Dockera pakiety broadcast nigdy nie trafią do Twojej sieci.

Przy pierwszym uruchomieniu nie ma jeszcze konta. W logu pojawi się jednorazowy **token konfiguracji**:

```
FIRST RUN: no account exists yet. Open the web page and enter this setup token: ...
```

Otwórz `http://192.168.1.10:8080` (adres maszyny z kontenerem), wpisz token i wybierz nazwę użytkownika oraz hasło co najmniej 10-znakowe. Domyślnego hasła nie ma. Token jest tylko w pamięci i znika po utworzeniu konta.

Bez Dockera: `go build -o wol . && ./wol` (Go 1.24 lub nowsze).

## Dodawanie urządzenia

1. **Nazwa** i **adres MAC** są wymagane.
2. **Adres IP lub nazwa hosta** jest opcjonalny, ale zalecany. Służy do pakietów unicast i do sprawdzania, czy komputer działa. Nazwa jest rozwiązywana przez DNS, a ostatni znany adres jest zapamiętywany, więc budzenie działa też wtedy, gdy komputer jest wyłączony i nazwa już się nie rozwiązuje.
3. W sekcji **Zaawansowane** wybierzesz metody, interfejs sieciowy, maskę dla dodatkowego broadcastu kierowanego, ping i porty TCP do sprawdzania, limit czasu i odstęp ponawiania.

Bez hosta i sprawdzania aplikacja nie wie, czy komputer wstał. Wysyła wtedy trzy próby i zgłasza „wysłano”.

## Użytkownicy, grupy i logi

Pierwsze konto jest administratorem. Administrator zarządza wszystkim w zakładce **Ustawienia**:

- **Użytkownicy.** Dodajesz konta i nadajesz im rolę. *Użytkownik* może tylko budzić i widzieć stan urządzeń przypisanych mu bezpośrednio lub przez grupę oraz zmieniać własne hasło. Nie widzi adresów MAC ani IP, ustawień urządzeń ani celów pakietów. *Administrator* może wszystko. Zawsze zostaje co najmniej jeden administrator i nie można usunąć własnego konta. Nowe hasło ustawione przez administratora wylogowuje tego użytkownika wszędzie.
- **Grupy.** Urządzenie należy najwyżej do jednej grupy (wybierasz ją przy edycji urządzenia). Grupa ma przycisk **Wybudź grupę**, który uruchamia każde jej urządzenie z jego własnymi metodami, ponawianiem i limitem czasu. Zwykły użytkownik budzi tylko te urządzenia z grupy, do których ma dostęp.

Zakładka **Logi** pokazuje prośby o budzenie, każdą próbę z wysłanymi pakietami oraz wynik (wstał po N sekundach, brak odpowiedzi, zatrzymano). Administrator widzi też logowania, nieudane logowania (z adresem) i zmiany użytkowników, grup i urządzeń. Zwykły użytkownik widzi tylko zdarzenia budzenia własnych urządzeń. Log jest w pamięci (ostatnie 500 zdarzeń) i po restarcie zaczyna od zera.

## Konfiguracja

| Zmienna | Domyślnie | Znaczenie |
|---|---|---|
| `WOL_ADDR` | `:8080` | Adres nasłuchu, np. `127.0.0.1:8080` |
| `WOL_DATA` | `./data` (`/data` w obrazie) | Katalog z `state.json` |
| `WOL_COOKIE_SECURE` | `false` | Ustaw `true`, gdy strona jest po HTTPS |
| `WOL_TRUST_PROXY` | `false` | Ufaj `X-Forwarded-For` i `X-Forwarded-Host` z reverse proxy |
| `WOL_PORTS` | `9,7` | Porty UDP, na które idzie magic packet |
| `WOL_STATUS_INTERVAL` | `15s` | Jak często odświeżać stan online (minimum `5s`) |
| `WOL_UID` / `WOL_GID` | `65532` | Użytkownik i grupa, na które aplikacja przełącza się po starcie (nie może być 0) |

## Bezpieczeństwo

- Nie wystawiaj aplikacji bezpośrednio do internetu. Użyj reverse proxy z HTTPS albo dostępu przez VPN.
- Za proxy ustaw `WOL_COOKIE_SECURE=true` i `WOL_TRUST_PROXY=true` oraz nasłuchuj na adresie pętli zwrotnej (`WOL_ADDR=127.0.0.1:8080`). `WOL_TRUST_PROXY` włączaj tylko wtedy, gdy przed aplikacją stoi dokładnie jedno zaufane proxy, inaczej adres klienta da się podszyć i ograniczanie prób traci sens.
- Przykład z Caddy:

  ```
  wol.example.com {
      reverse_proxy 127.0.0.1:8080
  }
  ```
- Kontener na chwilę startuje jako root tylko po to, żeby przekazać `/data` nieuprzywilejowanemu użytkownikowi (`65532`), po czym przełącza się na niego, zanim otworzy jakikolwiek port. Nie zachowuje żadnych uprawnień i nie może odzyskać roota. Do tego kontener potrzebuje uprawnień `CHOWN`, `SETUID` i `SETGID` (ustawionych już w pliku Compose powyżej). Po starcie zostaje tylko `NET_RAW`: dzięki niemu sprawdzanie stanu może wysyłać ping bez żadnych ustawień jądra. Możesz je usunąć, a sprawdzanie użyje wtedy tylko portów TCP.
- Hasła są zapisane jako PBKDF2-HMAC-SHA256 (600 000 iteracji) z losową solą. Tokeny sesji są losowe, przechowywane tylko jako hashe i wygasają po 30 dniach. Ciasteczko ma `HttpOnly` i `SameSite=Strict`.
- Po 5 nieudanych logowaniach adres jest blokowany na 15 minut.
- Żądania zmieniające dane muszą być JSON-em z tego samego originu. Ścisła polityka CSP zabrania skryptów inline.
- Aplikacja nie uruchamia zewnętrznych poleceń i nie przechowuje żadnych danych dostępowych do budzonych komputerów.
- `state.json` ma uprawnienia `0600` i zawiera hashe haseł, więc traktuj go (i jego kopie zapasowe) jak sekret.

**Zapomniane hasło administratora:** zatrzymaj aplikację, ustaw w `state.json` `"users": []` i `"sessions": []`, uruchom ją ponownie i użyj nowego tokenu konfiguracji z logu. Wszystkie konta są usuwane, urządzenia i grupy zostają, a nowe konto jest administratorem.

## Sprawdzanie stanu

Aplikacja pinguje komputer (ICMP) i opcjonalnie łączy się z portami TCP z Twojej listy. Port TCP, który odpowiada „połączenie odrzucone”, też oznacza, że komputer działa, bo odmówić może tylko działająca maszyna.

W kontenerze ping działa od razu dzięki `NET_RAW` (patrz plik Compose). Bez tego uprawnienia działa tylko wtedy, gdy jądro pozwala na to użytkownikowi, jako którego działa aplikacja:

```sh
sysctl net.ipv4.ping_group_range
# jeśli zakres nie obejmuje użytkownika kontenera (65532):
sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
```

Gdy ping jest niedostępny, log o tym informuje i używane są porty TCP.

## Gdy komputer się nie budzi

1. Włącz Wake-on-LAN w BIOS/UEFI (nazwy bywają różne: „Power On By PCI-E”, „Resume by LAN”) i wyłącz tryby „ErP” lub „Deep Sleep”.
2. W systemie pozwól karcie sieciowej wybudzać komputer. W Windows to zakładka *Zarządzanie energią* karty, a „Szybkie uruchamianie” powinno być wyłączone.
3. Używaj kabla. Budzenie przez Wi-Fi działa tylko na niektórych kartach.
4. Po kliknięciu *Wybudź* otwórz zakładkę **Logi**. Rozwiń *Cele pakietów* przy próbie, żeby zobaczyć każdy wypróbowany cel. Wypróbuj metody po kolei w ustawieniach zaawansowanych, żeby znaleźć tę, która działa w Twojej sieci.
5. Upewnij się, że aplikacja działa w tej samej sieci co komputer. Pakiety broadcast nie przechodzą przez routery, o ile router nie jest skonfigurowany do ich przekazywania.
6. Jeśli w logu jest `permission denied` dla `/data` albo błąd `setuid`/`setgid`, kontenerowi brakuje tych trzech uprawnień (`cap_add: [CHOWN, SETUID, SETGID, NET_RAW]` razem z `cap_drop: [ALL]`).

## Ograniczenia

- Tylko IPv4.
- Brak rozwiązywania nazw mDNS (`.local`). Nazwy DNS i adresy IP działają.
- Tylko dwie role (administrator i użytkownik). Konta zarządza się w aplikacji, nie ma zewnętrznego logowania (LDAP, SSO).
- Brak wbudowanego HTTPS, użyj reverse proxy.

## Wiersz poleceń

```sh
wol                          # uruchamia aplikację webową
wol send AA:BB:CC:DD:EE:FF   # wysyła jeden magic packet na 255.255.255.255:9
wol send -addr 192.168.1.255:9 AA:BB:CC:DD:EE:FF
wol healthcheck              # kod wyjścia 0, gdy aplikacja odpowiada
wol version
```

## Rozwój

```sh
go vet ./...
go test -race ./...
WOL_DATA=./data go run .
```

Kod jest podzielony według odpowiedzialności: `wake.go` (metody, ponawianie), `probe.go` i `icmp*.go` (sprawdzanie stanu), `store.go` (plik JSON), `auth.go` (hasła, ograniczanie prób), `server.go` (API HTTP i nagłówki bezpieczeństwa), `web/` (strona wbudowana w plik wykonywalny).

## Licencja

Jeszcze nie wybrana.
