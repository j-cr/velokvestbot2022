package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	tele "gopkg.in/telebot.v3"

	"me/velokvestbot/pkg/data"
	"me/velokvestbot/pkg/models"
	"me/velokvestbot/pkg/txt"
)

func (app *App) onStart(c tele.Context) error {
	// ignore the command if it's not in a private dialogue with our bot
	if !c.Message().Private() {
		return nil
	}

	user := c.Sender()
	id := user.ID
	name := guessName(user)

	log.Printf("user invoked /start: id=[%d], name=[%s]", id, name)

	done, err := app.repo.CreateUser(id, name)
	if err != nil {
		log.Println("[DB ERROR] CreateUser", id, name, err)
		return c.Send(txt.ErrGeneric)
	}
	if done {
		log.Printf("new user in db: id=[%d], name=[%s]", id, name)
	}

	return c.Send(txt.StartMessage(name))
}

func guessName(u *tele.User) string {
	switch {
	case u.LastName != "":
		return u.FirstName + " " + u.LastName
	case u.FirstName != "":
		return u.FirstName
	case u.Username != "":
		return u.Username
	default:
		return txt.NameForAnonymousUser
	}
}

func (app *App) onPing(c tele.Context) error {
	return c.Send(txt.PingMessage)
}

func (app *App) onScore(c tele.Context) error {
	// ignore the command if it's not in a private dialogue with our bot
	if !c.Message().Private() {
		return nil
	}

	user := c.Sender().ID
	score := app.repo.GetScore(user)
	return c.Reply(txt.YourScore(score))

}

/// ----------------------------------------------------------------------------

func (app *App) onCallback(c tele.Context) error {
	defer c.Respond() // respond to tg api to stop the spinner on the button

	r := app.repo
	user := c.Sender()
	point := c.Args()[1] // 0 is button's Unique, 1 is button's data

	if !app.raceIsRunning {
		return app.SendTo(user, txt.RaceIsNotRunning)
	}

	if !r.PointExists(point) {
		log.Printf("[ERROR] point doesn't exist: id=[%s]", point)
		return app.SendTo(user, txt.ErrGeneric)
	}

	if r.PointAlreadyVisited(user.ID, point) {
		r.ClearCurrentPoint(user.ID)
		return app.SendTo(user, txt.PointAlreadyVisited)
	}

	err := r.SetCurrentPoint(user.ID, point)
	if err != nil {
		log.Printf("[ERROR] can't set current point: user=[%d], point=[%s]", user.ID, point)
		return app.SendTo(user, txt.ErrGeneric)
	}

	log.Printf("current point set: user=[%d], point=[%s]", user.ID, point)

	return app.SendTo(user, txt.NewVisitAwaitingPhoto)
}

func (app *App) onPhoto(c tele.Context) error {
	// ignore the command if it's not in a private dialogue with our bot
	if !c.Message().Private() {
		return nil
	}

	if !app.raceIsRunning {
		return c.Reply(txt.RaceIsNotRunning)
	}

	r := app.repo
	user := c.Sender().ID
	photo := c.Message().Photo.UniqueID // shouldn't be nil since we're in onPhoto handler
	log.Printf("onPhoto: user=[%d], photo=[%s]", user, photo)

	point, isPointSelected := r.CurrentPoint(user)

	if !isPointSelected {
		return c.Reply(txt.NoPointSelected)
	}

	if r.PointAlreadyVisited(user, point) {
		r.ClearCurrentPoint(user)
		return c.Reply(txt.PointAlreadyVisited)
	}

	if r.PhotoAlreadyUploaded(photo) {
		return c.Reply(txt.PhotoAlreadyUploaded)
	}

	score, err := r.VisitPoint(user, point, photo)
	r.ClearCurrentPoint(user) // so users won't accidentally send a wrong photo later
	if err != nil {
		return c.Reply(txt.ErrGeneric)
	}

	log.Printf("visited point: user=[%d] point=[%s] score=[%d]", user, point, score)

	{ // reply to the user privately
		msg := txt.VisitedPoint(r.GetScore(user), score)
		err := c.Reply(msg)
		if err != nil {
			log.Println("[TG ERROR] onPhoto/reply-private", err)
		}
	}

	go func() { // forward the uploaded photo to the groupchat
		time.Sleep(3 * time.Second)
		target := &tele.Chat{ID: app.groupchatId}
		err := c.ForwardTo(target)
		if err != nil {
			log.Println("[TG ERROR] onPhoto/forward-to-groupchat", err)
		}
	}()

	return nil
}

/// ----------------------------------------------------------------------------
/// admin ----------------------------------------------------------------------

func (app *App) newPointPostMarkup(pointId string, url string) *tele.ReplyMarkup {
	buttons := app.bot.NewMarkup()
	btnHere := buttons.Data(txt.BtnHere, "btnHere"+pointId, pointId)
	// XXX: removed for now
	// btnOpenMap := buttons.URL(txt.BtnOpenMap, url)
	// buttons.Inline(buttons.Row(btnHere, btnOpenMap))
	buttons.Inline(buttons.Row(btnHere))
	return buttons
}

