DROP VIEW IF EXISTS daily_fuel_price_stats;
CREATE VIEW daily_fuel_price_stats AS
WITH latest_prices_today AS (
    -- Filter for prices updated today and pick the latest for each (node_id, fuel_type)
    SELECT
        fp.node_id,
        fp.fuel_type,
        fp.price,
        pfs.postcode,
        ROW_NUMBER() OVER (PARTITION BY fp.node_id, fp.fuel_type ORDER BY fp.price_last_updated DESC) as rn
    FROM fuel_prices fp
    JOIN petrol_filling_stations pfs ON fp.node_id = pfs.node_id
    WHERE DATE(fp.price_last_updated) = DATE('now')
),
latest_today AS (
    SELECT * FROM latest_prices_today WHERE rn = 1
),
with_postcode_area AS (
    -- Extract the postcode area (the leading alphabetic characters)
    SELECT
        fuel_type,
        price,
        UPPER(SUBSTR(TRIM(postcode), 1, LENGTH(TRIM(postcode)) - LENGTH(LTRIM(TRIM(postcode), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ')))) as postcode_area
    FROM latest_today
),
national_stats AS (
    SELECT
        'National' as scope,
        NULL as postcode_area,
        fuel_type,
        MIN(price) as min_price,
        ROUND(AVG(price),1) as avg_price,
        MAX(price) as max_price,
        SQRT(MAX(0, AVG(price * price) - AVG(price) * AVG(price))) as stddev_price,
        COUNT(*) as sample_size
    FROM with_postcode_area
    GROUP BY fuel_type
),
postcode_area_stats AS (
    SELECT
        'Postcode Area' as scope,
        postcode_area,
        fuel_type,
        MIN(price) as min_price,
        ROUND(AVG(price),1) as avg_price,
        MAX(price) as max_price,
        SQRT(MAX(0, AVG(price * price) - AVG(price) * AVG(price))) as stddev_price,
        COUNT(*) as sample_size
    FROM with_postcode_area
    GROUP BY postcode_area, fuel_type
)
SELECT * FROM national_stats
UNION ALL
SELECT * FROM postcode_area_stats;
