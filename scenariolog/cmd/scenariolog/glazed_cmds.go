package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-go-golems/docmgr/scenariolog/internal/scenariolog"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func addGlazedCommands(rootCmd *cobra.Command) error {
	searchCmd, err := NewSearchGlazedCommand()
	if err != nil {
		return err
	}
	summaryCmd, err := NewSummaryGlazedCommand()
	if err != nil {
		return err
	}
	failuresCmd, err := NewFailuresGlazedCommand()
	if err != nil {
		return err
	}
	timingsCmd, err := NewTimingsGlazedCommand()
	if err != nil {
		return err
	}

	cobraSearch, err := cli.BuildCobraCommand(searchCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
	if err != nil {
		return err
	}
	cobraSummary, err := cli.BuildCobraCommand(summaryCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
	if err != nil {
		return err
	}
	cobraFailures, err := cli.BuildCobraCommand(failuresCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
	if err != nil {
		return err
	}
	cobraTimings, err := cli.BuildCobraCommand(timingsCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
	if err != nil {
		return err
	}

	rootCmd.AddCommand(cobraSearch, cobraSummary, cobraFailures, cobraTimings)
	return nil
}

type SearchGlazedCommand struct {
	*cmds.CommandDescription
}

type SearchGlazedSettings struct {
	DBPath string `glazed.parameter:"db"`
	RunID  string `glazed.parameter:"run-id"`
	Query  string `glazed.parameter:"query"`
	Limit  int    `glazed.parameter:"limit"`
}

func NewSearchGlazedCommand() (*SearchGlazedCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	commandSettingsLayer, err := cli.NewCommandSettingsLayer()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"search",
		cmds.WithShort("Search indexed log lines (FTS5)"),
		cmds.WithLong("Search the FTS index of captured stdout/stderr artifacts and output matches as structured rows."),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"db",
				parameters.ParameterTypeString,
				parameters.WithHelp("Path to sqlite database file"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"run-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Run id to search within"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"query",
				parameters.ParameterTypeString,
				parameters.WithHelp("FTS query string (e.g. 'warning OR error')"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"limit",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Max number of hits to return"),
				parameters.WithDefault(100),
			),
		),
		cmds.WithLayersList(glazedLayer, commandSettingsLayer),
	)

	return &SearchGlazedCommand{CommandDescription: cmdDesc}, nil
}

var _ cmds.GlazeCommand = &SearchGlazedCommand{}

func (c *SearchGlazedCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	s := &SearchGlazedSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	db, err := scenariolog.Open(ctx, s.DBPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if err := scenariolog.Migrate(ctx, db); err != nil {
		return err
	}

	ok, err := hasTable(ctx, db, "log_lines_fts")
	if err != nil {
		return err
	}
	if !ok {
		return scenariolog.ErrFTSNotAvailable
	}

	rows, err := db.QueryContext(ctx,
		`SELECT
		    log_lines_fts.artifact_id,
		    artifacts.step_id,
		    artifacts.kind,
		    artifacts.path,
		    log_lines_fts.line_num,
		    log_lines_fts.text
		  FROM log_lines_fts
		  JOIN artifacts ON artifacts.artifact_id = log_lines_fts.artifact_id
		  WHERE log_lines_fts.run_id = ?
		    AND log_lines_fts MATCH ?
		  LIMIT ?;`,
		s.RunID,
		s.Query,
		s.Limit,
	)
	if err != nil {
		return errors.Wrap(err, "query fts")
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var artifactID int64
		var stepID sql.NullString
		var kind string
		var path string
		var lineNum int
		var text string
		if err := rows.Scan(&artifactID, &stepID, &kind, &path, &lineNum, &text); err != nil {
			return errors.Wrap(err, "scan fts row")
		}

		stepIDStr := ""
		if stepID.Valid {
			stepIDStr = stepID.String
		}

		row := types.NewRow(
			types.MRP("artifact_id", artifactID),
			types.MRP("step_id", stepIDStr),
			types.MRP("kind", kind),
			types.MRP("path", path),
			types.MRP("line_num", lineNum),
			types.MRP("text", text),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "iterate fts rows")
	}

	return nil
}

type SummaryGlazedCommand struct {
	*cmds.CommandDescription
}

type SummaryGlazedSettings struct {
	DBPath string `glazed.parameter:"db"`
	RunID  string `glazed.parameter:"run-id"`
}

func NewSummaryGlazedCommand() (*SummaryGlazedCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	commandSettingsLayer, err := cli.NewCommandSettingsLayer()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"summary",
		cmds.WithShort("Show run summary"),
		cmds.WithLong("Show high-level run metadata plus counts of steps and failures."),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"db",
				parameters.ParameterTypeString,
				parameters.WithHelp("Path to sqlite database file"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"run-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Run id (defaults to latest run)"),
				parameters.WithDefault(""),
			),
		),
		cmds.WithLayersList(glazedLayer, commandSettingsLayer),
	)

	return &SummaryGlazedCommand{CommandDescription: cmdDesc}, nil
}

