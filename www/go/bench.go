package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/profile"
)

// DB取得
func getDB() *sql.DB {

	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}
	name := os.Getenv("DB_NAME")
	if name == "" {
		name = "test"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("DB_PASSWORD")
	if password != "" {
		password = ":" + password
	}

	dsn := fmt.Sprintf(
		"%s%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&interpolateParams=true",
		user, password, host, port, name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	return db
}

// 初期化
func initialize(db *sql.DB) {
	var cmd string

	cmd = `
		CREATE TABLE IF NOT EXISTS users (
			id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
			name varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
			email varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
			email_verified_at timestamp NULL DEFAULT NULL,
			password varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
			remember_token varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
			created_at timestamp NULL DEFAULT NULL,
			updated_at timestamp NULL DEFAULT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY users_email_unique (email)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`
	_, err := db.Exec(cmd)
	if err != nil {
		panic(err)
	}

	cmd = "TRUNCATE users;"
	_, err = db.Exec(cmd)
	if err != nil {
		panic(err)
	}

	// var cnt int
	// err = db.QueryRow("select count(*) as cnt from users;").Scan(&cnt)
	// println(cnt)
}

// User ユーザー情報
type User struct {
	ID              int
	Name            string
	Email           string
	EmailVerifiedAt time.Time
	Password        string
	RememberToken   string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func work1(db *sql.DB) {
	var (
		err    error
		file   *os.File
		reader *csv.Reader
		lines  []string
	)

	// CSV読み込み
	//printTime("import CSV start")
	file, err = os.Open("../import_users.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader = csv.NewReader(file)
	_, err = reader.Read() // ヘッダースキップ
	for {
		lines, err = reader.Read()
		if err != nil {
			break
		}

		// lines[0]はidのため1から
		_, err = db.Exec(`
			INSERT INTO users (
				name,
				email,
				email_verified_at,
				password,
				remember_token,
				created_at,
				updated_at
			) values (
				?, ?, ?, ?, ?, ?, ?
			);
		`, lines[1], lines[2], lines[3], lines[4], lines[5], lines[6], lines[7])
		if err != nil {
			panic(err)
		}
	}
	//printTime("import CSV end")
}

// CSV読み込み＆DBインサート＆CSV書き出し
func work(db *sql.DB) {
	var (
		err  error
		file *os.File
		rows *sql.Rows
	)

	rows, err = db.Query("select * from users order by id")
	if err != nil {
		panic(err)
	}

	// CSV書き出し
	//printTime("export CSV start")

	file, err = os.Create("./export_users.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	w := bufio.NewWriterSize(file, 512*1024)
	defer w.Flush()

	// ヘッダー書き込み
	_, err = w.WriteString(`"id","name","email","email_verified_at","password","remember_token","created_at","updated_at"` + "\n")
	if err != nil {
		panic(err)
	}

	var user User
	for rows.Next() {
		err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.EmailVerifiedAt,
			&user.Password,
			&user.RememberToken,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			panic(err)
		}

		_, err = fmt.Fprintf(
			w,
			`%d,"%s","%s","%s","%s","%s","%s","%s"`+"\n",
			user.ID,
			user.Name,
			user.Email,
			user.EmailVerifiedAt.Format("2006-01-02 15:04:05"),
			user.Password,
			user.RememberToken,
			user.CreatedAt.Format("2006-01-02 15:04:05"),
			user.UpdatedAt.Format("2006-01-02 15:04:05"))
		if err != nil {
			panic(err)
		}
	}
	//printTime("export CSV end")
}

// 実行時間を出力
func printTime(message string) {
	println(time.Now().Format("2006-01-02 15:04:05.000000") + " " + message)
}

// メイン処理
func main() {
	db := getDB()
	defer db.Close()

	//initialize(db)
	//work1(db)

	defer profile.Start().Stop()

	start := time.Now()
	for i := 1; i <= 100; i++ {
		// 負荷処理
		work(db)
	}
	fmt.Printf("t=%v\n", time.Now().Sub(start))
}
