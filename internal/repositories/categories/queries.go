package categories

const getCategoriesQuery = `
	SELECT cat.name, cat.slug, cat.updated_at
	FROM category AS cat
	JOIN post ON post.category_id = cat.id
	GROUP BY cat.id
	ORDER BY cat.name
`

const getSitemapCategoriesQuery = `
	SELECT cat.slug, MAX(post.created_at)
	FROM category AS cat
	JOIN post ON post.category_id = cat.id
	GROUP BY cat.slug
`
