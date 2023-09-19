CREATE TABLE public.events (
    id serial PRIMARY KEY,
    uuid uuid NOT NULL,
    user_id varchar(50) NULL,
    "type" varchar(50) NULL,
    status varchar(50) NULL,
    caption varchar(155) NULL,
    message text NULL,
    is_seen bool DEFAULT false,
    created_at timestamp(0) NOT NULL DEFAULT now(),
    updated_at timestamp(0) NULL DEFAULT now()
);

CREATE INDEX events_user_id_idx ON public.events (user_id);
