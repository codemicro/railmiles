package core

import (
	"context"
	"errors"
	"github.com/codemicro/railmiles/railmiles/internal/db"
	"github.com/mattn/go-sqlite3"
)

func (c *Core) GetCallingPoints(from, to string) ([]string, error) {
	var route []*db.Route
	err := c.db.DB.NewSelect().Model(&route).Where(`"from" = ? and "to" = ?`, from, to).Scan(context.Background(), &route)
	if err != nil {
		return []string{}, nil
	}

	if len(route) == 0 {
		err := c.db.DB.NewSelect().Model(&route).Where(`"from" = ? and "to" = ?`, to, from).Scan(context.Background(), &route)
		if err != nil {
			if len(route) == 0 {
				return []string{}, nil
			}
			return nil, err
		}
		// reverse route
		for i, j := 0, len(route)-1; i < j; i, j = i+1, j-1 {
			route[i], route[j] = route[j], route[i]
		}
	}

	var routeStrings []string
	for _, r := range route {
		routeStrings = append(routeStrings, r.Station)
	}

	return routeStrings, err
}

func (c *Core) InsertRoute(from, to string, route []string) error {
	var routeParts []*db.Route
	r := &db.Route{
		From: from,
		To:   to,
	}
	for i, point := range route {
		rq := *r
		rq.Sequence = i
		rq.Station = point
		routeParts = append(routeParts, &rq)
	}
	_, err := c.db.DB.NewInsert().Model(routeParts).Exec(context.Background())
	if err != nil {
		var e sqlite3.Error
		if errors.As(err, &e) {
			if e.Code == sqlite3.ErrConstraint {
				return nil
			}
		}
	}
	return err
}
