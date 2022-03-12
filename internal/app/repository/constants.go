package repository

const PostgreSQLTable = `create table if not exists urls (
		id           serial not null primary key,
		user_id      text,
        short_url    text,
		original_url text
	);
    create unique index if not exists original_url_ix on urls(original_url);`
