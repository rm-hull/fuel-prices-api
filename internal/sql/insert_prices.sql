INSERT INTO fuel_prices (
    node_id,
    fuel_type,
    price_last_updated,
    price,
    recorded_at
)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(node_id, fuel_type, price_last_updated) DO UPDATE SET
    price = EXCLUDED.price;