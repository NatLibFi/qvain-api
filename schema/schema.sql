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

CREATE TABLE datasets (
	id          uuid PRIMARY KEY,
	creator     uuid,
	owner       uuid,
	
	created     timestamp with time zone DEFAULT now(),
	modified    timestamp with time zone DEFAULT now(),
	
	pushed      timestamp with time zone,
	pulled      timestamp with time zone,

	published   boolean DEFAULT false,
	valid       boolean DEFAULT false,
	
	family      int,
	schema      text,
	blob        jsonb
);

CREATE TABLE identities (
	uid     uuid PRIMARY KEY,
	extid   text NOT NULL
);

CREATE TABLE lastsync (
	uid      uuid PRIMARY KEY REFERENCES identities(uid),
	ts       timestamp with time zone,
	success  boolean DEFAULT false,
	msg      text
);
