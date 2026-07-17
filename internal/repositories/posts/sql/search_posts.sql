WITH
    -- Construct AND, OR and RAW queries from the input search phrase
    search_terms AS (
        SELECT
            lexeme AS and_query,
            to_tsquery('english', replace(lexeme::text, ' & ', ' | ')) AS or_query,
            replace(lexeme::text, ' & ', ' ') AS raw_query
        FROM plainto_tsquery('english', $1) AS lexeme
    ),
    -- Isolated GIN scan #1 - match posts
    post_matches AS (
        SELECT 
            p.id,
            (ts_rank(p.search_vector, st.and_query, 32) * 2) + 
            ts_rank(p.search_vector, st.or_query, 32) +
            (similarity(p.title, st.raw_query) * 0.25) + 
            (COALESCE(similarity(p.original_title, st.raw_query), 0) * 0.25) AS post_score
        FROM post AS p
        CROSS JOIN search_terms AS st
        -- If st.or_query is matched no need to look for st.and_query match
        WHERE p.search_vector @@ st.or_query
    ),
    -- Isolated GIN scan #2 - match reviews and calculate scores
    review_matches AS (
        SELECT
            pr.post_id,
            (MAX(ts_rank(pr.search_vector, st.and_query, 32)) * 1.5) +
            (MAX(ts_rank(pr.search_vector, st.or_query, 32)) * 0.75) +
            (MAX(COALESCE(similarity(pr.title, st.raw_query), 0) * 0.25)) AS review_score
        FROM post_review AS pr
        CROSS JOIN search_terms AS st
        WHERE pr.search_vector @@ st.or_query
        GROUP BY pr.post_id
    ),
    -- Merge the IDs and save the total score
    combined_matches AS (
        SELECT 
            COALESCE(pm.id, rm.post_id) AS post_id,
            COALESCE(pm.post_score, 0) + COALESCE(rm.review_score, 0) AS total_score
        FROM post_matches AS pm
        FULL OUTER JOIN review_matches AS rm ON pm.id = rm.post_id
    ),
    likes AS (
        SELECT post_id, COUNT(*) AS likes
        FROM post_like
        GROUP BY post_id
    ),
    ratings AS (
        SELECT
            post_id,
            ROUND(AVG(rating), 2)::float8 AS avg_rating,
            COUNT(rating) AS rating_count
        FROM post_rating
        GROUP BY post_id
    ),
    -- Get the data we need
    scored_posts AS (
        SELECT
            p.id,
            p.video_id,
            p.title,
            p.original_title,
            p.thumbnails,
            COALESCE(l.likes, 0) AS likes,
            r.avg_rating,
            COALESCE(r.rating_count, 0) AS rating_count,
            {{ .TotalCount }} AS total_results,
            p.upload_date,
            cm.total_score AS score
        FROM combined_matches AS cm
        JOIN post AS p ON p.id = cm.post_id 
        LEFT JOIN likes AS l ON l.post_id = p.id
        LEFT JOIN ratings AS r ON r.post_id = p.id
    )
    --- Filter posts
	SELECT * FROM scored_posts
    {{ .WhereCondition }} -- the WHERE condition if any
    ORDER BY score DESC, upload_date DESC, id DESC
    LIMIT $2;
