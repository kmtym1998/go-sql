package cmd

import (
	"database/sql"
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

// コマンド実行時に最初に呼ばれる初期化処理
func init() {
	// フラグの定義
	// 第1引数: フラグ名、第2引数: 省略したフラグ名
	// 第3引数: デフォルト値、第4引数: フラグの説明
	RootCmd.Flags().StringP("target", "t", "", "[必須] SQLファイルまでのpath or SQLファイルが入ったディレクトリまでのpath")
	RootCmd.Flags().StringP("database-url", "d", "", "[必須] DB接続情報 (postgresとmysqlしか対応してない)")
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
	// flagの読み取り
	target, err := cmd.Flags().GetString("target")
	if err != nil {
		fmt.Printf("target取得エラー: %v\n", err)
		os.Exit(1)
	}
	dsn, err := cmd.Flags().GetString("database-url")
	if err != nil {
		fmt.Printf("database-url取得エラー: %v\n", err)
		os.Exit(1)
	}

	// flagの値の必須チェック
	if err := validate("target", target); err != nil {
		fmt.Printf("バリデーションエラー: %v\n", err)
		os.Exit(1)
	}
	if validate("database-url", dsn); err != nil {
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
