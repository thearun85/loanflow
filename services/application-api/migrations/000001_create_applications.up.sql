CREATE TABLE applications (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    applicant_name  TEXT        NOT NULL,
    applicant_email TEXT        NOT NULL,
    amount_minor    BIGINT      NOT NULL CHECK (amount_minor > 0),
    currency        CHAR(3)     NOT NULL,
    term_months     INT         NOT NULL CHECK (term_months BETWEEN 6 AND 84),
    status          TEXT        NOT NULL DEFAULT 'SUBMITTED',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
