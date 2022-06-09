package models

import (
	. "me/velokvestbot/pkg/prelude"
)

const POINT_GREEN = 0
const POINT_YELLOW = 1
const POINT_RED = 2

type Point struct {
	ID   PointId
	Name string
	Kind int    // see constants above
	Url  string // link to the map
	// Pos  string // gps coordinates as a string (ll)
	// Description string
	// Picture     string // url?

}

// A single user that started our bot
type User struct {
	ID           UserId
	Name         string // first name + last name for example
	Kind         int    // same as point.Kind, zero default is ok
	CurrentPoint PointId

	Score int // !!!: isn't in the database, we compute it on the fly
}

// Visit represents an event of a user sending a message with a photo to prove
// he passed a checkpoint.
//
// Note that there may be multiple Visits for a single user\checkpoint pair,
// e.g. if a user have sent a wrong photo and sends a photo one more time; this
// should be accounted for when counting the total score.
type Visit struct {
	ID    int     // autoincrement
	User  UserId  // tg user id
	Point PointId // which point
	Photo string  // file_unique_id of the photo sent with the message
	Score int     // how many points we gave the user for this point
	Added int64   // only need to compare which is less, so use unix time
}

func CalculateScoreForPoint(p *Point, isUserFirst bool) (score int) {
	switch p.Kind {
	case POINT_GREEN:
		score = 1
	case POINT_YELLOW:
		score = 2
	case POINT_RED:
		score = 10
	default:
		return 0
	}

	if isUserFirst {
		score += 1
	}

	return
}

// when user visits point p we may update user's kind to match the point
func UpdateUserKind(u *User, p *Point) {
	if p.Kind > u.Kind {
		u.Kind = p.Kind
	}
}
