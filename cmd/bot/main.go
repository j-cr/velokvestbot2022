package main

import (
	"flag"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"log"
	"math/rand"
	"time"

	tele "gopkg.in/telebot.v3"

	"me/velokvestbot/pkg/models"
	. "me/velokvestbot/pkg/prelude"
	"me/velokvestbot/pkg/repo/sqlite"
)

var TOKEN = "XXX"

// where to post checkpoints:
var CHANNEL_ID = int64(0) // XXX

// where to forward the photos from the users:
var GROUPCHAT_ID = int64(0) // XXX

var MY_ID = int64(0) // XXX

var MY_USERNAMES = []string{
	"GroupAnonymousBot", // admin account for comment chats
}

/// ----------------------------------------------------------------------------

type App struct {
	token string
	// adminId        int64 // note that ids are different in dialogue and in groupchat
	adminUsernames []string
	channelId      int64 // where to post checkpoints
	groupchatId    int64 // where to forward the photos from the users

	raceIsRunning bool // true if we accept new points from the users

	bot  *tele.Bot
	repo interface {
		CreateUser(id UserId, name string) (bool, error)
		GetTopUsers(n int, kind int) ([]models.User, error)
		GetScore(u UserId) int

		AddPoint(point *models.Point) error

		PointExists(p PointId) bool
		SetCurrentPoint(u UserId, p PointId) error
		ClearCurrentPoint(u UserId) error
		CurrentPoint(u UserId) (PointId, bool)
		PhotoAlreadyUploaded(ph PhotoId) bool
		PointAlreadyVisited(u UserId, p PointId) bool
		VisitPoint(u UserId, p PointId, ph PhotoId) (int, error)

		// DeleteVisitByPoint(u UserId, p PointId) error // XXX: del?
		DeleteVisitByPhoto(ph PhotoId) error

		// UpdatePhoto(u UserId, p PointId, m MessageId) error // XXX: del?
	}
}

/// ----------------------------------------------------------------------------

func contains[T string](xs []T, y T) bool {
	for _, x := range xs {
		if x == y {
			return true
		}
	}
	return false
}

func (app *App) middlewareAdminOnly(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if contains(app.adminUsernames, c.Sender().Username) {
			return next(c) // continue execution chain
		} else {
			log.Printf("[WARN] normal user tried admin command: username=[%s], msg=[%s]",
				c.Sender().Username,
				c.Message().Text)
			return nil
		}
	}
}

func (app *App) registerHandlers() {
	b := app.bot

	b.Handle("/ping", app.onPing)             // for testing
	b.Handle("/start", app.onStart)           // private-only
	b.Handle("/score", app.onScore)           // private-only
	b.Handle(tele.OnPhoto, app.onPhoto)       // user uploads a photo; private-only
	b.Handle(tele.OnCallback, app.onCallback) // user clicks the "i'm here" button

	// admin-only commands:
	adminOnly := b.Group()

	adminOnly.Use(app.middlewareAdminOnly)
	{
		adminOnly.Handle("/add", app.onAdd)                     // XXX: temporary
		adminOnly.Handle("/storePoints", app.onStorePoints)     // XXX: dev-only
		adminOnly.Handle("/publishPoints", app.onPublishPoints) // XXX: dev-only
		adminOnly.Handle("/open", app.onOpenRace)               // start or resume the race
		adminOnly.Handle("/close", app.onCloseRace)             // finish or pause the race (don't accept new points)
		adminOnly.Handle("/leaderboard", app.onLeaderboard)
		adminOnly.Handle("/leaderboard10", app.onLeaderboardLong)
		adminOnly.Handle("/del", app.onDel)

	}
}

func main() {
	rand.Seed(time.Now().Unix())

	dsn := flag.String("dsn", "file:sqlite.db?cache=shared",
		"sql connection string (data source name)")
	token := flag.String("token", TOKEN, "telegram bot token")
	// adminId := flag.Int64("admin", MY_ID, "admin user id")
	// adminUsername := flag.String("admin", MY_USERNAME, "admin username")
	channelId := flag.Int64("channel", CHANNEL_ID, "where to post checkpoints")
	groupchatId := flag.Int64("groupchat", GROUPCHAT_ID, "where to forward the photos from the users")

	flag.Parse() // exits on error

	log.Printf("connecting to db: %s", *dsn)
	db, err := openDB(*dsn)
	check(err)
	defer db.Close()

	app := &App{
		token: *token,
		// adminId:     *adminId,
		adminUsernames: MY_USERNAMES,
		channelId:      *channelId,
		groupchatId:    *groupchatId,

		raceIsRunning: true, // true by default in case the bot crashes and is restarted

		bot:  newBot(*token),
		repo: &sqlite.Repo{DB: db},
	}

	log.Println("=== starting")
	app.registerHandlers()
	app.bot.Start()
}

func newBot(token string) *tele.Bot {
	// token, ok := os.LookupEnv("TG_TOKEN")
	// if !ok { token = TOKEN }

	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	// b.Use(middleware.Logger()) // XXX

	return b
}

func openDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", dsn)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

/// ----------------------------------------------------------------------------

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
