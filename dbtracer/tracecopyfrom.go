package dbtracer

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type traceCopyFromData struct {
	ColumnNames []string       // 24 bytes
	span        trace.Span     // 16 bytes
	startTime   time.Time      // 16 bytes
	TableName   pgx.Identifier // slice - 24 bytes
}

func (dt *dbTracer) TraceCopyFromStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceCopyFromStartData) context.Context {
	ctx, span := dt.startSpan(ctx, "postgresql.copy_from", trace.WithAttributes(
		PGXOperationTypeKey.String("copy_from"),
		attribute.String("db.table", data.TableName.Sanitize()),
	))
	return context.WithValue(ctx, dbTracerCopyFromCtxKey, &traceCopyFromData{
		startTime:   time.Now(),
		TableName:   data.TableName,
		ColumnNames: data.ColumnNames,
		span:        span,
	})
}

func (dt *dbTracer) TraceCopyFromEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromEndData) {
	copyFromData := ctx.Value(dbTracerCopyFromCtxKey).(*traceCopyFromData)
	defer copyFromData.span.End()

	endTime := time.Now()
	interval := endTime.Sub(copyFromData.startTime)
	dt.recordHistogramMetric(ctx, "copy_from", "copy_from", interval, data.Err)

	var logAttrs []slog.Attr
	var level slog.Level

	if data.Err != nil {
		dt.recordSpanError(copyFromData.span, data.Err)
		logAttrs = append(logAttrs, slog.String("error", data.Err.Error()))
		level = slog.LevelError
	} else {
		copyFromData.span.SetStatus(codes.Ok, "")
		logAttrs = append(logAttrs, slog.Int64("rowCount", data.CommandTag.RowsAffected()))
		level = slog.LevelInfo
	}

	if dt.shouldLog(data.Err) {
		logAttrs = append(logAttrs, slog.Any("tableName", copyFromData.TableName),
			slog.Any("columnNames", copyFromData.ColumnNames),
			slog.Duration("time", interval),
			slog.Uint64("pid", uint64(extractConnectionID(conn))),
		)

		dt.logger.LogAttrs(ctx, level,
			"copyfrom",
			logAttrs...,
		)
	}
}
