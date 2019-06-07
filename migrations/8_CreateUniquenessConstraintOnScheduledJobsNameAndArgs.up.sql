CREATE UNIQUE INDEX unique_jobs_schedule_name_args ON jobs_schedule (name,args) WHERE (enabled is true);
