# go-sql

## TODO

- リリースの自動化

## Usage

```
接続情報が与えられたDBに対してSQLを実行する。
SQLファイルまでのパスを指定するとそれを単体実行。
SQLファイルのあるディレクトリを指定するとその配下のSQLファイルを全実行する。

Usage:
  go-sql [flags]

Flags:
  -d, --database-url string   [必須] DB接続情報 (postgresとmysqlしか対応してない)
  -h, --help                  help for go-sql
  -t, --target string         [必須] SQLファイルまでのpath or SQLファイルが入ったディレクトリまでのpath
```

## Example

```
$ go-sql -d "postgres://postgres:secret@localhost:5432?sslmode=disable" -t ~/sql
```
