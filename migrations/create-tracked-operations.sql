CREATE SCHEMA IF NOT EXISTS ana;

CREATE TYPE ana.operation_status AS ENUM(
  'ready', 'running', 'finished', 'failed'
);

CREATE TABLE IF NOT EXISTS ana.tracked_operations (
  reference_time timestamptz                  NOT NULL,
  started_at     timestamptz                  NOT NULL,
  finished_at    timestamptz,
  timeout        timestamptz,
  expiration     timestamptz,
  error_count    integer                      NOT NULL DEFAULT 0,
  status         ana.operation_status NOT NULL DEFAULT 'running',
  target         varchar                      NOT NULL,
  key            varchar                      NOT NULL,
  payload        bytea                        NOT NULL,
  result         bytea,
  error_message  varchar,

  PRIMARY KEY(target, key, reference_time)
);

CREATE OR REPLACE FUNCTION ana.fetch_or_start(
  _key            varchar,
  _target         varchar,
  _payload        bytea,
  _reference_time timestamptz,
  _timeout        interval,
  _expiration     interval
) RETURNS ana.tracked_operations
LANGUAGE plpgsql
AS $$
DECLARE
  _operation ana.tracked_operations;
BEGIN
  SELECT * INTO _operation
  FROM ana.tracked_operations
  WHERE key = _key AND target = _target
  FOR UPDATE;

  IF FOUND THEN
    RETURN _operation;
  END IF;

  INSERT INTO ana.tracked_operations (
    status,
    key,
    target,
    payload,
    reference_time,
    timeout,
    expiration,
    started_at
  ) VALUES (
    'running',
    _key,
    _target,
    _payload,
    _reference_time,
    NOW()           + NULLIF(_timeout,    '0'::interval),
    _reference_time + NULLIF(_expiration, '0'::interval),
    NOW()
  ) RETURNING * INTO _operation;

  _operation.status = 'ready';
  RETURN _operation;
END;
$$;
