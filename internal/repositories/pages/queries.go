package pages

const getSinglePageQuery = `
	SELECT title, content
	FROM page 
	WHERE slug = $1
`

const updatePageQuery = `
	UPDATE page
	SET title = $2, content = $3, updated_at = NOW()
	WHERE slug = $1
`

const insertPageQuery = `
	INSERT INTO page (slug, title, content)
	VALUES ($1, $2, $3)
`
