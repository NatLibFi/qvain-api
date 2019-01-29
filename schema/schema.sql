-- SQL schema for Qvain
--
-- Make sure the tables are owned by the user the application connects as,
-- which should not be a role with admin privileges.
--
-- You can see the current active role and session owner with these commands:
--
--   SELECT current_user;
--   SELECT session_user;
--
-- Change ownership in Postgresql (in case you make manual changes as admin):
--
--   REASSIGN OWNED BY idiot TO qvain;
--

SET ROLE qvain;

-- Function `next_object_id` generates a 64-bit ID sequence like Instagram and Twitter Snowflake.
--
-- This ID is used as primary key for user object storage.
--
-- Based on articles by Instagram and Rob Conery:
--   https://instagram-engineering.com/sharding-ids-at-instagram-1cf5a71e5a5c
--   https://rob.conery.io/2014/05/28/a-better-id-generator-for-postgresql/
CREATE OR REPLACE FUNCTION next_object_id(OUT result bigint) AS $$
DECLARE
    our_epoch bigint := 1530540000666;
    seq_id bigint;
    now_millis bigint;
    shard_id int := 1;
BEGIN
    SELECT nextval('object_id_seq') % 1024 INTO seq_id;

    SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000) INTO now_millis;
    result := (now_millis - our_epoch) << 23;
    result := result | (shard_id << 10);
    result := result | (seq_id);
END;
$$ LANGUAGE PLPGSQL;

CREATE SEQUENCE object_id_seq;

-- Table `datasets` contains datasets of different types (families).
--
-- The `blob` field has the actual dataset as it is known to external services;
-- the other fields are internal metadata.
CREATE TABLE datasets (
	id          uuid PRIMARY KEY,
	creator     uuid,
	owner       uuid,

	created     timestamp with time zone DEFAULT now(),
	modified    timestamp with time zone DEFAULT now(),
	synced      timestamp with time zone,
	seq         integer DEFAULT 0,

	published   boolean DEFAULT false,
	valid       boolean DEFAULT false,

	family      int,
	schema      text,
	blob        jsonb
);

-- Table `identities` lists app users and their external identities.
--
-- Performance-wise, t's a toss up between having a JSONB field or joining one-to-many with a normalised table,
-- but it might be easier to query for a dynamically chosen JSONB key when we need to show the identity most relevant to the requested view.
--
-- `uid` is Qvain's user id.
-- `extids` is a JSON object with external service name as key and external account as value.
-- `login` is a boolean that, if true, indicates that the account is an actual user that can log in.
-- `profile` is a JSON object for storing user settings across sessions.
--
-- Note: The JSONB field is indexed, don't stuff too much stuff in it.
--       You can drop unwanted identities trivially with a JSON operator (the WHERE clause speeds up the operation if not all users have that key):
--       UPDATE identities SET extids = extids - 'some_id' WHERE extids->'some_id' IS NOT NULL;
CREATE TABLE identities (
	uid     uuid PRIMARY KEY,
	extids  jsonb DEFAULT '{}'::JSONB,
	login   boolean DEFAULT false,
	profile jsonb DEFAULT '{}'::JSONB
	--CONSTRAINT unique_identity UNIQUE (extid)
);


-- Index `idx_btree_extid_fairdata` indexes the fairdata identity;
-- this sort of index needs to be done for every JSONB field one might want to query against.
CREATE INDEX idx_btree_extid_fairdata ON identities USING BTREE ((extids->>'fairdata'));

-- Index `idx_gin_extid_all` indexes all key/value combinations in `extids`;
-- this index has all key->path->value paths but supports existence checking only.
CREATE INDEX idx_gin_extid_all ON identities USING GIN (extids jsonb_path_ops);

-- Table `lastsync` stores the time of last synchronisation for a user's records from an external service.
CREATE TABLE lastsync (
	uid      uuid PRIMARY KEY REFERENCES identities(uid) ON DELETE CASCADE ON UPDATE CASCADE,
	ts       timestamp with time zone,
	success  boolean DEFAULT false,
	msg      text
);

