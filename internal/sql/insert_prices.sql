INSERT INTO fuel_prices (
    node_id,
    fuel_type,
    price_last_updated,
    price,
    price_change_effective_timestamp,
    recorded_at
)
VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(node_id, fuel_type, price_last_updated) DO UPDATE SET
    price = EXCLUDED.price,
    price_change_effective_timestamp = EXCLUDED.price_change_effective_timestamp;