-- +goose Up
-- +goose StatementBegin
CREATE TABLE transaction_audit_log (
    transaction_id UUID,
    payload String,  -- здесь хранится валидный JSON
    version Int,
    event_time DateTime DEFAULT now()
) ENGINE = MergeTree()
 ORDER BY (event_time);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE transaction_audit_log;
-- +goose StatementEnd
