package cron

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/mergestat/mergestat/internal/jobs/sync/podman"
	"github.com/mergestat/sqlq"
	"github.com/rs/zerolog"
	"time"
)

// ContainerSync provides a cron function that periodically schedules execution
// of configured mergestat.container_sync_schedules.
func ContainerSync(ctx context.Context, dur time.Duration, upstream *sql.DB) {
	var log = zerolog.Ctx(ctx)

	const listSyncsQuery = `
		SELECT DISTINCT ON (syncs.id) syncs.id
		FROM mergestat.container_sync_schedules schd, mergestat.container_syncs syncs
			LEFT OUTER JOIN mergestat.container_sync_executions exec ON exec.sync_id = syncs.id
			LEFT OUTER JOIN sqlq.jobs job ON job.id = exec.job_id
		WHERE syncs.id = schd.sync_id AND 
			(job.status IS NULL OR job.status NOT IN ('pending', 'running'))
		ORDER BY syncs.id, exec.created_at`

	const createExecutionQuery = "INSERT INTO mergestat.container_sync_executions (sync_id, job_id) VALUES ($1, $2)"

	var fn = func() error {
		var err error
		var tx *sql.Tx
		if tx, err = upstream.BeginTx(ctx, &sql.TxOptions{}); err != nil {
			return err
		}
		defer tx.Rollback() //nolint:errcheck

		var syncs []uuid.UUID
		var rows *sql.Rows
		if rows, err = tx.QueryContext(ctx, listSyncsQuery); err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id uuid.UUID
			if err = rows.Scan(&id); err != nil {
				return err
			}
			syncs = append(syncs, id)
		}

		var createExecution *sql.Stmt
		if createExecution, err = tx.PrepareContext(ctx, createExecutionQuery); err != nil {
			return err
		}
		defer createExecution.Close()

		for _, sync := range syncs {
			var job *sqlq.Job
			if job, err = sqlq.Enqueue(tx, "default", podman.NewContainerSync(sync)); err != nil {
				return err
			}

			if _, err = createExecution.ExecContext(ctx, sync, job.ID); err != nil {
				return err
			}
		}

		return tx.Commit()
	}

	// reuse existing loop-select functionality in Basic()
	Basic(ctx, dur, func() {
		if err := fn(); err != nil {
			log.Err(err).Msg("failed to start container sync")
		}
	})
}
