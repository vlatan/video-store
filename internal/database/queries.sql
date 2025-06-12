-- AND and OR query
WITH search_terms AS (
    SELECT
        to_tsquery('english', $1) AS and_query,
        to_tsquery('english', $2) AS or_query
),
post_with_search_vector AS (
    SELECT
        p.id,
        p.video_id,
        p.title,
        p.thumbnails, (
            setweight(to_tsvector('english', coalesce(p.title, '')), 'A') ||
            setweight(to_tsvector('english', coalesce(p.short_description, '')), 'B')
        ) AS dynamic_search_vector
    FROM post AS p
)
SELECT
    psv.video_id,
    psv.title,
    psv.thumbnails, (
        SELECT COUNT(*)
        FROM post_like
        WHERE post_like.post_id = psv.id
    ) AS likes
FROM
    post_with_search_vector AS psv,
    search_terms AS st
WHERE psv.dynamic_search_vector @@ st.and_query
OR psv.dynamic_search_vector @@ st.or_query
ORDER BY (ts_rank(psv.dynamic_search_vector, st.and_query) * 2) + ts_rank(psv.dynamic_search_vector, st.or_query) DESC
LIMIT $3 OFFSET $4


-- Just the OR query
SELECT 
	video_id, 
	title, 
	thumbnails, (
		SELECT COUNT(*) 
		FROM post_like
		WHERE post_like.post_id = post.id
	) AS likes
FROM post, to_tsquery('english', $1) AS query
WHERE (
	setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
	setweight(to_tsvector('english', coalesce(short_description, '')), 'B')
) @@ query
ORDER BY ts_rank(
	setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
	setweight(to_tsvector('english', coalesce(short_description, '')), 'B'),
	query
) DESC
LIMIT $2 OFFSET $3