SELECT
    fuel_type,
    price,
    price_last_updated,
    price_change_effective_timestamp
FROM fuel_prices fp
WHERE fp.node_id = ?
AND fp.fuel_type = ?
ORDER BY fp.price_change_effective_timestamp;