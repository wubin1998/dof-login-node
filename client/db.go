package main

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-sql-driver/mysql"
)

const (
	dbUser = "game"
	dbName = "mysql"
)

func (a *App) dsn(c Config) string {
	mc := mysql.NewConfig()
	mc.User = dbUser
	mc.Passwd = c.DBPassword
	mc.Net = "tcp"
	mc.Addr = net.JoinHostPort(c.Server, strconv.Itoa(c.Port))
	mc.DBName = dbName
	mc.Collation = "utf8_general_ci"
	mc.Timeout = 5 * time.Second
	mc.ReadTimeout = 12 * time.Second
	mc.WriteTimeout = 12 * time.Second
	mc.AllowNativePasswords = true
	mc.AllowOldPasswords = true
	return mc.FormatDSN()
}

func (a *App) openDB(ctx context.Context, c Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", a.dsn(c))
	if err != nil {
		return nil, friendlyDBError(err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(2 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, friendlyDBError(err)
	}
	return db, nil
}

// Login returns the account uid, or -1 when the credentials are wrong.
func (a *App) Login(ctx context.Context, user, password string) (int, error) {
	cfg := a.store.Get()
	db, err := a.openDB(ctx, cfg)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var uid int64
	err = db.QueryRowContext(ctx,
		"select uid from d_taiwan.accounts where accountname=? and password=?",
		user, md5Password(password),
	).Scan(&uid)
	if errors.Is(err, sql.ErrNoRows) {
		return -1, nil
	}
	if err != nil {
		return 0, friendlyDBError(err)
	}

	if _, err := db.ExecContext(ctx,
		"update d_taiwan.limit_create_character set count=0 where m_id=?", uid,
	); err != nil {
		return 0, friendlyDBError(err)
	}
	return int(uid), nil
}

// Register creates an account and all of the side rows the game server needs.
// It returns false when the account name is already taken.
func (a *App) Register(ctx context.Context, user, password string) (bool, error) {
	cfg := a.store.Get()
	db, err := a.openDB(ctx, cfg)
	if err != nil {
		return false, err
	}
	defer db.Close()

	var existing int64
	err = db.QueryRowContext(ctx, "select uid from d_taiwan.accounts where accountname=?", user).Scan(&existing)
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, friendlyDBError(err)
	}

	var maxUID sql.NullInt64
	if err := db.QueryRowContext(ctx, "select max(uid) from d_taiwan.accounts").Scan(&maxUID); err != nil {
		return false, friendlyDBError(err)
	}
	uid := int64(1)
	if maxUID.Valid {
		uid = maxUID.Int64 + 1
	}

	md5 := md5Password(password)
	steps := []struct {
		query string
		args  []any
	}{
		{`insert into d_taiwan.accounts (uid,accountname,password) values (?,?,?)`, []any{uid, user, md5}},
		{`insert into d_taiwan.limit_create_character (m_id) values (?)`, []any{uid}},
		{`insert into d_taiwan.member_info (m_id,user_id) values (?,?)`, []any{uid, uid}},
		{`insert into d_taiwan.member_join_info (m_id) values (?)`, []any{uid}},
		{`insert into d_taiwan.member_miles (m_id) values (?)`, []any{uid}},
		{`insert into d_taiwan.member_white_account (m_id) values (?)`, []any{uid}},
		{`insert into taiwan_login.member_login (m_id) values (?)`, []any{uid}},
		{`insert into taiwan_billing.cash_cera (account,cera,mod_date,reg_date) values (?,100000,now(),now())`, []any{uid}},
		{`insert into taiwan_billing.cash_cera_point (account,cera_point,reg_date,mod_date) values (?,100000,now(),now())`, []any{uid}},
		{`insert into taiwan_cain_2nd.member_avatar_coin (m_id) values (?)`, []any{uid}},
	}

	for _, step := range steps {
		if _, err := db.ExecContext(ctx, step.query, step.args...); err != nil {
			a.logf("register step failed for %q: %v", user, err)
			return false, friendlyDBError(err)
		}
	}
	return true, nil
}

func friendlyDBError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return errors.New("连接服务器超时，请确认服务器已启动")
	}
	if errors.Is(err, syscall.ECONNREFUSED) {
		return errors.New("服务器拒绝连接，请检查 IP 与端口")
	}

	var me *mysql.MySQLError
	if errors.As(err, &me) {
		switch me.Number {
		case 1045:
			return errors.New("数据库账号或密码错误")
		case 1049:
			return errors.New("目标数据库不存在")
		case 1130:
			return errors.New("服务器不允许当前地址连接数据库")
		case 1146:
			return errors.New("数据库表不存在，请确认服务端版本")
		case 1054:
			return errors.New("数据库字段不匹配，请确认服务端版本")
		}
	}

	msg := err.Error()
	switch {
	case strings.Contains(msg, "connectex"), strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "no such host"), strings.Contains(msg, "i/o timeout"):
		return errors.New("无法连接服务器，请检查 IP、端口以及服务器状态")
	case strings.Contains(msg, "use of closed network connection"):
		return errors.New("连接已断开，请重试")
	}
	return errors.New("数据库操作失败：" + msg)
}
