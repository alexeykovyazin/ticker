# ticker

Windows command-line app and Windows Service that periodically inserts `CURRENT_TIMESTAMP` into a Firebird 3+ database.

Every interval it: **connects → ensures the target table exists → inserts one row → commits → disconnects**. Insert failures are logged and the loop continues.

## Configuration

Place `config.json` next to `ticker.exe`. The app always loads that file from the executable directory.

| Field | Required | Default | Description |
| --- | --- | --- | --- |
| `host` | yes | — | Firebird host |
| `port` | no | `3050` | Firebird port |
| `database` | yes | — | Database path on the Firebird server |
| `user` | yes | — | Database user |
| `password` | yes | — | Database password |
| `table` | no | `TICKS` | Table name (unquoted Firebird identifier) |
| `column` | no | `TICK` | Timestamp column name (unquoted Firebird identifier) |
| `interval_ms` | no | `1000` | Tick interval in milliseconds |
| `service.name` | no | `Ticker` | Windows service name |
| `service.display_name` | no | `Ticker` | Windows service display name |
| `service.description` | no | *(see sample)* | Windows service description |

Table and column names must match `^[A-Z][A-Z0-9_]*$` after uppercasing (letters, digits, underscore; start with a letter).

If the table does not exist, it is created as:

```sql
CREATE TABLE <table> (<column> TIMESTAMP)
```

Each tick inserts:

```sql
INSERT INTO <table> (<column>) VALUES (CURRENT_TIMESTAMP)
```

### Sample config

This repo includes `config.json` and a sample Firebird database `EMPLOYEE.FDB`. Point `database` at a path Firebird can open (for example copy `EMPLOYEE.FDB` into a Firebird-accessible folder and set that path), then set `host` / `port` / credentials for your server.

```json
{
  "host": "127.0.0.1",
  "port": 3050,
  "database": "C:/data/EMPLOYEE.FDB",
  "user": "SYSDBA",
  "password": "masterkey",
  "table": "TICKS",
  "column": "TICK",
  "interval_ms": 1000,
  "service": {
    "name": "Ticker",
    "display_name": "Ticker",
    "description": "Inserts timestamps into Firebird at a fixed interval"
  }
}
```

## Build

```powershell
go build -o ticker.exe .\cmd\ticker
```

## Run (foreground)

```powershell
.\ticker.exe run
```

Stop with `Ctrl+C`.

## Install / run as Windows Service

Installing a service typically requires an elevated (Administrator) shell.

```powershell
.\ticker.exe install
.\ticker.exe start
```

Stop and uninstall:

```powershell
.\ticker.exe stop
.\ticker.exe uninstall
```

`install` registers the executable with the Windows Service Control Manager using the hidden `service` argument (used by SCM; not intended for interactive use). Service name and metadata come from the `service` section in `config.json`.
