package categories

const getCategoriesQuery = `
	SELECT name, slug FROM category
	WHERE id IN (SELECT DISTINCT category_id FROM post)
`
