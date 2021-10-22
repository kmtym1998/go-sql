package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

// RootCmd is root command
var RootCmd = &cobra.Command{
	Use:   "go-sql",
	Short: "接続情報が与えられたDBに対してSQLを実行する",
	Long:  "接続情報が与えられたDBに対してSQLを実行する。\nSQLファイルまでのパスを指定するとそれを単体実行。\nSQLファイルのあるディレクトリを指定するとその配下のSQLファイルを全実行する。",
	Run:   callback,
}

type Cfg struct {
	DSN []DSNCfg `json:"dsn"`
}

type DSNCfg struct {
	Name     string `json:"name"`
	Driver   string `json:"driver"`
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	DBName   string `json:"db_name"`
	SSLMode  string `json:"ssl_mode"`
}

// コマンド実行時に最初に呼ばれる初期化処理
func init() {
	// フラグの定義
	// 第1引数: フラグ名、第2引数: 省略したフラグ名
	// 第3引数: デフォルト値、第4引数: フラグの説明
	RootCmd.Flags().StringP("target", "t", "", "[必須] SQLファイルまでのpath or SQLファイルが入ったディレクトリまでのpath")
	RootCmd.Flags().StringP("database-url", "d", "", "[必須] DB接続情報 (postgresとmysqlしか対応してない)")
	RootCmd.Flags().StringP("config", "c", "", "[任意] 設定JSONファイルのパス (.go-gql.json)")
}

// 必須バリデーション
func validate(argName string, argVal string) error {
	if argVal == "" {
		return fmt.Errorf("%sを入力してください", argName)
	}

	return nil
}

// pathからファイルの中身を文字列で取得する
func getContent(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	bf, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	v := string(bf)

	return v, nil
}

// コマンドの中身の処理
func callback(cmd *cobra.Command, args []string) {
	// NOTE: dsnの読み取り優先順位
	// 1. コマンドラインからの入力値
	// 2. .go-sql.jsonの値
	// 3. 環境変数 GO_SQL_DATABASE_URL の値
	var dsn string
	var dsnTmp string
	dsn = os.Getenv("GO_SQL_DATABASE_URL")

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		fmt.Printf("config取得エラー: %v\n", err)
		os.Exit(1)
	}
	// config のJSONを読む
	if configPath != "" {
		raw, err := ioutil.ReadFile(configPath)
		if err != nil {
			fmt.Printf("設定ファイルを取得できませんでした: %v", err.Error())
			os.Exit(1)
		}
		var cfg Cfg
		if err := json.Unmarshal(raw, &cfg); err != nil {
			fmt.Printf("設定ファイルの型が不正です: %v", err.Error())
			os.Exit(1)
		}

		for _, dsn := cfg.DSN{

		}
	}

	// flagの読み取り
	target, err := cmd.Flags().GetString("target")
	if err != nil {
		fmt.Printf("target取得エラー: %v\n", err)
		os.Exit(1)
	}
	dsnTmp, err = cmd.Flags().GetString("database-url")
	if err != nil {
		fmt.Printf("database-url取得エラー: %v\n", err)
		os.Exit(1)
	}
	if dsnTmp != "" {
		dsn = dsnTmp
	}

	// flagの値の必須チェック
	if err := validate("target", target); err != nil {
		fmt.Printf("バリデーションエラー: %v\n", err)
		os.Exit(1)
	}
	if err := validate("database-url", dsn); err != nil {
		fmt.Printf("バリデーションエラー: %v\n", err)
		os.Exit(1)
	}

	// 実行予定のクエリが入る
	var queries []string
	ext := filepath.Ext(target)
	if ext == ".sql" {
		// targetの拡張子が.sqlだったらそのままファイルを読む
		query, err := getContent(target)
		if err != nil {
			fmt.Printf("ファイル読み取りエラー: %v\n", err)
			os.Exit(1)
		}
		queries = append(queries, query)
	} else if ext == "" {
		files, _ := ioutil.ReadDir(target)
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			query, err := getContent(target + "/" + f.Name())
			if err != nil {
				fmt.Printf("ファイル読み取りエラー: %v\n", err)
				os.Exit(1)
			}
			queries = append(queries, query)
		}
	} else {
		fmt.Println("不正な拡張子")
		os.Exit(1)
	}

	// DATABASE URLをパース
	u, err := url.Parse(dsn)
	if err != nil {
		fmt.Printf("URLのパースエラー: %v\n", err)
		os.Exit(1)
	}

	// DB接続
	db, err := sql.Open(u.Scheme, dsn)
	if err != nil {
		fmt.Printf("DB接続失敗: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// クエリの実行
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("トランザクション開始時エラー: %v", err)
		os.Exit(1)
	}
	for _, q := range queries {
		_, err := tx.Query(q)
		if err != nil {
			fmt.Printf("クエリ: %v\n", q)
			fmt.Printf("クエリ実行エラー: %v\n", err)
			if err := tx.Rollback(); err != nil {
				fmt.Printf("ロールバックエラー: %v\n", err)
			}
			os.Exit(1)
		}
	}

	if err := tx.Commit(); err != nil {
		fmt.Printf("トランザクションコミットエラー: %v", err)
	}
	fmt.Println("完了🦩")
}