var _ cmds.GlazeCommand = &SummaryGlazedCommand{}

func (c *SummaryGlazedCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	s := &SummaryGlazedSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	db, err := scenariolog.Open(ctx, s.DBPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if err := scenariolog.Migrate(ctx, db); err != nil {
		return err
	}

	runID := s.RunID
	if runID == "" {
		runID, err = latestRunID(ctx, db)
		if err != nil {
			return err
		}
	}

	var rootDir string
	var suite sql.NullString
	var startedAt string
	var completedAt sql.NullString
	var exitCode sql.NullInt64
	var durationMs sql.NullInt64
	var totalSteps int64
	var failedSteps int64

	err = db.QueryRowContext(ctx, `
SELECT
  root_dir,
  suite,
  started_at,
  completed_at,
  exit_code,
  duration_ms,
  (SELECT COUNT(*) FROM steps WHERE steps.run_id = scenario_runs.run_id) AS total_steps,
  (SELECT COUNT(*) FROM steps WHERE steps.run_id = scenario_runs.run_id AND steps.exit_code != 0) AS failed_steps
FROM scenario_runs
WHERE run_id = ?;`, runID).Scan(
		&rootDir, &suite, &startedAt, &completedAt, &exitCode, &durationMs, &totalSteps, &failedSteps,
	)
	if err != nil {
		return errors.Wrap(err, "select scenario_runs summary")
	}

	ftsEnabled, err := hasTable(ctx, db, "log_lines_fts")
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("run_id", runID),
		types.MRP("root_dir", rootDir),
		types.MRP("suite", nullStringToString(suite)),
		types.MRP("started_at", startedAt),
		types.MRP("completed_at", nullStringToString(completedAt)),
		types.MRP("exit_code", nullIntToInt(exitCode)),
		types.MRP("duration_ms", nullIntToInt64(durationMs)),
		types.MRP("total_steps", totalSteps),
		types.MRP("failed_steps", failedSteps),
		types.MRP("fts_enabled", ftsEnabled),
	)
	return gp.AddRow(ctx, row)
}

type FailuresGlazedCommand struct {
	*cmds.CommandDescription
}

type FailuresGlazedSettings struct {
	DBPath string `glazed.parameter:"db"`
	RunID  string `glazed.parameter:"run-id"`
}

func NewFailuresGlazedCommand() (*FailuresGlazedCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	commandSettingsLayer, err := cli.NewCommandSettingsLayer()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"failures",
		cmds.WithShort("List failing steps"),
		cmds.WithLong("List steps with non-zero exit code, including stderr artifact paths."),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"db",
				parameters.ParameterTypeString,
				parameters.WithHelp("Path to sqlite database file"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"run-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Run id (defaults to latest run)"),
				parameters.WithDefault(""),
			),
		),
		cmds.WithLayersList(glazedLayer, commandSettingsLayer),
	)
	return &FailuresGlazedCommand{CommandDescription: cmdDesc}, nil
}

var _ cmds.GlazeCommand = &FailuresGlazedCommand{}