-- Table `objects` stores user saved objects.
CREATE TABLE objects (
    id       bigint NOT NULL DEFAULT next_object_id(),
    owner    uuid REFERENCES identities(uid) ON DELETE CASCADE ON UPDATE CASCADE,
    family   int,
    schema   text,
    type     text,
    blob     jsonb
);

-- View `view_fairdata_dataset` is the API view of a Fairdata dataset.
-- Note: Sub-queries were faster than joins for test data.
CREATE OR REPLACE VIEW view_fairdata_dataset AS
    SELECT id, row_to_json(ds) "json" FROM (
        SELECT id, created, modified, synced, published, family AS type, schema, blob AS dataset,
        (SELECT extids->'fairdata' FROM identities WHERE uid = creator) AS creator,
        (SELECT extids->'fairdata' FROM identities WHERE uid = owner) AS owner
        FROM datasets
    ) ds, pg_sleep(0.5);

-- View `view_fairdata_list` is the API view of a Fairdata dataset listing.
CREATE OR REPLACE VIEW view_fairdata_list AS
    SELECT json_agg(dslist) "by_owner"
    FROM (
        SELECT id, owner, created, modified, published,
            blob#>'{research_dataset,identifier}' identifier,
            blob#>'{research_dataset,title}' title,
            blob#>'{research_dataset,description}' description,
            blob#>'{preservation_state}' preservation_state
        FROM datasets
        -- WHERE owner = $1
    ) dslist;

-- Function `register_identity` creates a new user on login from an external service.
CREATE OR REPLACE FUNCTION register_identity(_uid UUID, _svc TEXT, _extid TEXT) RETURNS TABLE (
 uid UUID,
 is_new boolean
) AS
$func$
BEGIN
   RETURN QUERY
   SELECT ids.uid, false
   FROM   identities ids
   WHERE  extids @> jsonb_build_object(_svc, _extid)
   LIMIT  1;

   IF NOT FOUND THEN
      RETURN QUERY
      INSERT INTO identities(uid, extids, login)
      VALUES (_uid, jsonb_build_object(_svc, _extid), true)
      RETURNING identities.uid, true;
   END IF;
END
$func$ LANGUAGE plpgsql;

-- Function `owner_exception` returns a boolean value or exception if the wanted owner doesn't match the actual owner.
--
-- Example ($1 is the current application user and $2 is the id of the wanted dataset):
--   SELECT owner_exception($1, owner), blob->'dataset_version_set' from datasets where id = $2;
--
-- NOTE: Not used until the performance impact is more clear.
CREATE OR REPLACE FUNCTION owner_exception(in wanted uuid, in actual uuid)
RETURNS boolean AS
$$
BEGIN
  IF wanted = actual THEN
    RETURN true;
  ELSE
    RAISE EXCEPTION 'not owner';
	RETURN false;
  END IF;
END;
$$
LANGUAGE plpgsql;

-- Predefined users for testing purposes.
--
-- It is safe both to include or exclude this;
-- it just makes sure certain users will have a specific UID.
--
-- On conflict, one can also choose to merge JSON keys:
--   ON CONFLICT (uid) DO UPDATE SET extids = identities.extids || excluded.extids;
INSERT INTO identities(uid, extids) VALUES
    ('053bffbcc41edad4853bea91fc42ea18', '{"fairdata":"2c3683230a580e286c5f5c4b4263f3b80e35f6d1@fairdataid"}'), -- wvh
    ('053d18ecb29e752cb7a35cd77b34f5fd', '{}'), -- epk
    ('05593961536b76fa825281ccaedd4d4f', '{}'), -- am
    ('055ea4dade5ab2145954f56d4b51cef0', '{}'), -- hk
    ('055ea531a6cac569425bed94459266ee', '{}') -- jml
ON CONFLICT DO NOTHING;
