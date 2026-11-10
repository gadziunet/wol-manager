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
- Postęp na żywo i lista każdego celu pakietu z wynikiem w każdej próbie.
- Stan online każdego urządzenia, odświeżany w tle.
- Ustawienia per urządzenie: metody, interfejs sieciowy, maska, porty do sprawdzania, limit czasu, odstęp ponawiania.
- Strona dopasowana do telefonu i komputera, jasny i ciemny motyw, polski i angielski.
- Konto z zahashowanym hasłem, ograniczanie prób logowania, ścisłe nagłówki bezpieczeństwa.
- Jeden statyczny plik ok. 7 MB. Tylko biblioteka standardowa, więc nie ma nic do aktualizowania poza samym Go.

## Szybki start (Docker Compose)

```yaml
services:
  wol:
    build: .
    container_name: wol
    restart: unless-stopped
    network_mode: host
    user: "65532:65532"
    read_only: true
    cap_drop: [ALL]
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

## Konfiguracja

| Zmienna | Domyślnie | Znaczenie |
|---|---|---|
| `WOL_ADDR` | `:8080` | Adres nasłuchu, np. `127.0.0.1:8080` |
| `WOL_DATA` | `./data` (`/data` w obrazie) | Katalog z `state.json` |
| `WOL_COOKIE_SECURE` | `false` | Ustaw `true`, gdy strona jest po HTTPS |
| `WOL_TRUST_PROXY` | `false` | Ufaj `X-Forwarded-For` i `X-Forwarded-Host` z reverse proxy |
| `WOL_PORTS` | `9,7` | Porty UDP, na które idzie magic packet |
| `WOL_STATUS_INTERVAL` | `15s` | Jak często odświeżać stan online (minimum `5s`) |

## Bezpieczeństwo

- Nie wystawiaj aplikacji bezpośrednio do internetu. Użyj reverse proxy z HTTPS albo dostępu przez VPN.
- Za proxy ustaw `WOL_COOKIE_SECURE=true` i `WOL_TRUST_PROXY=true` oraz nasłuchuj na adresie pętli zwrotnej (`WOL_ADDR=127.0.0.1:8080`). `WOL_TRUST_PROXY` włączaj tylko wtedy, gdy przed aplikacją stoi dokładnie jedno zaufane proxy, inaczej adres klienta da się podszyć i ograniczanie prób traci sens.
- Przykład z Caddy:

  ```
  wol.example.com {
      reverse_proxy 127.0.0.1:8080
  }
  ```
- Hasła są zapisane jako PBKDF2-HMAC-SHA256 (600 000 iteracji) z losową solą. Tokeny sesji są losowe, przechowywane tylko jako hashe i wygasają po 30 dniach. Ciasteczko ma `HttpOnly` i `SameSite=Strict`.
- Po 5 nieudanych logowaniach adres jest blokowany na 15 minut.
- Żądania zmieniające dane muszą być JSON-em z tego samego originu. Ścisła polityka CSP zabrania skryptów inline.
- Aplikacja nie uruchamia zewnętrznych poleceń i nie przechowuje żadnych danych dostępowych do budzonych komputerów.
- `state.json` ma uprawnienia `0600` i zawiera hashe haseł, więc traktuj go (i jego kopie zapasowe) jak sekret.

**Zapomniane hasło:** zatrzymaj aplikację, ustaw w `state.json` `"users": []` i `"sessions": []`, uruchom ją ponownie i użyj nowego tokenu konfiguracji z logu. Urządzenia zostają.

## Sprawdzanie stanu

Aplikacja pinguje komputer (ICMP) i opcjonalnie łączy się z portami TCP z Twojej listy. Port TCP, który odpowiada „połączenie odrzucone”, też oznacza, że komputer działa, bo odmówić może tylko działająca maszyna.

Ping działa bez uprawnień na Linuksie, jeśli jądro na to pozwala użytkownikowi, jako którego działa aplikacja:

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
4. Po kliknięciu *Wybudź* otwórz **Szczegóły** na karcie urządzenia. Widać tam każdy wypróbowany cel. Wypróbuj metody po kolei w ustawieniach zaawansowanych, żeby znaleźć tę, która działa w Twojej sieci.
5. Upewnij się, że aplikacja działa w tej samej sieci co komputer. Pakiety broadcast nie przechodzą przez routery, o ile router nie jest skonfigurowany do ich przekazywania.

## Ograniczenia

- Tylko IPv4.
- Brak rozwiązywania nazw mDNS (`.local`). Nazwy DNS i adresy IP działają.
- Jedno konto, tworzone przy pierwszym uruchomieniu. Hasło można zmienić w aplikacji.
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
