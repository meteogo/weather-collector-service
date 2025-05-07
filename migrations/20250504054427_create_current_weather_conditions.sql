-- +goose Up
-- +goose StatementBegin
CREATE TABLE current_weather_conditions (
    city_name                 VARCHAR(255)        NOT NULL PRIMARY KEY,
    latitude                  DOUBLE PRECISION    NOT NULL,
    longitude                 DOUBLE PRECISION    NOT NULL,
    captured_at               TIMESTAMP           NOT NULL,
    temperature               DOUBLE PRECISION    NOT NULL,
    relative_humidity_percent SMALLINT            NOT NULL,
    wind_speed                DOUBLE PRECISION    NOT NULL,
    weather_code              INTEGER             NOT NULL,
    cloud_cover_percent       SMALLINT            NOT NULL,
    precipitation_millimeters DOUBLE PRECISION    NOT NULL,
    visibility_meters         DOUBLE PRECISION    NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE current_weather_conditions;
-- +goose StatementEnd
