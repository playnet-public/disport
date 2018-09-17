package disport

import (
	"context"
	"fmt"
	"time"

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
func Report(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, subject *discordgo.User) error {
	ctx, err := tag.New(ctx,
		tag.Insert(Subject, subject.ID),
	)
	if err != nil {
		log.From(ctx).Error("adding tags", zap.Error(err))
	}

	log.From(ctx).Debug("recording metric", zap.String("subject", subject.ID), zap.String("payload", m.Message.Content))
	stats.Record(ctx, ReportCount.M(int64(1)))

	s.ChannelMessageDelete(m.ChannelID, m.ID)
	if m.Author.ID == subject.ID {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Heads up!\n%s just issued a self report 😳\nDon't let your head down and do better next time, you are not as bad as you think you are 🙂", subject.Mention()))
		return err
	}

	embed, err := s.ChannelMessageSendEmbed(m.ChannelID, ReportEmbed(ctx, s, m.Author, subject))
	if err != nil {
		log.From(ctx).Error("sending embed", zap.Error(err))
		return err
	}

	return AddVoteReactions(ctx, s, embed)
}

// ReportEmbed for user reporting subject
func ReportEmbed(ctx context.Context, s *discordgo.Session, user, subject *discordgo.User) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: fmt.Sprintf("[Disport] @%s#%s reported", subject.Username, subject.Discriminator),
		Author: &discordgo.MessageEmbedAuthor{
			Name:    user.Username,
			IconURL: user.AvatarURL("100x100"),
		},
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Color:       0x587987,
		Description: "Confidence Score and other data will go here.",
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name: "Confidence Score", Value: "0.000 - not implemented", Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name: "Msg Rate", Value: "0 - not implemented", Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name: "Avg Msg Length", Value: "0.000 - not implemented", Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name: "Avg Word Count", Value: "0.000 - not implemented", Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name: "---", Value: "---", Inline: false,
			},
			&discordgo.MessageEmbedField{
				Name: "Support the report", Value: ":white_check_mark:", Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name: "Oppose the report", Value: ":negative_squared_cross_mark:", Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: subject.AvatarURL("100x100"),
			Text:    fmt.Sprintf("=> @%s#%s", subject.Username, subject.Discriminator),
		},
	}
}

// AddVoteReactions to message
func AddVoteReactions(ctx context.Context, s *discordgo.Session, msg *discordgo.Message) error {
	err := s.MessageReactionAdd(msg.ChannelID, msg.ID, "✅")
	if err != nil {
		log.From(ctx).Error("adding reaction", zap.String("embed", msg.ID), zap.Error(err))
		return err
	}

	err = s.MessageReactionAdd(msg.ChannelID, msg.ID, "❎")
	if err != nil {
		log.From(ctx).Error("adding reaction", zap.String("embed", msg.ID), zap.Error(err))
		return err
	}

	return nil
}
