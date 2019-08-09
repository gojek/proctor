CREATE TABLE schedule_context
(
    id          bigint not null primary key,
    schedule_id bigint,
    execution_context_id  bigint unique,
    created_at  timestamp default now(),
    updated_at  timestamp default now(),
    UNIQUE (execution_context_id),
    FOREIGN KEY (execution_context_id) references execution_context (id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (schedule_id) references schedule (id) ON DELETE CASCADE ON UPDATE CASCADE
);


