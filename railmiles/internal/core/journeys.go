package core

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/codemicro/railmiles/railmiles/internal/db"
	"github.com/codemicro/railmiles/railmiles/internal/util"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
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
	Count int     `json:"count"`
	Miles float32 `json:"miles"`
}

func (c *Core) GetJourneyStats(since timeSince) (*JourneyStats, error) {
	q := c.db.DB.NewSelect().
		Model((*db.Journey)(nil)).
		ColumnExpr("sum(distance), count(*)")

	dur, err := since.SQLDuration()
	if err != nil {
		return nil, util.Wrap(err, "getting journey stats")
	}

	if dur != "" {
		q = q.Where(`"journey"."date" > date('now', ?)`, dur)
	}

	js := new(JourneyStats)
	if err := q.Scan(context.Background(), &js.Miles, &js.Count); err != nil {
		return nil, fmt.Errorf("querying total miles: %w", err)
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
	if err != nil {
		return err
	}
	_, err = c.db.DB.NewUpdate().Model((*db.Journey)(nil)).Set("return_id = null").Where("return_id = ?", id).Exec(context.Background())
	if err != nil {
		return err
	}
	_, err = c.db.DB.NewDelete().Model((*db.Route)(nil)).Where("journey_id = ?", id).Exec(context.Background())
	return err
}

func (c *Core) InsertJourney(journey *db.Journey) error {
	_, err := c.db.DB.NewInsert().Model(journey).Exec(context.Background())
	return err
}

func (c *Core) UpdateJourney(journey *db.Journey) error {
	_, err := c.db.DB.NewUpdate().Model(journey).WherePK().Exec(context.Background())
	return err
}

var ErrReturnAlreadyExists = errors.New("return journey already exists")

func (c *Core) CreateReturnJourney(id uuid.UUID) (uuid.UUID, error) {
	sourceJourney := new(db.Journey)
	if err := c.db.DB.NewSelect().Model(sourceJourney).Where("id = ?", id).Scan(context.Background()); err != nil {
		return uuid.UUID{}, err
	}
	if sourceJourney.ReturnID != nil {
		return uuid.UUID{}, ErrReturnAlreadyExists
	}
	calls, err := c.GetCallingPoints(id)
	if err != nil {
		return uuid.UUID{}, err
	}

	var newJourney *db.Journey
	{
		x := *sourceJourney
		newJourney = &x
	}

	newJourney.To, newJourney.From = newJourney.From, newJourney.To
	if len(newJourney.Via) != 0 {
		// When creating a new journey and adding a return at this same time, newJourney.Via being a copy of sourceJourney.Via and being reversed in place would cause the original to be reversed in place too.
		n := make([]*db.StationName, len(newJourney.Via))
		copy(n, newJourney.Via)
		newJourney.Via = n
	}
	slices.Reverse(newJourney.Via)
	newJourney.ID = uuid.New()
	newJourney.ReturnID = &sourceJourney.ID
	if err := c.InsertJourney(newJourney); err != nil {
		return uuid.UUID{}, err
	}
	if len(calls) != 0 {
		slices.Reverse(calls)
		if err := c.InsertRoute(newJourney.ID, calls); err != nil {
			return uuid.UUID{}, err
		}
	}

	sourceJourney.ReturnID = &newJourney.ID
	if err := c.UpdateJourney(sourceJourney); err != nil {
		return uuid.UUID{}, err
	}

	return newJourney.ID, nil
}
