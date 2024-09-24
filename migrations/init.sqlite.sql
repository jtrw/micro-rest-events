create table events
(
    id         integer not null
        constraint events_pk
            primary key autoincrement,
    uuid       VARCHAR(64)  default NULL,
    user_id    varchar(50)  default NULL,
    type       varchar(50)  default null,
    status     varchar(50)  default null,
    caption    varchar(155) default null,
    message    text         default null,
    is_seen    tinyint(1)   default 0,
    created_at timestamp    default current_timestamp,
    updated_at timestamp    default current_timestamp
);

create index events_status_index
    on events (status);

create index events_status_user_id_index
    on events (status, user_id);

create index events_user_id_index
    on events (user_id);
