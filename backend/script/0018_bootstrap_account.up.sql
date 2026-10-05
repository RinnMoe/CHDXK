-- Reserve id=1 without taking over an existing SSO account or its data.
DO $$
DECLARE fk RECORD; replacement INTEGER;
BEGIN
  IF EXISTS (SELECT 1 FROM users WHERE id=1) THEN
    SELECT GREATEST(COALESCE(MAX(id), 1)+1, nextval('users_id_seq')) INTO replacement FROM users;
    FOR fk IN SELECT conname, conrelid::regclass AS tbl, pg_get_constraintdef(oid) AS definition
      FROM pg_constraint WHERE contype='f' AND confrelid='users'::regclass
    LOOP
      EXECUTE format('ALTER TABLE %s DROP CONSTRAINT %I', fk.tbl, fk.conname);
      EXECUTE format('ALTER TABLE %s ADD CONSTRAINT %I %s', fk.tbl, fk.conname,
        regexp_replace(fk.definition, ' ON UPDATE (NO ACTION|RESTRICT|CASCADE|SET NULL|SET DEFAULT)', '') || ' ON UPDATE CASCADE');
    END LOOP;
    UPDATE users SET id=replacement WHERE id=1;
    UPDATE audit_logs SET actor_user_id=replacement WHERE actor_user_id=1;
  END IF;
  PERFORM setval('users_id_seq', GREATEST(COALESCE((SELECT MAX(id) FROM users),1)+1,2), false);
END $$;
CREATE TABLE bootstrap_accounts (
  user_id INTEGER PRIMARY KEY CHECK (user_id=1) REFERENCES users(id) ON DELETE RESTRICT,
  token_hash TEXT NOT NULL,
  issuer TEXT NOT NULL DEFAULT '',
  subject TEXT NOT NULL DEFAULT '',
  nickname TEXT NOT NULL DEFAULT '',
  has_points_account BOOLEAN NOT NULL DEFAULT false
);
ALTER TABLE point_rewards ADD COLUMN external_subject TEXT NOT NULL DEFAULT '';
UPDATE point_rewards r SET external_subject=i.subject FROM user_identities i WHERE i.user_id=r.user_id;
