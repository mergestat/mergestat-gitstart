package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mergestat/mergestat/internal/db"
	"github.com/mergestat/sqlq"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// AutoImport implements the githubRepository auto-import job to automatically
// sync githubRepository from user- or org- accounts.
func AutoImport(pool *pgxpool.Pool) sqlq.HandlerFunc {
	var queries = db.New(pool)

	return func(ctx context.Context, job *sqlq.Job) (err error) {
		var l = zerolog.Ctx(ctx)
		var logger = job.Logger()
		var jobErrors error

		// start sending periodic keep-alive pings!
		go job.SendKeepAlive(ctx, job.KeepAlive-(5*time.Second)) //nolint:errcheck

		// fetch a list of all configured imports that are due now
		var imports []db.ListRepoImportsDueForImportRow
		if imports, err = queries.ListRepoImportsDueForImport(ctx); err != nil {
			logger.Errorf("failed to list repo import job: %v", err)
			l.Error().Msgf("failed to list repo import job: %v", err)
			return errors.Wrapf(sqlq.ErrSkipRetry, "failed to list repo import job: %v", err)
		}

		logger.Infof("handling %d import(s)", len(imports))
		l.Info().Msgf("handling %d import(s)", len(imports))
		for _, imp := range imports {
			logger.Infof("executing import %s", imp.ID)
			l.Info().Msgf("executing import %s", imp.ID)
			var tx pgx.Tx // each import is executed within its own transaction
			if tx, err = pool.Begin(ctx); err != nil {
				return errors.Wrapf(err, "failed to start new database transaction")
			}
			var importStatus = db.UpdateImportStatusParams{Status: "RUNNING", ID: imp.ID}
			if err = queries.UpdateImportStatus(ctx, importStatus); err != nil {
				return errors.Wrapf(sqlq.ErrSkipRetry, "failed to update import status: %v", err)
			}
			// execute the import
			// if the execution fails for some reason, only that import is marked as failed
			// the job still continues executing.
			var importError error
			if imp.VendorName == "github" {
				importError = handleGithubImport(ctx, queries.WithTx(tx), imp)
			} else if imp.VendorName == "bitbucket" {
				importError = handleBitbucketImport(ctx, queries.WithTx(tx), imp)
			} else if imp.VendorName == "gitlab" {
				importError = handleGitlabImport(ctx, queries.WithTx(tx), imp)
			} else {
				importError = errors.Errorf("unknown vendor: %s", imp.VendorName)
			}

			if importError != nil {
				logger.Warnf("import(%s) failed: %v", imp.ID, importError.Error())
				l.Warn().Msgf("import(%s) failed: %v", imp.ID, importError.Error())

				if err = tx.Rollback(ctx); err != nil {
					jobErrors = errors.Wrap(err, "failed to rollback transaction")
					return jobErrors
				}

				jobErrors = errors.Wrap(importError, "failed to handle import")

			} else {
				// if the import was successful, commit the changes
				if err = tx.Commit(ctx); err != nil {
					jobErrors = errors.Wrapf(err, "failed to commit database transaction")
					return jobErrors
				}
				logger.Infof("import(%s) was successful", imp.ID)
				l.Info().Msgf("import(%s) was successful", imp.ID)
			}

			importStatus = db.UpdateImportStatusParams{Status: "SUCCESS", ID: imp.ID}
			if importError != nil {
				importStatus.Status, importStatus.Error = "FAILURE", importError.Error()
			}

			// import status updates happen outside the transaction in which the import was processed.
			// this is so that even if the transaction is marked as errored, we could still go ahead and update
			// the job status.
			if err = queries.UpdateImportStatus(ctx, importStatus); err != nil {
				jobErrors = errors.Wrapf(sqlq.ErrSkipRetry, "failed to update import status: %v", err)
				return jobErrors
			}
		}
		return jobErrors
	}
}

// adapted from https://stackoverflow.com/a/45428032/6611700
func difference(existing, new []string) []string {
	if len(existing) == 0 {
		return new // early exit path
	}

	var ex = make(map[string]struct{}, len(existing)) // set of all existing items
	for _, x := range existing {
		ex[x] = struct{}{}
	}

	var diff []string
	for _, x := range new {
		if _, found := ex[x]; !found {
			diff = append(diff, x)
		}
	}

	return diff
}
