CREATE TABLE jobs_schedule (
  id uuid not null primary key,
  name text,
  args text,
  tags text,
  notification_emails text,
  time text,
  user_email text,
  created_at timestamp default now(),
  updated_at timestamp default now()
);
