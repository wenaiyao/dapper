package dapper

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	_ "github.com/ziutek/mymysql/godrv"
)

const (
	testDBName = "dapper_test"
	testDBUser = "dapper"
	testDBPass = "dapper"
)

type tweet struct {
	Id       int64     `dapper:"id,primarykey,serial"`
	UserId   int64     `dapper:"user_id"`
	Message  string    `dapper:"message"`
	Retweets int64     `dapper:"retweets"`
	Created  time.Time `dapper:"created"`
}

type tweetById struct {
	Id int64
}

type tweetByUserId struct {
	UserId int64
}

type tweetByUserAndMinRetweets struct {
	UserId      int64
	NumRetweets int64
}

type sampleQuery struct {
	Id          int64 `dapper:"id,primarykey,autoincrement"`
	Ignore      string `dapper:"-"`
	UserId      int64
}

func (t *tweet) String() string {
	return fmt.Sprintf("tweet[Id=%v,UserId=%v,Message=%v,Retweets=%v,Created=%v]",
		t.Id, t.UserId, t.Message, t.Retweets, t.Created)
}

type user struct {
	Id        int64    `dapper:"id,primarykey,serial"`
	Name      string   `dapper:"name"`
	Karma     *float64 `dapper:"karma"`
	Suspended bool     `dapper:"suspended"`
}

func (u *user) String() string {
	return fmt.Sprintf("user[Id=%v,Name=%v,Karma=%v,Suspended=%v]",
		u.Id, u.Name, u.Karma, u.Suspended)
}

