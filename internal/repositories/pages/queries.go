package pages

const getSinglePageQuery = `
	SELECT slug, title, content
	FROM page 
	WHERE slug = $1
`

const updatePageQuery = `
	UPDATE page
	SET title = $2, content = $3
	WHERE slug = $1
`

const insertPageQuery = `
	INSERT INTO page (slug, title, content)
	VALUES ($1, $2, $3)
`

const deletePageQuery = `
	DELETE FROM page
	WHERE slug = $1
`
