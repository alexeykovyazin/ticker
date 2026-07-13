# ticker

Windows command line app + service that periodically inserts `CURRENT_TIMESTAMP` into a Firebird database.

## Configuration

Place `config.json` next to `ticker.exe` (this repo includes a sample).

## Build

```powershell
go build -o ticker.exe .\cmd\ticker
```

## Run (foreground)

```powershell
.\ticker.exe run
```

## Install/run as Windows Service

> Installing a service typically requires an elevated (Administrator) shell.

```powershell
.\ticker.exe install
.\ticker.exe start
```

Stop and uninstall:

```powershell
.\ticker.exe stop
.\ticker.exe uninstall
```

