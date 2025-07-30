CREATE TABLE IF NOT EXISTS currencies (
    code VARCHAR(3) PRIMARY KEY
);

INSERT INTO currencies (code)
VALUES ('USD'), ('EUR'), ('GBP'), ('RUB'), ('MXN')
ON CONFLICT (code) DO NOTHING;

CREATE TABLE IF NOT EXISTS rates (
    id SERIAL PRIMARY KEY,
    currency1 VARCHAR(3) NOT NULL,
    currency2 VARCHAR(3) NOT NULL,
    rate DOUBLE PRECISION,
    update_time TIMESTAMP,
    CONSTRAINT fk_rates_currency1 FOREIGN KEY (currency1) REFERENCES currencies(code),
    CONSTRAINT fk_rates_currency2 FOREIGN KEY (currency2) REFERENCES currencies(code),
    CONSTRAINT unique_currency_pair UNIQUE (currency1, currency2)
);

INSERT INTO rates (currency1, currency2)
SELECT c1.code, c2.code
FROM currencies c1
CROSS JOIN currencies c2
WHERE c1.code != c2.code
ON CONFLICT (currency1, currency2) DO NOTHING;

CREATE TYPE reqstatus AS ENUM ('submitted', 'ok', 'failed');
CREATE TABLE IF NOT EXISTS update_requests (
    id SERIAL PRIMARY KEY,
    currency1 VARCHAR(3) NOT NULL,
    currency2 VARCHAR(3) NOT NULL,
    request_status reqstatus NOT NULL,
    CONSTRAINT fk_update_currency1 FOREIGN KEY (currency1) REFERENCES currencies(code),
    CONSTRAINT fk_update_currency2 FOREIGN KEY (currency2) REFERENCES currencies(code)
);
