package database

import (
	"context"
	"errors"
	"time"

	"github.com/mbrockmandev/tometracker/internal/models"
)

// ActivityLog Methods
func (p *PostgresDBRepo) CreateActivityLog(activity *models.ActivityLog) (int, error) {
	return 0, nil
}

func (p *PostgresDBRepo) GetAllActivityLogs() ([]*models.ActivityLog, error) {
	return nil, nil
}

func (p *PostgresDBRepo) GetActivityLogById(id int) (*models.ActivityLog, error) {
	return nil, nil
}

func (p *PostgresDBRepo) DeleteActivityLog(id int) error {
	return nil
}

func (p *PostgresDBRepo) UpdateActivityLog(id int, activity *models.ActivityLog) error {
	return nil
}

type BusyTime struct {
	Hour  time.Time
	Count int
}

func (p *PostgresDBRepo) ReportBusyTimes() ([]*BusyTime, error) {
	query := `
		select
			date_trunc('hour', al.timestamp) as hour,
			count(*) as activity_count
		from
			activity_logs al
		where
			al.timestamp >= $1
		group by
			date_trunc('hour', al.timestamp)
		order by
			activity_count desc,
			date_trunc('hour', al.timestamp) asc
		limit
			10;
	`

	oneMonthAgo := time.Now().AddDate(0, -1, 0)

	rows, err := p.DB.Query(query, oneMonthAgo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var busyHours []*BusyTime

	for rows.Next() {
		var hour time.Time
		var count int
		if err := rows.Scan(&hour, &count); err != nil {
			return nil, err
		}

		busyHours = append(busyHours, &BusyTime{
			Hour:  hour,
			Count: count,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return busyHours, nil
}

func (p *PostgresDBRepo) LogActivity(activity *models.ActivityLog) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var exists bool
	stmt := `
		select
			exists
				(select 1 from users where id = $1)
	`
	err := p.DB.QueryRowContext(ctx, stmt, activity.UserID).Scan(&exists)

	if err != nil {
		return err
	}
	if !exists {
		return errors.New("user does not exist")
	}

	stmt = `
		insert into
			activity_logs
				(user_id, action, timestamp)
			values
				($1, $2, $3)
	`

	_, err = p.DB.ExecContext(ctx, stmt, activity.UserID, activity.Action, activity.Timestamp)
	return err
}