func (c *FailuresGlazedCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	s := &FailuresGlazedSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	db, err := scenariolog.Open(ctx, s.DBPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if err := scenariolog.Migrate(ctx, db); err != nil {
		return err
	}

	runID := s.RunID
	if runID == "" {
		runID, err = latestRunID(ctx, db)
		if err != nil {
			return err
		}
	}

	rows, err := db.QueryContext(ctx, `
SELECT
  s.step_num,
  s.step_name,
  s.exit_code,
  s.duration_ms,
  COALESCE(a.path, '') AS stderr_path
FROM steps s
LEFT JOIN artifacts a ON a.step_id = s.step_id AND a.kind = 'stderr'
WHERE s.run_id = ?
  AND s.exit_code IS NOT NULL
  AND s.exit_code != 0
ORDER BY s.step_num;`, runID)
	if err != nil {
		return errors.Wrap(err, "query failures")
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var stepNum int
		var stepName string
		var exitCode int
		var durationMs sql.NullInt64
		var stderrPath string
		if err := rows.Scan(&stepNum, &stepName, &exitCode, &durationMs, &stderrPath); err != nil {
			return errors.Wrap(err, "scan failures row")
		}

		row := types.NewRow(
			types.MRP("run_id", runID),
			types.MRP("step_num", stepNum),
			types.MRP("step_name", stepName),
			types.MRP("exit_code", exitCode),
			types.MRP("duration_ms", nullIntToInt64(durationMs)),
			types.MRP("stderr_path", stderrPath),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "iterate failures rows")
	}

	return nil
}

type TimingsGlazedCommand struct {
	*cmds.CommandDescription
}

type TimingsGlazedSettings struct {
	DBPath string `glazed.parameter:"db"`
	RunID  string `glazed.parameter:"run-id"`
	Top    int    `glazed.parameter:"top"`
}

func NewTimingsGlazedCommand() (*TimingsGlazedCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	commandSettingsLayer, err := cli.NewCommandSettingsLayer()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"timings",
		cmds.WithShort("Show slowest steps"),
		cmds.WithLong("Show slowest steps in the run (sorted by duration_ms DESC)."),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"db",
				parameters.ParameterTypeString,
				parameters.WithHelp("Path to sqlite database file"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"run-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Run id (defaults to latest run)"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"top",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Return top N steps"),
				parameters.WithDefault(10),
			),
		),
		cmds.WithLayersList(glazedLayer, commandSettingsLayer),
	)
	return &TimingsGlazedCommand{CommandDescription: cmdDesc}, nil
}

var _ cmds.GlazeCommand = &TimingsGlazedCommand{}

func (c *TimingsGlazedCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	s := &TimingsGlazedSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	db, err := scenariolog.Open(ctx, s.DBPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if err := scenariolog.Migrate(ctx, db); err != nil {
		return err
	}

	runID := s.RunID
	if runID == "" {
		runID, err = latestRunID(ctx, db)
		if err != nil {
			return err
		}
	}

	rows, err := db.QueryContext(ctx, `
SELECT step_num, step_name, duration_ms, exit_code
FROM steps
WHERE run_id = ?
ORDER BY duration_ms DESC
LIMIT ?;`, runID, s.Top)
	if err != nil {
		return errors.Wrap(err, "query timings")
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var stepNum int
		var stepName string
		var durationMs sql.NullInt64
		var exitCode sql.NullInt64
		if err := rows.Scan(&stepNum, &stepName, &durationMs, &exitCode); err != nil {
			return errors.Wrap(err, "scan timings row")
		}

		row := types.NewRow(
			types.MRP("run_id", runID),
			types.MRP("step_num", stepNum),
			types.MRP("step_name", stepName),
			types.MRP("duration_ms", nullIntToInt64(durationMs)),
			types.MRP("exit_code", nullIntToInt(exitCode)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "iterate timings rows")
	}

	return nil
}

func latestRunID(ctx context.Context, db *sql.DB) (string, error) {
	var runID string
	err := db.QueryRowContext(ctx, "SELECT run_id FROM scenario_runs ORDER BY started_at DESC LIMIT 1;").Scan(&runID)
	if err != nil {
		return "", errors.Wrap(err, "select latest run_id")
	}
	if runID == "" {
		return "", errors.New("no runs found")
	}
	return runID, nil
}

func hasTable(ctx context.Context, db *sql.DB, name string) (bool, error) {
	var c int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type IN ('table','view') AND name = ?;",
		name,
	).Scan(&c)
	if err != nil {
		return false, errors.Wrap(err, "sqlite_master lookup")
	}
	return c > 0, nil
}

func nullStringToString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func nullIntToInt64(i sql.NullInt64) int64 {
	if i.Valid {
		return i.Int64
	}
	return 0
}

func nullIntToInt(i sql.NullInt64) int {
	if i.Valid {
		return int(i.Int64)
	}
	return 0
}


