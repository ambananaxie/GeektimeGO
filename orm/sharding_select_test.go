//go:build sharding
package orm

import (
	"fmt"
	"gitee.com/geektime-geekbang/geektime-go/orm/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestShardingSelector_findDstByPredicate(t *testing.T) {
	r := model.NewRegistry()
	m, err := r.Get(&Order{})
	m.Sf = func(skVal any) (string, string) {
		db := skVal.(int64) / 100
		tbl := skVal.(int64) % 10
		return fmt.Sprintf("order_db_%d", db), fmt.Sprintf("order_tab_%d", tbl)
	}
	m.Sk = "UserId"
	require.NoError(t, err)
	s := ShardingSelector[Order]{
		builder: builder {
			core: core{
				model: m,
			},
		},
	}
	testCases := []struct{
		name string
		p Predicate
		wantDsts []Dst
		wantErr error
	} {
		{
			name: "only eq",
			p: C("UserId").EQ(int64(123)),
			wantDsts: []Dst{
				{
					DB: "order_db_1",
					Table: "order_tab_3",
				},
			},
		},
		{
			name: "and left broadcast",
			p: C("Id").EQ(12).And(C("UserId").EQ(int64(123))),
			wantDsts: []Dst{
				{
					DB: "order_db_1",
					Table: "order_tab_3",
				},
			},
		},
		{
			name: "and right broadcast",
			p: C("UserId").EQ(int64(123)).And(C("Id").EQ(12)),
			wantDsts: []Dst{
				{
					DB: "order_db_1",
					Table: "order_tab_3",
				},
			},
		},
		{
			name: "and empty",
			p: C("UserId").EQ(int64(123)).And(C("UserId").EQ(int64(124))),
			wantDsts: []Dst{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dsts, err := s.findDstByPredicate(tc.p)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantDsts, dsts)
		})
	}
}

type Order struct {
	UserId int64
}

