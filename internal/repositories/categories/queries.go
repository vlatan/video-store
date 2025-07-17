package categories

const getCategoriesQuery = `
	SELECT name, slug, updated_at
	FROM category
	WHERE id IN (SELECT DISTINCT category_id FROM post)
`
