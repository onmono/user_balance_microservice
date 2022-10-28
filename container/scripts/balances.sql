CREATE TABLE IF NOT EXISTS public.user_balance
(
    id              uuid                        NOT NULL UNIQUE,
    user_id         uuid                        NOT NULL UNIQUE,
    balance         bigint CHECK (balance >= 0) NOT NULL,
    last_updated_at timestamp                   NOT NULL
);


-- SELECT * from user_balance WHERE user_id = '23588874-41a2-4741-80e9-9579ddeffd3e';

ALTER TABLE ONLY public.user_balance
    ADD CONSTRAINT user_balance_pkey PRIMARY KEY (id);

CREATE TABLE IF NOT EXISTS public.reserve_info
(
    id         uuid                     NOT NULL UNIQUE,
    reserve_id uuid                     NOT NULL,
    user_id    uuid                     NOT NULL,
    service_id uuid                     NOT NULL,
    order_id   uuid                     NOT NULL,
    price      bigint CHECK (price > 0) NOT NULL,
    timestamp  timestamp                NOT NULL,
    CONSTRAINT fk_reserve_user_id
        FOREIGN KEY (user_id)
            REFERENCES public.user_balance (user_id),
    CONSTRAINT fk_reserve_reserve_id
        FOREIGN KEY (reserve_id)
            REFERENCES public.user_balance (user_id)
);

ALTER TABLE ONLY public.reserve_info
    ADD CONSTRAINT reserve_pkey PRIMARY KEY (id);

-- индексы для slave репликации, чтение сразу по индексу в область
CREATE UNIQUE INDEX user_id_user_balance_index
    ON public.user_balance (user_id);

CREATE INDEX user_reserve_index
    ON public.reserve_info (reserve_id);

CREATE TABLE IF NOT EXISTS public.accounting_revenue
(
    id         uuid                     NOT NULL UNIQUE,
    user_id    uuid                     NOT NULL,
    service_id uuid                     NOT NULL,
    order_id   uuid                     NOT NULL,
    sum      bigint CHECK (accounting_revenue.sum > 0) NOT NULL,
    timestamp  timestamp                NOT NULL
);

ALTER TABLE ONLY public.accounting_revenue
    ADD CONSTRAINT accounting_revenue_pkey PRIMARY KEY (id);

-- индексы для slave репликации, чтение сразу по индексу в область
CREATE INDEX user_id_accounting_revenue_index
    ON public.accounting_revenue (user_id);