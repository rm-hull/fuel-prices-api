SELECT
    node_id,
    mft_organisation_name,
    public_phone_number,
    trading_name,
    is_same_trading_and_brand_name,
    brand_name,
    temporary_closure,
    permanent_closure,
    permanent_closure_date,
    is_motorway_service_station,
    is_supermarket_service_station,
    address_line_1,
    address_line_2,
    city,
    country,
    county,
    postcode,
    latitude,
    longitude,
    opening_times_json,
    amenities_json,
    fuel_types_json
FROM petrol_filling_stations
WHERE latitude BETWEEN ? AND ?
  AND longitude BETWEEN ? AND ?;