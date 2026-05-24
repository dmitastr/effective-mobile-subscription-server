CREATE TABLE IF NOT EXISTS subscriptions (
    id serial,
    user_id UUID not null,
    name TEXT not null,
    price INT not null,
    start_date timestamp not null,
    end_date timestamp,
    PRIMARY KEY (user_id, name,start_date )
);