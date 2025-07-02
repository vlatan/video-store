package models

type Category struct {
	Name string `db:"name"`
	Slug string `db:"slug"`
}
