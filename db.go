package toolkit

import (
	"database/sql"
	"fmt"
	"time"

	//
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// DbConfig config
type DbConfig struct {
	Driver       string
	DNS          string
	MaxOpenConns int
	MaxIdle      int
	MaxLifetime  time.Duration
}

var (
	config DbConfig
)

// DB define
type DB struct {
	conn *sqlx.DB
	tx   *sqlx.Tx
}

// SetDbConfig set
func SetDbConfig(cfg DbConfig) {
	config.Driver = cfg.Driver
	config.DNS = cfg.DNS
	config.MaxOpenConns = cfg.MaxOpenConns
	config.MaxIdle = cfg.MaxIdle
	config.MaxLifetime = cfg.MaxLifetime * time.Second
}

// NewDB new DB object
func NewDB() *DB {
	return &DB{}
}

// NoRows check norows error
func (d *DB) NoRows(err error) bool {
	if err == sql.ErrNoRows {
		return true
	}
	return false
}

// Connect connect to database
func (d *DB) Connect() (err error) {
	d.conn, err = sqlx.Connect(config.Driver, config.DNS)
	//*
	d.conn.DB.SetMaxOpenConns(config.MaxOpenConns)
	d.conn.DB.SetMaxIdleConns(config.MaxIdle)
	d.conn.DB.SetConnMaxLifetime(config.MaxLifetime)
	// */
	return
}

// Close close database connect
func (d *DB) Close() {
	d.conn.Close()
}

// BeginTrans begin trans
func (d *DB) BeginTrans() {
	d.tx = d.conn.MustBegin()
}

// Commit commit
func (d *DB) Commit() {
	d.tx.Commit()
}

// Rollback rollback
func (d *DB) Rollback() {
	d.tx.Rollback()
}

// TransExec trans execute
func (d *DB) TransExec(
	query string,
	args interface{}) (LastInsertId, RowsAffected int64, err error) {
	if rs, err := d.tx.NamedExec(query, args); err == nil {
		RowsAffected, _ = rs.RowsAffected()
		LastInsertId, _ = rs.LastInsertId()
	}
	return
}

// Rows get rows
func (d *DB) Rows(dest interface{}, query string, args interface{}) error {
	err := d.Connect()
	if err != nil {
		return err
	}
	defer d.conn.Close()

	nstmt, err := d.conn.PrepareNamed(query)
	if err != nil {
		return err
	}
	defer nstmt.Close()

	err = nstmt.Select(dest, args)

	return err
}

// Row get row
func (d *DB) Row(dest interface{}, query string, args interface{}) error {
	err := d.Connect()
	if err != nil {
		return err
	}
	defer d.conn.Close()

	nstmt, err := d.conn.PrepareNamed(query)
	if err != nil {
		return err
	}
	defer nstmt.Close()

	err = nstmt.Get(dest, args)

	return err
}

// Insert insert into
func (d *DB) Insert(
	query string,
	args interface{}) (LastInsertId, RowsAffected int64, err error) {
	err = d.Connect()
	if err != nil {
		return
	}
	defer d.conn.Close()

	if rs, err := d.conn.NamedExec(query, args); err == nil {
		LastInsertId, _ = rs.LastInsertId()
		RowsAffected, _ = rs.RowsAffected()
	}
	return
}

// Update update/delete
func (d *DB) Update(
	query string,
	args interface{}) (RowsAffected int64, err error) {
	err = d.Connect()
	if err != nil {
		return
	}
	defer d.conn.Close()

	if rs, err := d.conn.NamedExec(query, args); err == nil {
		RowsAffected, _ = rs.RowsAffected()
	}
	return
}

// Limit MySQL limit
func (d *DB) Limit(page, pagesize int) string {
	return fmt.Sprintf(" limit %d, %d", (page-1)*pagesize, pagesize)
}
