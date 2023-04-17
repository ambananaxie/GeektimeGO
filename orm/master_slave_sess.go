package orm

import (
	"context"
	"database/sql"
	"math/rand"
)

type MasterSlaveDB struct {
	master *DB
	slaves []*DB
}

func (m *MasterSlaveDB) getCore() core {
	return m.master.getCore()
}

func (m *MasterSlaveDB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {

	// 要判断是不是在事务里面
	// 有没有加锁，for update

	// 这里你就要考虑负载均衡的问题了
	idx := rand.Intn(len(m.slaves))
	return m.slaves[idx].queryContext(ctx, query, args...)
}

func (m *MasterSlaveDB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return m.master.execContext(ctx, query, args...)
}


type Cluster struct {
	// order_db_1: xxxx
	// order_db_2: xxxx
	DBs map[string]*MasterSlaveDB
}


func (m *Cluster) getCore() core {
	panic("implement me")
}

func (m *Cluster) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	//dbName := ctx.Value("db")
	// 我怎么知道你要查询哪个 db
	panic("implement me")
}

func (m *Cluster) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	// 我怎么知道你要查询哪个 db
	panic("implement me")
}