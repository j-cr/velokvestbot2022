package sqlite

import (
	"log"
	"time"

	"me/velokvestbot/pkg/models"
	. "me/velokvestbot/pkg/prelude"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Repo struct {
	DB *sqlx.DB
}

// returns true if created new user, false if user is already in the db
func (r *Repo) CreateUser(id UserId, name string) (bool, error) {
	{
		exists, err := r.UserExists(id)
		if err != nil {
			return false, err
		}

		if exists {
			return false, nil
		}
	}

	_sql := `
INSERT INTO users ( id,  name,  kind,  currentPoint)
            VALUES(:id, :name, :kind, :currentPoint)
`
	_, err := r.DB.NamedExec(_sql, map[string]any{
		"id":           id,
		"name":         name,
		"kind":         models.POINT_GREEN,
		"currentPoint": "",
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *Repo) UserExists(id UserId) (bool, error) {
	_sql := `
 SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)
`
	var result int

	err := r.DB.Get(&result, _sql, id)
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// returns false on error
func (r *Repo) PointExists(id PointId) bool {
	_sql := `
 SELECT EXISTS(SELECT 1 FROM points WHERE id = ?)
`
	var result int

	err := r.DB.Get(&result, _sql, id)
	if err != nil {
		check("PointExists", err)
		return false
	}

	return result == 1
}

func (r *Repo) SetCurrentPoint(u UserId, p PointId) error {
	_sql := `
UPDATE users SET currentPoint = :point WHERE id = :user
`
	_, err := r.DB.NamedExec(_sql, map[string]any{
		"user":  u,
		"point": p,
	})
	check("SetCurrentPoint", err)

	return err
}

func (r *Repo) ClearCurrentPoint(u UserId) error {
	return r.SetCurrentPoint(u, PointId(""))
	// or: SET currentPoint = NULL, etc
}

func (r *Repo) CurrentPoint(u UserId) (PointId, bool) {
	_sql := `
SELECT currentPoint FROM users WHERE id = ?
`
	var result string

	err := r.DB.Get(&result, _sql, u)
	check("CurrentPoint", err)

	if result == "" {
		return PointId(""), false
	}

	return PointId(result), true
}

func (r *Repo) PointAlreadyVisited(u UserId, p PointId) bool {
	_sql := `
 SELECT EXISTS(SELECT 1 FROM visits WHERE user = ? AND point = ?)
`
	var result int

	err := r.DB.Get(&result, _sql, u, p)
	check("PointAlreadyVisited", err)

	return result == 1
}

func (r *Repo) GetPoint(id PointId) (*models.Point, error) {
	_sql := ` SELECT * FROM points WHERE id = ? `

	result := models.Point{}
	err := r.DB.Get(&result, _sql, id)
	check("GetPoint", err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// whether the user got to this point first or not
func (r *Repo) IsUserFirstHere(u UserId, p PointId) (bool, error) {
	_sql := ` SELECT EXISTS(SELECT 1 FROM visits WHERE user = ? AND point = ?) `

	var alreadyExists int
	err := r.DB.Get(&alreadyExists, _sql, u, p)
	if err != nil {
		check("IsUserFirstHere", err)
		return false, err
	}

	return alreadyExists != 1, nil
}

// returns the score the user got for this point
func (r *Repo) VisitPoint(u UserId, p PointId, ph PhotoId) (int, error) {
	_sql := `
INSERT INTO visits ( user,  point,  photo,  score,  added)
             VALUES(:user, :point, :photo, :score, :added)
`

	point, err := r.GetPoint(p)
	if err != nil {
		return 0, err
	}

	err = r.UpdateUserKind(u, point)
	if err != nil {
		return 0, err
	}

	isFirst, err := r.IsUserFirstHere(u, p)
	if err != nil {
		return 0, err
	}

	score := models.CalculateScoreForPoint(point, isFirst)

	_, err = r.DB.NamedExec(_sql, map[string]any{
		"user":  u,
		"point": p,
		"photo": ph,
		"score": score,
		"added": time.Now().Unix(),
	})
	if err != nil {
		check("VisitPoint", err)
		return 0, err
	}

	return score, nil
}

func (r *Repo) UpdateUserKind(u UserId, p *models.Point) error {
	_sql := ` UPDATE users SET kind = :kind WHERE id = :id AND kind < :kind `

	_, err := r.DB.NamedExec(_sql, map[string]any{
		"id":   u,
		"kind": p.Kind,
	})

	if err != nil {
		check("UpdateUserKind", err)
		return err
	}

	return nil
}

// return top n users by score with the given kind
func (r *Repo) GetTopUsers(n int, kind int) ([]models.User, error) {
	_sql := `
SELECT users.name AS name,
  COALESCE((SELECT SUM(visits.score) FROM visits WHERE visits.user = users.id), 0)
  AS score
FROM users
WHERE users.kind = ?
ORDER BY score DESC
LIMIT ?
`

	var results []models.User
	err := r.DB.Select(&results, _sql, kind, n)
	if err != nil {
		check("GetTopUsers/select", err)
		return nil, err
	}

	return results, nil
}

// XXX: unused
// func (r *Repo) GetUser(id UserId) (*models.User, error) {
// 	_sql := ` SELECT * FROM users WHERE id = ? `

// 	result := models.User{}
// 	err := r.DB.Get(&result, _sql, id)
// 	check("GetUser", err)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &result, nil
// }

// XXX: unused
// func (r *Repo) DeleteVisitByPoint(u UserId, p PointId) error {
// 	_sql := ` DELETE FROM visits WHERE user = :user AND point = :point `

// 	_, err := r.DB.NamedExec(_sql, map[string]any{
// 		"user":  u,
// 		"point": p,
// 	})
// 	if err != nil {
// 		check("DeleteVisitByPoint", err)
// 		return err
// 	}

// 	return nil
// }

func (r *Repo) DeleteVisitByPhoto(ph PhotoId) error {
	_sql := ` DELETE FROM visits WHERE photo = ? `

	_, err := r.DB.Exec(_sql, ph)
	if err != nil {
		check("DeleteVisitByMessage", err)
		return err
	}

	return nil
}

func (r *Repo) PhotoAlreadyUploaded(ph PhotoId) bool {
	_sql := ` SELECT EXISTS(SELECT 1 FROM visits WHERE photo = ?) `
	var result int

	err := r.DB.Get(&result, _sql, ph)
	check("PhotoAlreadyUploaded", err)

	return result == 1
}

// TODO:
// func (r *Repo) UpdatePhoto(u UserId, p PointId, m MessageId) {
// 	_sql := `
// UPDATE visits SET message = :message WHERE user = :user AND point = :point
// `
// 	_, err := r.DB.NamedExec(_sql, map[string]any{
// 		"user":    u,
// 		"point":   p,
// 		"message": m,
// 	})
// 	check("UpdatePhoto", err)
// }

func (r *Repo) GetScore(u UserId) int {
	_sql := ` SELECT COALESCE(SUM(score), 0) FROM visits WHERE user = ? `
	var result int

	err := r.DB.Get(&result, _sql, u)
	check("GetScore", err)

	return result
}

func (r *Repo) AddPoint(p *models.Point) error {
	_sql := `
INSERT INTO points (id, name, kind, url) VALUES (?,?,?,?)
`
	_, err := r.DB.Exec(_sql, p.ID, p.Name, p.Kind, p.Url)
	if err != nil {
		check("AddPoint", err)
		return err
	}

	return nil
}

/// ----------------------------------------------------------------------------

func check(where string, err error) {
	if err != nil {
		log.Println("[DB ERROR]", where, err)
	}
}
