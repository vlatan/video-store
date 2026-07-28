
WITH ratings_stats AS (
    SELECT
        post_id,
        ROUND(AVG(rating), 2)::float8 AS avg_rating,
        COUNT(rating) AS rating_count
    FROM post_rating
    GROUP BY post_id
)
SELECT
    rs.avg_rating,
    rs.rating_count,
    prev.title AS headline,
    prev.review AS content
FROM post_rating AS prat
JOIN post_review AS prev ON prev.rating_id = prat.id
LEFT JOIN ratings_stats AS rs ON rs.post_id = prat.post_id
WHERE prat.post_id = $1;
