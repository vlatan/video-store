
-- We are not deleting the newly added columns,
-- nor deleting the old ones,
-- because this migration IS tied with a code change.
-- We want to be able to revert back the code safely if necessary.