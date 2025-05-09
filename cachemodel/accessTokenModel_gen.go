// Code generated by goctl. DO NOT EDIT!

package cachemodel

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	accessTokenFieldNames          = builder.RawFieldNames(&AccessToken{})
	accessTokenRows                = strings.Join(accessTokenFieldNames, ",")
	accessTokenRowsExpectAutoSet   = strings.Join(stringx.Remove(accessTokenFieldNames, "`id`", "`create_time`", "`update_time`", "`create_at`", "`update_at`"), ",")
	accessTokenRowsWithPlaceHolder = strings.Join(stringx.Remove(accessTokenFieldNames, "`id`", "`create_time`", "`update_time`", "`create_at`", "`update_at`"), "=?,") + "=?"
)

type (
	accessTokenModel interface {
		Insert(ctx context.Context, data *AccessToken) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*AccessToken, error)
		Update(ctx context.Context, data *AccessToken) error
		Delete(ctx context.Context, id int64) error
	}

	defaultAccessTokenModel struct {
		conn  sqlx.SqlConn
		table string
	}

	AccessToken struct {
		Id       int64     `db:"id"`       // id
		Token    string    `db:"token"`    // token字符串
		Time     time.Time `db:"time"`     // 更新时间
		Overtime int64     `db:"overtime"` // token生命长度
	}
)

func newAccessTokenModel(conn sqlx.SqlConn) *defaultAccessTokenModel {
	return &defaultAccessTokenModel{
		conn:  conn,
		table: "`access_token`",
	}
}

func (m *defaultAccessTokenModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

func (m *defaultAccessTokenModel) FindOne(ctx context.Context, id int64) (*AccessToken, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", accessTokenRows, m.table)
	var resp AccessToken
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultAccessTokenModel) Insert(ctx context.Context, data *AccessToken) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?)", m.table, accessTokenRowsExpectAutoSet)
	ret, err := m.conn.ExecCtx(ctx, query, data.Token, data.Time, data.Overtime)
	return ret, err
}

func (m *defaultAccessTokenModel) Update(ctx context.Context, data *AccessToken) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, accessTokenRowsWithPlaceHolder)
	_, err := m.conn.ExecCtx(ctx, query, data.Token, data.Time, data.Overtime, data.Id)
	return err
}

func (m *defaultAccessTokenModel) tableName() string {
	return m.table
}
