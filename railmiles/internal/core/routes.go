package core

import (
	"context"
	"github.com/codemicro/railmiles/railmiles/internal/db"
	"github.com/google/uuid"
)

func (c *Core) GetCallingPoints(journeyID uuid.UUID) ([]string, error) {
	var route []string
	err := c.db.DB.NewSelect().Model((*db.Route)(nil)).Column("station").Where(`journey_id = ?`, journeyID).Order("sequence").Scan(context.Background(), &route)
	return route, err
}

func (c *Core) InsertRoute(journeyID uuid.UUID, route []string) error {
	var routeParts []*db.Route
	r := &db.Route{
		JourneyID: journeyID,
	}
	for i, point := range route {
		rq := *r
		rq.Sequence = i
		rq.Station = point
		routeParts = append(routeParts, &rq)
	}
	_, err := c.db.DB.NewInsert().Model(&routeParts).Exec(context.Background())
	return err
}
