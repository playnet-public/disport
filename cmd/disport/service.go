package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/playnet-public/disport/pkg/disport"
	"github.com/playnet-public/disport/pkg/service"

	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi"
	"github.com/seibert-media/golibs/log"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

const (
	appName = "Disport"
	appKey  = "disport"
)

// Spec for the service
type Spec struct {
	service.BaseSpec

	Addr  string `envconfig:"metrics" required:"true" help:"metrics port"`
	Token string `envconfig:"discord_token" required:"true" help:"discord bot token"`
	Guild string `envconfig:"discord_guild" required:"true" help:"discord guild id"`
}

var (
	// Guild ID of the recorded metric
	Guild, _ = tag.NewKey("guild")
	// Channel ID of the recorded metric
	Channel, _ = tag.NewKey("channel")
	// User ID of the recorded metric
	User, _ = tag.NewKey("user")
)

var (
	// ReportCountView .
	ReportCountView = &view.View{
		Name:        "report/count",
		Measure:     disport.ReportCount,
		Description: "The number of reports issued",
		TagKeys:     []tag.Key{Channel, User, disport.Subject},
		Aggregation: view.Count(),
	}
)

func main() {
	var svc Spec
	ctx := service.Init(appKey, appName, &svc)
	defer service.Defer(ctx)

	log.From(ctx).Info("creating prometheus exporter")
	exporter, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.From(ctx).Fatal("creating prometheus exporter", zap.Error(err))
	}
	view.RegisterExporter(exporter)

	log.From(ctx).Info("registering views")
	if err := view.Register(ReportCountView); err != nil {
		log.From(ctx).Fatal("registering views", zap.Error(err))
	}
	view.SetReportingPeriod(1 * time.Second)

	log.From(ctx).Info("creating discord client")
	discord, err := discordgo.New("Bot " + svc.Token)
	if err != nil {
		log.From(ctx).Fatal("creating discord client", zap.Error(err))
	}

	ctx = log.WithFields(ctx, zap.String("guild", svc.Guild))

	discord.AddHandler(handler(ctx, svc.Guild))

	if err := discord.Open(); err != nil {
		log.From(ctx).Fatal("opening discord connection")
	}
	defer discord.Close()

	router := chi.NewRouter()
	router.Get("/metrics", exporter.ServeHTTP)
	var srv = http.Server{
		Addr:    svc.Addr,
		Handler: router,
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		log.From(ctx).Info("shutting down server")
		err = srv.Shutdown(ctx)
		if err != nil {
			log.From(ctx).Fatal("shutting down server", zap.Error(err))
		}
	}()

	log.From(ctx).Info("serving metrics", zap.String("addr", svc.Addr))
	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.From(ctx).Fatal("serving metrics", zap.String("addr", svc.Addr), zap.Error(err))
	}

	log.From(ctx).Info("finished")
}

// handler will get called on every message and is responsible for updating the respective metrics
func handler(ctx context.Context, guild string) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		ctx := log.WithFields(ctx,
			zap.String("author", m.Author.ID),
			zap.String("channel", m.ChannelID),
			zap.String("message", m.ID),
		)

		if m.Author.ID == s.State.User.ID {
			return
		}

		if len(m.Mentions) < 2 {
			return
		}

		ctx, err := tag.New(ctx,
			tag.Insert(Guild, guild),
			tag.Insert(Channel, m.ChannelID),
			tag.Insert(User, m.Author.ID),
		)
		if err != nil {
			log.From(ctx).Error("adding tags", zap.Error(err))
		}

		cancel := make(chan struct{})
		for _, mention := range m.Mentions {
			go HandleMention(ctx, s, m, mention, cancel)
		}
	}
}

// HandleMention by checking for a mention of the bot itself
// If a mention is detected, the cancel channel is being closed and all other HandleMention instances may continue
func HandleMention(ctx context.Context,
	s *discordgo.Session, m *discordgo.MessageCreate, u *discordgo.User,
	cancel chan struct{}) {

	if u.ID == s.State.User.ID {
		log.From(ctx).Debug("sending signal")
		close(cancel)
		return
	}

	log.From(ctx).Debug("waiting for signal")
	select {
	case _, cancel := <-cancel:
		if cancel {
			log.From(ctx).Debug("canceling", zap.String("reason", "signal"))
			return
		}
	case <-time.After(5 * time.Second):
		log.From(ctx).Debug("skipping message", zap.String("reason", "timeout"))
		return
	}

	disport.Report(ctx, s, m, u)
}
