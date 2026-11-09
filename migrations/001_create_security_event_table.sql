create table if not exists security_event (
    id int primary key generated always as identity,
    object_uuid text not null,
    object_type text not null,
    object_confidence real not null,
    created_at timestamptz not null default now()
);

---- create above / drop below ----

drop table if exists security_event cascade;
