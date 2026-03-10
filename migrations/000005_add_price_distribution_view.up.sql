DROP VIEW IF EXISTS fuel_price_distribution_stats;
CREATE VIEW fuel_price_distribution_stats AS
WITH latest_prices_raw AS (
    -- Pick the latest record for each (node_id, fuel_type) pair
    SELECT
        fp.node_id,
        fp.fuel_type,
        fp.price,
        pfs.postcode,
        ROW_NUMBER() OVER (PARTITION BY fp.node_id, fp.fuel_type ORDER BY fp.price_last_updated DESC) as rn
    FROM fuel_prices fp
    JOIN petrol_filling_stations pfs ON fp.node_id = pfs.node_id
),
latest_snapshot AS (
    SELECT * FROM latest_prices_raw WHERE rn = 1
),
with_postcode_area AS (
    -- Extract the postcode area (the leading alphabetic characters)
    SELECT
        fuel_type,
        price,
        UPPER(SUBSTR(TRIM(postcode), 1, LENGTH(TRIM(postcode)) - LENGTH(LTRIM(TRIM(postcode), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ')))) as postcode_area
    FROM latest_snapshot
),
national_distribution AS (
    SELECT
        'National' as scope,
        NULL as postcode_area,
        fuel_type,
        (CAST(price / 2 AS INT) * 2) as price_bucket,
        COUNT(*) as sample_size
    FROM with_postcode_area
    GROUP BY fuel_type, price_bucket
),
postcode_area_distribution AS (
    SELECT
        'Postcode Area' as scope,
        postcode_area,
        fuel_type,
        (CAST(price / 2 AS INT) * 2) as price_bucket,
        COUNT(*) as sample_size
    FROM with_postcode_area
    WHERE postcode_area IS NOT NULL AND postcode_area <> ''
    GROUP BY postcode_area, fuel_type, price_bucket
)
SELECT * FROM national_distribution
UNION ALL
SELECT * FROM postcode_area_distribution;
