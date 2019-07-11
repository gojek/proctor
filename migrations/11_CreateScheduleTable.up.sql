CREATE TABLE schedule
(
    id                  bigint,
    job_name            varchar(255),
    args                text,
    cron                varchar(255),
    notification_emails varchar(255),
    user_email          varchar(255),
    "group"             varchar(255),
    enabled             bool,
    created_at          timestamp default now(),
    updated_at          timestamp default now()
);
