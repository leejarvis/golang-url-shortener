DROP TABLE IF EXISTS urls;
CREATE TABLE urls (
        id serial primary key,
        url text not null,
        created_at timestamp not null default now()
);
