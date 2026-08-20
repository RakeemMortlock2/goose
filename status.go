package goose

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"
	"text/tabwriter"
	"time"
)

// Status prints the status of all migrations.
func Status(db *sql.DB, dir string, opts ...Options) error {
	return StatusContext(context.Background(), db, dir, opts...)
}

// StatusContext prints the status of all migrations.
func StatusContext(ctx context.Context, db *sql.DB, dir string, opts ...Options) error {
	option := &options{}
	for _, opt := range opts {
		opt(option)
	}

	migrations, err := CollectMigrations(dir, minVersion, maxVersion)
	if err != nil {
		return err
	}

	dbMigrations, err := GetMigrationHistory(db)
	if err != nil {
		return err
	}

	localMap := make(map[int64]*Migration)
	for _, m := range migrations {
		localMap[m.Version] = m
	}

	dbMap := make(map[int64]*MigrationRecord)
	for _, dbm := range dbMigrations {
		dbMap[dbm.VersionID] = dbm
	}

	versionSet := make(map[int64]bool)
	for _, m := range migrations {
		versionSet[m.Version] = true
	}
	for _, dbm := range dbMigrations {
		if dbm.IsApplied {
			versionSet[dbm.VersionID] = true
		}
	}

	var versions []int64
	for v := range versionSet {
		versions = append(versions, v)
	}
	sort.Slice(versions, func(i, j int) bool {
		return versions[i] < versions[j]
	})

	w := tabwriter.NewWriter(stdout, 16, 4, 0, ' ', 0)
	fmt.Fprintln(w, "    Applied At\tMigration")
	fmt.Fprintln(w, "    =======================================")

	for _, v := range versions {
		if m, ok := localMap[v]; ok {
			var appliedAt string
			if dbm, ok := dbMap[v]; ok && dbm.IsApplied {
				appliedAt = dbm.TStamp.Format(time.ANSIC)
			} else {
				appliedAt = "Pending"
			}
			fmt.Fprintf(w, "    %s\t- %s\n", appliedAt, filepath.Base(m.Source))
		} else {
			var appliedAt string
			if dbm, ok := dbMap[v]; ok && dbm.IsApplied {
				appliedAt = dbm.TStamp.Format(time.ANSIC)
			} else {
				appliedAt = "Pending"
			}
			fmt.Fprintf(w, "    %s\t- %05d (missing local file)\n", appliedAt, v)
		}
	}

	return w.Flush()
}

func printStatus(w *tabwriter.Writer, db *sql.DB, version int64, filepath string) { 
	var appliedAt string

	if version == 0 {
		appliedAt = time.Now().Format(time.ANSIC)
	} else {
		row := db.QueryRow(GetDialect().dbVersionQuery(tableName), version)
		var t time.Time
		if err := row.Scan(&t); err == nil {
			appliedAt = t.Format(time.ANSIC)
		} else {
			appliedAt = "Pending"
		}
	}

	fmt.Fprintf(w, "    %s\t- %s\n", appliedAt, filepath)
}
