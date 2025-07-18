package categories

const getCategoriesQuery = `
	SELECT name, slug, updated_at
	FROM category
	WHERE id IN (SELECT DISTINCT category_id FROM post)
`

const getSitemapCategoriesQuery = `
	SELECT cat.slug, MAX(post.created_at)
	FROM category AS cat
	JOIN post ON post.category_id = cat.id
	GROUP BY cat.slug
`
