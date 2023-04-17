//go:build sharding
package orm

import (
	"context"
	"database/sql"
)

type ShardingDB struct {
	// key 就是 Dst 里面的 DB
	DBs map[string]*MasterSlavesDB
}

type MasterSlavesDB struct {
	Master *sql.DB
	//Table []string
	Slaves []*sql.DB
}

func (m *MasterSlavesDB) query(ctx context.Context, sql string, args...any) (*sql.Rows, error) {
	// 这边要做两件事情
	// 1. 决定走 master 还是走 slave
	// 2. 如果走 slave，怎么负载均衡
	db := m.Slaves[0]
	return db.QueryContext(ctx, sql, args...)
}
