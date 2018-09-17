package disport

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/seibert-media/golibs/log"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

var (
	// Subject ID of the recorded metric
	Subject, _ = tag.NewKey("subject")
)

var (
	// ReportCount .
	ReportCount = stats.Int64("report/count", "Count of reports", "1")
)

// Report a user to the admin
func Report(ctx context.Context, m *discordgo.MessageCreate, u *discordgo.User) error {
	ctx, err := tag.New(ctx,
		tag.Insert(Subject, u.ID),
	)
	if err != nil {
		log.From(ctx).Error("adding tags", zap.Error(err))
	}

	log.From(ctx).Debug("recording metric", zap.String("subject", u.ID), zap.String("payload", m.Message.Content))
	stats.Record(ctx, ReportCount.M(int64(1)))

	return nil
}
