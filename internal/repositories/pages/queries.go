package pages

const getSinglePageQuery = `
	SELECT title, content
	FROM page 
	WHERE slug = $1
`
