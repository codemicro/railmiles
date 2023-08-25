package core

import (
	"context"
	"database/sql"
	"errors"
	"github.com/codemicro/railmiles/railmiles/internal/db"
)

func (c *Core) GetCallingPoints(from, to string) ([]string, error) {
	route := new(db.Route)
	err := c.db.DB.NewSelect().Model(route).Where(`"from" = ? and "to" = ?`, from, to).Scan(context.Background(), route)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err := c.db.DB.NewSelect().Model(route).Where(`"from" = ? and "to" = ?`, to, from).Scan(context.Background(), route)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return []string{}, nil
				}
				return nil, err
			}
			// reverse route
			for i, j := 0, len(route.Route)-1; i < j; i, j = i+1, j-1 {
				route.Route[i], route.Route[j] = route.Route[j], route.Route[i]
			}
		}
	}

	return route.Route, err
}

func (c *Core) InsertRoute(route *db.Route) error {
	_, err := c.db.DB.NewInsert().Ignore().Model(route).Exec(context.Background())
	return err
}
