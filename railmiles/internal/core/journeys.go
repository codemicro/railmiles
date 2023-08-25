package core

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/codemicro/railmiles/railmiles/internal/db"
	"github.com/codemicro/railmiles/railmiles/internal/util"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type timeSince uint8

const (
	AllTime timeSince = iota
	LastMonth
	YearToDate
)

func (ts timeSince) SQLDuration() (string, error) {
	var dur string
	switch ts {
	case AllTime:
	case LastMonth:
		dur = "-1 month"
	case YearToDate:
		dur = "start of year"
	default:
		return "", fmt.Errorf("unknown timeSince %d", ts)
	}
	return dur, nil
}

type GetJourneysArgs struct {
	Since  timeSince
	Offset int
	Limit  int
}

func (c *Core) GetJourneys(args *GetJourneysArgs) ([]*db.Journey, error) {
	var journeys []*db.Journey

	q := c.db.DB.NewSelect().
		Model(&journeys).
		OrderExpr(`"journey"."date" DESC`)

	if args.Offset != 0 {
		q = q.Offset(args.Offset)
	}

	if args.Limit != 0 {
		q = q.Limit(args.Limit)
	}

	dur, err := args.Since.SQLDuration()
	if err != nil {
		return nil, util.Wrap(err, "getting journeys")
	}

	if dur != "" {
		q = q.Where(`"journey"."date" > date('now', ?)`, dur)
	}

	if err := q.Scan(context.Background()); err != nil {
		return nil, fmt.Errorf("querying past journeys: %w", err)
	}
	return journeys, nil
}

type JourneyStats struct {
	Count int `json:"count"`
	// RawCount is the count as in the number of rows, not the count as in the
	// number of trips (ie. counts return journeys as one)
	RawCount int     `json:"rawCount"`
	Miles    float32 `json:"miles"`
}

func (c *Core) GetJourneyStats(since timeSince) (*JourneyStats, error) {
	var (
		rtns      []bool
		distQtys  []float32
		countQtys []int
	)

	q := c.db.DB.NewSelect().
		Model((*db.Journey)(nil)).
		ColumnExpr("return, sum(distance), count(*)").
		Group("return")

	dur, err := since.SQLDuration()
	if err != nil {
		return nil, util.Wrap(err, "getting journey stats")
	}

	if dur != "" {
		q = q.Where(`"journey"."date" > date('now', ?)`, dur)
	}

	if err := q.Scan(context.Background(), &rtns, &distQtys, &countQtys); err != nil {
		return nil, fmt.Errorf("querying total miles: %w", err)
	}

	js := new(JourneyStats)
	for i := 0; i < len(rtns); i += 1 {
		if rtns[i] { // if this is a return journey:
			js.Count += countQtys[i] * 2
			js.Miles += distQtys[i] * 2
		} else {
			js.Count += countQtys[i]
			js.Miles += distQtys[i]
		}
		js.RawCount += countQtys[i]
	}

	return js, nil
}

func PopulateFullStationNames(journeys []*db.Journey) {
	for _, journey := range journeys {
		journey.From.Full = GetStationName(journey.From.Shortcode)
		journey.To.Full = GetStationName(journey.To.Shortcode)
		for _, via := range journey.Via {
			via.Full = GetStationName(via.Shortcode)
		}
	}
}

func (c *Core) GetJourney(id uuid.UUID) (*db.Journey, error) {
	j := new(db.Journey)
	err := c.db.DB.NewSelect().Model(j).Where("id = ?", id).Scan(context.Background(), j)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return j, nil
}

func (c *Core) DeleteJourney(id uuid.UUID) error {
	_, err := c.db.DB.NewDelete().Model((*db.Journey)(nil)).Where("id = ?", id).Exec(context.Background())
	return err
}

func (c *Core) InsertJourney(journey *db.Journey) error {
	_, err := c.db.DB.NewInsert().Model(journey).Exec(context.Background())
	return err
}