func (app *App) publishPoint(point *models.Point) error {
	target := &tele.Chat{ID: app.channelId}
	text := txt.FormatPointName(point)
	buttons := app.newPointPostMarkup(point.ID, point.Url)

	_, err := app.bot.Send(target, text, buttons)
	if err != nil {
		log.Println("[TG ERROR] can't publish point:", point.ID, err)
		return err
	}

	log.Printf("[ADMIN] publishPoint: %#v", point)
	return nil
}

// XXX: for tests only
func newMockedPoint() *models.Point {
	id := fmt.Sprint(rand.Intn(100))
	// types := []int{models.POINT_GREEN, models.POINT_YELLOW, models.POINT_RED, models.POINT_OTHER}
	types := []int{models.POINT_GREEN, models.POINT_YELLOW, models.POINT_RED}

	return &models.Point{
		ID:   id,
		Name: "point:" + id,
		Kind: types[rand.Intn(len(types))],
		Url:  "https://google.com?q=" + id,
	}
}

// XXX: for tests
func (app *App) onAdd(c tele.Context) error {
	point := newMockedPoint()

	err := app.repo.AddPoint(point)
	if err != nil {
		return c.Reply(txt.ErrGeneric)
	}

	return app.publishPoint(point)
}

func preparePoints(kind int, titles []string) []*models.Point {
	var points []*models.Point

	for i, t := range titles {
		points = append(points, &models.Point{
			ID:   txt.KindToText[kind] + fmt.Sprint(i),
			Name: t,
			Kind: kind,
			Url:  "", // XXX: unused
		})
	}

	return points
}

func (app *App) storePoints(points []*models.Point) {
	for _, p := range points {
		app.repo.AddPoint(p)
	}
}

func (app *App) onStorePoints(c tele.Context) error {
	app.storePoints(preparePoints(models.POINT_GREEN, data.PointsGreen))
	app.storePoints(preparePoints(models.POINT_YELLOW, data.PointsYellow))
	app.storePoints(preparePoints(models.POINT_RED, data.PointsRed))

	return c.Reply("stored")
}

func (app *App) publishPoints(points []*models.Point) {
	for _, p := range points {
		app.publishPoint(p)
		time.Sleep(3 * time.Second)
	}
}

func (app *App) onPublishPoints(c tele.Context) error {
	app.publishPoints(preparePoints(models.POINT_GREEN, data.PointsGreen))
	app.publishPoints(preparePoints(models.POINT_YELLOW, data.PointsYellow))
	app.publishPoints(preparePoints(models.POINT_RED, data.PointsRed))

	return c.Reply("done")
}

func (app *App) onOpenRace(c tele.Context) error {
	target := &tele.Chat{ID: app.groupchatId}
	_, err := app.bot.Send(target, txt.RaceStarted)
	log.Println("started the race")
	app.raceIsRunning = true
	return err
}
func (app *App) onCloseRace(c tele.Context) error {
	target := &tele.Chat{ID: app.groupchatId}
	_, err := app.bot.Send(target, txt.RaceStopped)
	log.Println("stopped the race")
	app.raceIsRunning = false
	return err
}

func sendLeaderboard(n int, app *App, c tele.Context) error {
	u0, err0 := app.repo.GetTopUsers(n, models.POINT_GREEN)
	u1, err1 := app.repo.GetTopUsers(n, models.POINT_YELLOW)
	u2, err2 := app.repo.GetTopUsers(n, models.POINT_RED)

	if err0 != nil || err1 != nil || err2 != nil {
		return c.Reply(txt.ErrGeneric)
	}

	text := txt.Leaderboard(u0, u1, u2)

	target := &tele.Chat{ID: app.groupchatId}
	return app.SendTo(target, text)
}

// top-3 in each group
func (app *App) onLeaderboard(c tele.Context) error {
	return sendLeaderboard(3, app, c)
}

// top 10 in each group
func (app *App) onLeaderboardLong(c tele.Context) error {
	return sendLeaderboard(10, app, c)
}

func (app *App) onDel(c tele.Context) error {
	if !c.Message().IsReply() {
		return c.Reply(txt.ChooseMsgToDelete)
	}

	msg := c.Message().ReplyTo
	if msg == nil {
		return errors.New("[WARN] onDel: ReplyTo field is nil")
	}
	if msg.Photo == nil {
		return errors.New("[WARN] onDel: photo is nil")
	}

	err := app.repo.DeleteVisitByPhoto(msg.Photo.UniqueID)
	if err != nil {
		return c.Reply(txt.ErrGeneric)
	}

	log.Printf("[INFO] onDel: photo=[%s]", msg.Photo.UniqueID)

	return app.bot.Delete(msg)
}

/// utils ----------------------------------------------------------------------

func (app *App) SendTo(user tele.Recipient, msg string) error {
	_, err := app.bot.Send(user, msg)
	return err
}
