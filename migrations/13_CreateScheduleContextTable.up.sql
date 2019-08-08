CREATE TABLE schedule_context
(
    id          bigint not null primary key,
    schedule_id bigint,
    context_id  bigint unique,
    created_at  timestamp default now(),
    updated_at  timestamp default now(),
    UNIQUE (context_id),
    FOREIGN KEY (context_id) references execution_context (id),
    FOREIGN KEY (schedule_id) references schedule (id)
);


