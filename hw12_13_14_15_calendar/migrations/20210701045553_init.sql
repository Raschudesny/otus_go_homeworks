-- +goose Up
-- +goose StatementBegin
CREATE TABLE events
(
    id          uuid PRIMARY KEY,
    title       text        not null,
    start_time  timestamptz not null,
    end_time    timestamptz not null,
    description text,
    owner_id    text        not null
);

CREATE INDEX event_times_index ON events (start_time, end_time);
CREATE INDEX event_owner_index ON events (owner_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index event_owner_index;
drop index event_times_index;
drop table events;
-- +goose StatementEnd
