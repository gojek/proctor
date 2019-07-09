CREATE TABLE execution_context
(
    id         bigint,
    job_name   varchar(255),
    user_email varchar(255),
    image_tag  text,
    args       text,
    output     bytea,
    status     varchar(255),
    created_at timestamp default now(),
    updated_at timestamp default now()
);