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
	errLogFieldNames          = builder.RawFieldNames(&ErrLog{})
	errLogRows                = strings.Join(errLogFieldNames, ",")
	errLogRowsExpectAutoSet   = strings.Join(stringx.Remove(errLogFieldNames, "`id`", "`create_time`", "`update_time`", "`create_at`", "`update_at`"), ",")
	errLogRowsWithPlaceHolder = strings.Join(stringx.Remove(errLogFieldNames, "`id`", "`create_time`", "`update_time`", "`create_at`", "`update_at`"), "=?,") + "=?"
)

type (
	errLogModel interface {
		Insert(ctx context.Context, data *ErrLog) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*ErrLog, error)
		Update(ctx context.Context, data *ErrLog) error
		Delete(ctx context.Context, id int64) error
	}

	defaultErrLogModel struct {
		conn  sqlx.SqlConn
		table string
	}

	ErrLog struct {
		Id        int64     `db:"id"`        // id
		Info      string    `db:"info"`      // 报错信息
		Interface string    `db:"interface"` // 报错接口
		Time      time.Time `db:"time"`      // 报错时间
	}
)

func newErrLogModel(conn sqlx.SqlConn) *defaultErrLogModel {
	return &defaultErrLogModel{
		conn:  conn,
		table: "`err_log`",
	}
}

func (m *defaultErrLogModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

func (m *defaultErrLogModel) FindOne(ctx context.Context, id int64) (*ErrLog, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", errLogRows, m.table)
	var resp ErrLog
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

func (m *defaultErrLogModel) Insert(ctx context.Context, data *ErrLog) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?)", m.table, errLogRowsExpectAutoSet)
	ret, err := m.conn.ExecCtx(ctx, query, data.Info, data.Interface, data.Time)
	return ret, err
}

func (m *defaultErrLogModel) Update(ctx context.Context, data *ErrLog) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, errLogRowsWithPlaceHolder)
	_, err := m.conn.ExecCtx(ctx, query, data.Info, data.Interface, data.Time, data.Id)
	return err
}

func (m *defaultErrLogModel) tableName() string {
	return m.table
}
