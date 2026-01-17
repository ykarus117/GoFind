package Store

import (
	"GoSafe/src/db"
	"database/sql"
	"errors"
)

type Search struct {
	database *db.Database
	cache    map[string][]found
}

type found struct {
	Typ         string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"Description"`
	Id          string `json:"id"`
}

func NewSearch(database *db.Database) *Search {
	return &Search{
		database: database,
		cache:    make(map[string][]found),
	}
}

func (s *Search) Search(user int, text string) ([]found, error) {
	cachedRes, ok := s.cache[text]
	if ok {
		return cachedRes, nil
	}
	res, err := s.database.DB.Query("SELECT id, name, description, type FROM search WHERE search MATCH ? AND user = ? LIMIT 20;", text, string(rune(user)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{
				Message: text,
			}
		}
		return nil, &InternalError{
			Message: "Search Error",
			Err:     err,
		}
	}

	newRes := make([]found, 0)
	for res.Next() {
		var f found
		err = res.Scan(&f.Id, &f.Name, &f.Description, &f.Typ)
		if err != nil {
			return nil, &InternalError{
				Message: "Search Error",
				Err:     err,
			}
		}
		newRes = append(newRes, f)
	}
	s.cache[text] = newRes
	return newRes, nil
}

func (s *Search) Completition(user int, text string) ([]found, error) {
	cachedRes, ok := s.cache[text]
	if ok {
		return cachedRes, nil
	}

	newRes := make([]found, 0)
	res, err := s.database.DB.Query("SELECT id, name, type FROM search WHERE search MATCH ? AND user = ? LIMIT 20;", text+"*", string(rune(user)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{
				Message: text,
			}
		}
		return nil, &InternalError{
			Message: "Search Error",
			Err:     err,
		}
	}

	for res.Next() {
		var f found
		err = res.Scan(&f.Id, &f.Name, &f.Typ)
		if err != nil {
			return nil, &InternalError{
				Message: "Search Error",
				Err:     err,
			}
		}
		newRes = append(newRes, f)
	}
	s.cache[text] = newRes
	return newRes, nil
}
