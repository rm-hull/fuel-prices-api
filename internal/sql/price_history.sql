WITH ranked_prices AS (
    SELECT
        fuel_type,
        price,
        price_last_updated,
        price_change_effective_timestamp,
        LAG(price) OVER (ORDER BY price_change_effective_timestamp) AS prev_price
    FROM fuel_prices
    WHERE node_id = ? AND fuel_type = ?
)
SELECT
    fuel_type,
    price,
    price_last_updated,
    price_change_effective_timestamp
FROM ranked_prices
WHERE prev_price IS NULL OR price != prev_price
ORDER BY price_change_effective_timestamp;