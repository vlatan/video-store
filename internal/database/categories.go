package database

type Category struct {
	Name string `db:"name"`
	Slug string `db:"slug"`
}

const getCategoriesQuery = `
SELECT name, slug FROM category
WHERE id IN (SELECT category_id FROM post)
`

// Get a limited number of posts with offset
func (s *service) GetCategories() ([]Category, error) {

	rows, err := s.db.Query(getCategoriesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {

		// Get categories from DB
		var category Category
		if err := rows.Scan(&category.Name, &category.Slug); err != nil {
			return []Category{}, err
		}

		// Include the category in the result
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return []Category{}, err
	}

	return categories, nil
}
