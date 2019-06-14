CREATE TABLE jobs_execution_audit_log (
  id serial not null primary key,
  job_name text,
  image_name text,
  job_submitted_for_execution text,
  job_args text,
  job_submission_status text,
  created_at timestamp default now()
);
