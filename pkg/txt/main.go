package txt

import (
	"fmt"
	"me/velokvestbot/pkg/models"
	"strings"
)

// unicode character: colored circle/point
var KindToChar = map[int]string{
	models.POINT_GREEN:  str(0xF0, 0x9F, 0x9F, 0xA2),
	models.POINT_YELLOW: str(0xF0, 0x9F, 0x9F, 0xA1),
	models.POINT_RED:    str(0xF0, 0x9F, 0x94, 0xB4),
}

// for dev
var KindToText = map[int]string{
	models.POINT_GREEN:  "green",
	models.POINT_YELLOW: "yellow",
	models.POINT_RED:    "red",
}

func str(xs ...byte) string { return string(xs) }

/// ----------------------------------------------------------------------------

// if for some reason username and first+last name are empty
const NameForAnonymousUser = "Анонимус"

const BtnHere = "Я здесь!"
const BtnOpenMap = "Открыть карту"

const NoPointSelected = "Точка не выбрана. Сперва выберите точку, нажав кнопку 'Я здесь!'."
const UpdatedPhotoOnAlreadyVisitedPoint = "Фото обновлено."
const NewVisitAwaitingPhoto = "Хорошо, теперь отправьте мне фото с этой точки..."
const AlreadyVisitedAwaitingPhoto = "Точка уже посещена, но вы можете обновить фото..."
const PointAlreadyVisited = "Вы уже посещали эту точку, отправить фотографию заново нельзя."
const PhotoAlreadyUploaded = "Так не пойдет: это фото вы уже загружали. Проверьте, то ли фото вы мне послали, и попробуйте еще раз."

const RaceStarted = "Три... два... один... Погнали!\n Теперь бот принимает отметки о прохождении!"
const RaceStopped = "Всё! Приём отметок о прохождении остановлен."
const RaceIsNotRunning = "Прием отметок о прохождении пока (или уже) закрыт."

const ErrGeneric = "Упс, что-то пошло не так... Попробуйте еще раз чуть позже!"

// testing, admin
const PingMessage = "Я тут."
const ChooseMsgToDelete = "Что удаляем-то?"

func VisitedPoint(newScore, addedScore int) string {
	return fmt.Sprintf("Отлично! Точка засчитана, всего очков: %d (+%d)", newScore, addedScore)
}

func YourScore(score int) string {
	return fmt.Sprintf("Всего очков: %d", score)
}

func StartMessage(user string) string {
	return fmt.Sprintf("Привет, %s! Ну что, погнали?", user)
}

/// ----------------------------------------------------------------------------

func FormatPointName(p *models.Point) string {
	// str := func(xs ...byte) string { return string(xs) }

	// m := map[int]string{
	// 	models.POINT_GREEN:  str(0xF0, 0x9F, 0x9F, 0xA2),
	// 	models.POINT_YELLOW: str(0xF0, 0x9F, 0x9F, 0xA1),
	// 	models.POINT_RED:    str(0xF0, 0x9F, 0x94, 0xB4),
	// }

	// dot, ok := m[p.Kind]

	dot, ok := KindToChar[p.Kind]
	if ok {
		return dot + " " + p.Name
	} else {
		return p.Name
	}
}

/// ----------------------------------------------------------------------------

func Leaderboard(gs, ys, rs []models.User) string {
	t0 := table(KindToChar[models.POINT_GREEN]+" Зелёная дистанция:", gs)
	t1 := table(KindToChar[models.POINT_YELLOW]+" Жёлтая дистанция:", ys)
	t2 := table(KindToChar[models.POINT_RED]+" Красная дистанция:", rs)
	return t0 + t1 + t2
}

func row(place int, name string, score int) string {
	return fmt.Sprintf("%d. %s [%d]\n", place, name, score)
}

func table(title string, users []models.User) string {
	b := strings.Builder{}

	b.WriteString(title + "\n")

	for i, user := range users {
		b.WriteString(row(i+1, user.Name, user.Score))
	}
	b.WriteString("\n")

	return b.String()
}

/// ----------------------------------------------------------------------------