func setup(t *testing.T) *sql.DB {
	connectionString := fmt.Sprintf("%s/%s/%s", testDBName, testDBUser, testDBPass)
	db, err := sql.Open("mymysql", connectionString)
	if err != nil {
		t.Fatalf("error connection to database: %v", err)
	}

	// Drop tables
	_, err = db.Exec("DROP TABLE IF EXISTS tweets")
	if err != nil {
		t.Fatalf("error dropping tweets table: %v", err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS users")
	if err != nil {
		t.Fatalf("error dropping users table: %v", err)
	}

	// Create tables
	_, err = db.Exec(`
CREATE TABLE users (
        id int(11) not null auto_increment,
        name varchar(100) not null,
        karma decimal(19,5),
        suspended tinyint(1) default '0',
        primary key (id)
)`)
	if err != nil {
		t.Fatalf("error creating users table: %v", err)
	}

	_, err = db.Exec(`
CREATE TABLE tweets (
        id int(11) not null auto_increment,
        user_id int(11) not null,
        message text,
        retweets int,
        created timestamp not null default current_timestamp,
        primary key (id),
        foreign key (user_id) references users (id) on delete cascade
)`)
	if err != nil {
		t.Fatalf("error creating tweets table: %v", err)
	}

	// Insert seed data
	_, err = db.Exec("INSERT INTO users (id,name) VALUES (1, 'Oliver')")
	if err != nil {
		t.Fatalf("error inserting user: %v", err)
	}
	_, err = db.Exec("INSERT INTO users (id,name) VALUES (2, 'Sandra')")
	if err != nil {
		t.Fatalf("error inserting user: %v", err)
	}

	_, err = db.Exec("INSERT INTO tweets (id,user_id,message,retweets) VALUES (1, 1, 'Google Go rocks', 179)")
	if err != nil {
		t.Fatalf("error inserting tweet: %v", err)
	}
	_, err = db.Exec("INSERT INTO tweets (id,user_id,message,retweets) VALUES (2, 1, '... so does Google Maps', 19)")
	if err != nil {
		t.Fatalf("error inserting tweet: %v", err)
	}
	_, err = db.Exec("INSERT INTO tweets (id,user_id,message,retweets) VALUES (3, 2, 'Holidays! Yay!', 1)")
	if err != nil {
		t.Fatalf("error inserting tweet: %v", err)
	}

	return db
}

func TestTypeCache(t *testing.T) {
	db := setup(t)
	defer db.Close()

	if len(typeCache) != 0 {
		t.Errorf("expected type cache to be empty, got %d entries", len(typeCache))
	}

	// Test typeInfo
	ti, err := AddType(reflect.TypeOf(sampleQuery{}))
	if err != nil {
		t.Errorf("error adding type sampleQuery: %v", err)
	}
	if ti == nil {
		t.Errorf("expected to return typeInfo, got nil")
	}
	if len(ti.FieldNames) != 3 {
		t.Errorf("expected typeInfo to have %d fields, got %d", 3, len(ti.FieldNames))
	}

	// Test field Id
	fi, found := ti.FieldInfos["Id"]
	if !found {
		t.Errorf("expected typeInfo to have an Id field")
	}
	if fi.FieldName != "Id" {
		t.Errorf("expected field Id to have name: Id")
	}
	if fi.ColumnName != "id" {
		t.Errorf("expected field Id to have column name: id (lower-case)")
	}
	if !fi.IsPrimaryKey {
		t.Errorf("expected field Id to be primary key")
	}
	if !fi.IsAutoIncrement {
		t.Errorf("expected field Id to be auto-increment")
	}
	if fi.IsTransient {
		t.Errorf("expected field Id to not be transient")
	}

	// Test field UserId
	fi, found = ti.FieldInfos["UserId"]
	if !found {
		t.Errorf("expected typeInfo to have a UserId field")
	}
	if fi.FieldName != "UserId" {
		t.Errorf("expected field UserId to have name: UserId")
	}
	if fi.ColumnName != "UserId" {
		t.Errorf("expected field UserId to have column name: User")
	}
	if fi.IsPrimaryKey {
		t.Errorf("expected field UserId to not be primary key")
	}
	if fi.IsAutoIncrement {
		t.Errorf("expected field UserId to not be auto-increment")
	}
	if fi.IsTransient {
		t.Errorf("expected field UserId to not be transient")
	}

	// Test field Ignore
	fi, found = ti.FieldInfos["Ignore"]
	if !found {
		t.Errorf("expected typeInfo to have an Ignore field")
	}
	if fi.FieldName != "Ignore" {
		t.Errorf("expected field Ignore to have name: Ignore")
	}
	if fi.ColumnName != "" {
		t.Errorf("expected field Ignore to have an empty column name")
	}
	if fi.IsPrimaryKey {
		t.Errorf("expected field Ignore to not be primary key")
	}
	if fi.IsAutoIncrement {
		t.Errorf("expected field Ignore to not be auto-increment")
	}
	if !fi.IsTransient {
		t.Errorf("expected field Ignore to be transient")
	}
}

func TestFirst(t *testing.T) {
	db := setup(t)
	defer db.Close()
	
	in := user{Id: 1}
	out := user{}
	err := First(db, "select * from users where id=:Id", in, &out)
	if err != nil {
		t.Fatalf("error on First: %v", err)
	}
	if out.Id != 1 {
		t.Errorf("expected user.Id == %d, got %d", 1, out.Id)
	}
	if out.Name != "Oliver" {
		t.Errorf("expected user.Name == %s, got %s", "Oliver", out.Name)
	}
	if out.Karma != nil {
		t.Errorf("expected user.Karma == nil, got %v", out.Karma)
	}
	if out.Suspended {
		t.Errorf("expected user.Suspended == %v, got %v", false, out.Suspended)
	}
}

func TestFirstWithProjection(t *testing.T) {
	db := setup(t)
	defer db.Close()
	
	in := user{Id: 1}
	out := user{}
	err := First(db, "select name from users where id=:Id", in, &out)
	if err != nil {
		t.Fatalf("error on First: %v", err)
	}
	if out.Id != 0 {
		t.Errorf("expected user.Id == %d, got %d", 0, out.Id)
	}
	if out.Name != "Oliver" {
		t.Errorf("expected user.Name == %s, got %s", "Oliver", out.Name)
	}
	if out.Karma != nil {
		t.Errorf("expected user.Karma == nil, got %v", out.Karma)
	}
	if out.Suspended {
		t.Errorf("expected user.Suspended == %v, got %v", false, out.Suspended)
	}
}

func TestQuery(t *testing.T) {
	db := setup(t)
	defer db.Close()

	results, err := Query(db, "select * from users order by name", nil, reflect.TypeOf(user{}))
	if err != nil {
		t.Fatalf("error on Query: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected len(results) == %d, got %d", 2, len(results))
	}
	for _, result := range results {
		user, ok := result.(*user)
		if !ok {
			t.Errorf("expected user as result")
		}
		if user.Id <= 0 {
			t.Errorf("expected user to have an Id > 0, got %d", user.Id)
		}
	}
}

func TestQueryWithProjections(t *testing.T) {
	db := setup(t)
	defer db.Close()

	results, err := Query(db, "select id,name from users order by name", nil, reflect.TypeOf(user{}))
	if err != nil {
		t.Fatalf("error on Query: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected len(results) == %d, got %d", 2, len(results))
	}
	for _, result := range results {
		user, ok := result.(*user)
		if !ok {
			t.Errorf("expected user as result")
		}
		if user.Id <= 0 {
			t.Errorf("expected user to have an Id > 0, got %d", user.Id)
		}
	}
}